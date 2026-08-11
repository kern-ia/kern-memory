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
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO graph_edges (id, from_kind, from_id, to_kind, to_id, relation, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.FromKind, m.FromID, m.ToKind, m.ToID, m.Relation,
		m.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return memory.Memory{}, fmt.Errorf("graph: write: %w", err)
	}
	m.Kind = memory.KindGraph
	return m, nil
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
