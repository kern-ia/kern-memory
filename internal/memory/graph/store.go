// Package graph is the memory contract's relationship layer (memory.KindGraph): directed
// edges between existing .okf/vector memories. It stores references only — it has no
// cross-layer read access, so it never validates that FromID/ToID actually exist
// (docs/epics/epic-1-graph-layer/EPIC_1.md Notes). See
// docs/planning/specs/03-graph-storage-schema.md for the schema decision.
package graph

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yoann/kern-memory/internal/memory"
)

// busyTimeoutMS mirrors internal/memory/okf/store.go — same lesson, same fix.
const busyTimeoutMS = 5000

// maxRelationLength caps Relation: it's a label, not prose
// (docs/planning/specs/11-security-privacy.md — validation replaces pseudonymization here).
const maxRelationLength = 64

// maxDepth hard-caps how many hops a multi-hop Query walks, regardless of what the caller
// requests via Query.Depth (docs/planning/specs/10-traversal-depth-limit.md) — the same
// clamp-not-trust precedent as vector.Store.Query's Limit clamp. An uncapped recursive walk
// over a graph that may contain cycles has no natural stopping point; capping the hop count
// is also what makes the recursive query below terminate on a cyclic graph, not just fast
// on an acyclic one.
const maxDepth = 3

// recursiveTraversalQuery walks graph_edges outward from a starting (from_kind, from_id)
// node, breadth-first by hop. Termination does not depend on the graph having no cycles: the
// recursive term's own "WHERE t.hop < ?" is checked against a hop counter that increases by
// exactly 1 per recursive step, so the whole query produces at most `depth` recursive
// iterations no matter how many edges (or cycles) exist — a cyclic graph revisits nodes at
// increasing hop counts instead of looping forever. UNION ALL is intentional (cheaper than
// UNION here); duplicate edges reached via different paths are deduplicated by the caller.
const recursiveTraversalQuery = `
WITH RECURSIVE traversal(id, from_kind, from_id, to_kind, to_id, relation, created_at, hop) AS (
	SELECT id, from_kind, from_id, to_kind, to_id, relation, created_at, 1
	FROM graph_edges
	WHERE from_kind = ? AND from_id = ?
	UNION ALL
	SELECT e.id, e.from_kind, e.from_id, e.to_kind, e.to_id, e.relation, e.created_at, t.hop + 1
	FROM graph_edges e
	JOIN traversal t ON e.from_kind = t.to_kind AND e.from_id = t.to_id
	WHERE t.hop < ?
)
SELECT id, from_kind, from_id, to_kind, to_id, relation, created_at FROM traversal;`

const schema = `
CREATE TABLE IF NOT EXISTS graph_edges (
	id         TEXT PRIMARY KEY,
	from_kind  TEXT NOT NULL,
	from_id    TEXT NOT NULL,
	to_kind    TEXT NOT NULL,
	to_id      TEXT NOT NULL,
	relation   TEXT NOT NULL,
	created_at TEXT NOT NULL
);`

// Store is a memory.Store backed by SQLite (modernc.org/sqlite, pure Go, no cgo), storing
// directed edges between existing memories.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the database at path and ensures the schema exists.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("graph: open %q: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d;", busyTimeoutMS)); err != nil {
		db.Close()
		return nil, fmt.Errorf("graph: set busy_timeout: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("graph: schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// Write inserts a directed edge m.FromID -> m.ToID labeled m.Relation, generating an ID
// when the caller didn't supply one. It validates Relation is shaped like a label, not
// prose (non-empty, under maxRelationLength, no newline), returning a validation error
// before touching the database when it isn't. It does not check that FromID/ToID
// reference memories that actually exist — the graph layer has no read access to the
// .okf/vector layers that own them.
func (s *Store) Write(ctx context.Context, m memory.Memory) (memory.Memory, error) {
	if err := validateRelation(m.Relation); err != nil {
		return memory.Memory{}, err
	}
	if m.ID == "" {
		m.ID = newID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	// from_kind/to_kind disambiguate which layer (.okf/vector) each endpoint's id belongs
	// to (docs/planning/specs/03-graph-storage-schema.md's (kind, id) composite reference)
	// — ids have no shared namespace across layers.
	//
	// ON CONFLICT upserts on a repeated explicit ID, mirroring okf.Store's own
	// ON CONFLICT(id) DO UPDATE (issue #27 — graph_edges previously had no such clause, so a
	// repeated ID errored instead of upserting like the other two layers).
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO graph_edges (id, from_kind, from_id, to_kind, to_id, relation, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET from_kind = excluded.from_kind, from_id = excluded.from_id,
		 to_kind = excluded.to_kind, to_id = excluded.to_id, relation = excluded.relation,
		 created_at = excluded.created_at`,
		m.ID, m.FromKind, m.FromID, m.ToKind, m.ToID, m.Relation,
		m.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return memory.Memory{}, fmt.Errorf("graph: write: %w", err)
	}
	m.Kind = memory.KindGraph
	return m, nil
}

// Query returns edges reachable from (q.FromKind, q.FromID). q.Depth <= 1 (including unset,
// 0) returns only direct outgoing edges. q.Depth > 1 walks the graph via a recursive SQL
// traversal, but never further than maxDepth hops — min(q.Depth, maxDepth), the caller's
// requested depth is never trusted past the server-side cap (see maxDepth's comment).
// Similarity is always 1: an edge is an exact structural match, not a ranked recall — the
// same convention okf.Store.Query already uses.
func (s *Store) Query(ctx context.Context, q memory.Query) ([]memory.Recall, error) {
	depth := q.Depth
	if depth <= 0 {
		depth = 1
	}
	if depth > maxDepth {
		depth = maxDepth
	}

	var (
		rows *sql.Rows
		err  error
	)
	if depth <= 1 {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, from_kind, from_id, to_kind, to_id, relation, created_at
			 FROM graph_edges WHERE from_kind = ? AND from_id = ?`,
			q.FromKind, q.FromID)
	} else {
		rows, err = s.db.QueryContext(ctx, recursiveTraversalQuery, q.FromKind, q.FromID, depth)
	}
	if err != nil {
		return nil, fmt.Errorf("graph: query: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	var out []memory.Recall
	for rows.Next() {
		var (
			m         memory.Memory
			createdAt string
		)
		if err := rows.Scan(&m.ID, &m.FromKind, &m.FromID, &m.ToKind, &m.ToID, &m.Relation, &createdAt); err != nil {
			return nil, fmt.Errorf("graph: scan: %w", err)
		}
		if seen[m.ID] {
			continue
		}
		seen[m.ID] = true
		m.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		m.Kind = memory.KindGraph
		out = append(out, memory.Recall{Memory: m, Similarity: 1})
	}
	return out, rows.Err()
}

// validateRelation rejects anything that isn't shaped like a short label: empty, over the
// cap, or containing a newline (i.e. free prose someone stuffed into the wrong field).
func validateRelation(relation string) error {
	if relation == "" {
		return fmt.Errorf("graph: relation must not be empty")
	}
	if len(relation) > maxRelationLength {
		return fmt.Errorf("graph: relation exceeds %d bytes", maxRelationLength)
	}
	if strings.ContainsAny(relation, "\n\r") {
		return fmt.Errorf("graph: relation must not contain a newline")
	}
	return nil
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
