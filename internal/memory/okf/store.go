// Package okf is the memory contract's declarative layer (memory.KindOKF): facts and
// rules, looked up by tag, not embedded. "Différenciateur souverain" per the état-de-l'art
// doc — relisible, auditable, versionable, unlike a vector's opaque float array.
package okf

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yoann/kern-memory/internal/memory"
)

// busyTimeoutMS mirrors internal/store/sqlite.go — same lesson, same fix, this repo's own
// established convention (docs/retro.md, 2026-07-30).
const busyTimeoutMS = 5000

const schema = `
CREATE TABLE IF NOT EXISTS okf_memories (
	id         TEXT PRIMARY KEY,
	text       TEXT NOT NULL,
	tags_json  TEXT NOT NULL,
	meta_json  TEXT NOT NULL,
	created_at TEXT NOT NULL
);`

// Store is a memory.Store backed by SQLite (modernc.org/sqlite, pure Go, no cgo).
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the database at path and ensures the schema exists.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("okf: open %q: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d;", busyTimeoutMS)); err != nil {
		db.Close()
		return nil, fmt.Errorf("okf: set busy_timeout: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("okf: schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// Write inserts m, generating an ID when the caller didn't supply one. Writing an id that
// already exists UPSERTS (overwrites text/tags/metadata/created_at) rather than erroring —
// a caller with a stable id (e.g. Kern-UI's marketing calendar, keyed by run id) needs to
// re-write the same memory as its content changes, not track "does this id exist yet".
func (s *Store) Write(ctx context.Context, m memory.Memory) (memory.Memory, error) {
	if m.ID == "" {
		m.ID = newID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	tagsJSON, err := json.Marshal(m.Tags)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("okf: marshal tags: %w", err)
	}
	metaJSON, err := json.Marshal(m.Metadata)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("okf: marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO okf_memories (id, text, tags_json, meta_json, created_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET text = excluded.text, tags_json = excluded.tags_json,
		 meta_json = excluded.meta_json, created_at = excluded.created_at`,
		m.ID, m.Text, string(tagsJSON), string(metaJSON), m.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return memory.Memory{}, fmt.Errorf("okf: write: %w", err)
	}
	return m, nil
}

// Query returns every memory whose Tags is a superset of q.Tags (all requested tags must
// be present), or every memory when q.Tags is empty. Similarity is always 1 — an exact
// lookup, not a ranked recall.
func (s *Store) Query(ctx context.Context, q memory.Query) ([]memory.Recall, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, text, tags_json, meta_json, created_at FROM okf_memories ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("okf: query: %w", err)
	}
	defer rows.Close()

	var out []memory.Recall
	for rows.Next() {
		var (
			m                  memory.Memory
			tagsJSON, metaJSON string
			createdAt          string
		)
		if err := rows.Scan(&m.ID, &m.Text, &tagsJSON, &metaJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("okf: scan: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsJSON), &m.Tags); err != nil {
			return nil, fmt.Errorf("okf: unmarshal tags: %w", err)
		}
		if err := json.Unmarshal([]byte(metaJSON), &m.Metadata); err != nil {
			return nil, fmt.Errorf("okf: unmarshal metadata: %w", err)
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		m.Kind = memory.KindOKF

		if !hasAllTags(m.Tags, q.Tags) {
			continue
		}
		out = append(out, memory.Recall{Memory: m, Similarity: 1})
	}
	return out, rows.Err()
}

func hasAllTags(has, want []string) bool {
	for _, w := range want {
		found := false
		for _, h := range has {
			if strings.EqualFold(h, w) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
