package graph

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/yoann/kern-memory/internal/memory"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir() + "/graph.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpenCreatesTheSchema(t *testing.T) {
	path := t.TempDir() + "/graph.db"
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := s.db.Exec("SELECT id, from_kind, from_id, to_kind, to_id, relation, created_at FROM graph_edges"); err != nil {
		t.Errorf("graph_edges table not usable: %v", err)
	}
}

func TestOpenIsIdempotentAgainstAnExistingDatabase(t *testing.T) {
	path := t.TempDir() + "/graph.db"
	s1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open (should be idempotent): %v", err)
	}
	defer s2.Close()
}

func TestWritePersistsARowReadableDirectly(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	out, err := s.Write(ctx, memory.Memory{
		Kind:     memory.KindGraph,
		FromID:   "okf-abc",
		ToID:     "vector-def",
		Relation: "supports",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if out.ID == "" {
		t.Fatal("expected a generated ID")
	}

	var fromID, toID, relation string
	row := s.db.QueryRowContext(ctx,
		"SELECT from_id, to_id, relation FROM graph_edges WHERE id = ?", out.ID)
	if err := row.Scan(&fromID, &toID, &relation); err != nil {
		t.Fatalf("direct read: %v", err)
	}
	if fromID != "okf-abc" || toID != "vector-def" || relation != "supports" {
		t.Errorf("got (%q, %q, %q), want (okf-abc, vector-def, supports)", fromID, toID, relation)
	}
}

func TestWritePersistsFromKindAndToKind(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	out, err := s.Write(ctx, memory.Memory{
		Kind:     memory.KindGraph,
		FromID:   "okf-abc",
		FromKind: "okf",
		ToID:     "vector-def",
		ToKind:   "vector",
		Relation: "supports",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var fromKind, toKind string
	row := s.db.QueryRowContext(ctx,
		"SELECT from_kind, to_kind FROM graph_edges WHERE id = ?", out.ID)
	if err := row.Scan(&fromKind, &toKind); err != nil {
		t.Fatalf("direct read: %v", err)
	}
	if fromKind != "okf" || toKind != "vector" {
		t.Errorf("got (from_kind=%q, to_kind=%q), want (okf, vector)", fromKind, toKind)
	}
}

func TestWriteKeepsAGivenID(t *testing.T) {
	s := open(t)
	out, err := s.Write(context.Background(), memory.Memory{
		ID: "edge-1", FromID: "a", ToID: "b", Relation: "rel",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if out.ID != "edge-1" {
		t.Errorf("ID = %q, want the caller's own", out.ID)
	}
}

func TestWriteRejectsAnEmptyRelation(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	_, err := s.Write(ctx, memory.Memory{FromID: "a", ToID: "b", Relation: ""})
	if err == nil {
		t.Fatal("expected an error for empty Relation")
	}
	assertNoRows(t, s.db)
}

func TestWriteRejectsARelationLongerThanTheCap(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	_, err := s.Write(ctx, memory.Memory{FromID: "a", ToID: "b", Relation: strings.Repeat("x", maxRelationLength+1)})
	if err == nil {
		t.Fatal("expected an error for an overlong Relation")
	}
	assertNoRows(t, s.db)
}

func TestWriteRejectsARelationContainingANewline(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	_, err := s.Write(ctx, memory.Memory{FromID: "a", ToID: "b", Relation: "supports\nprose"})
	if err == nil {
		t.Fatal("expected an error for a Relation containing a newline")
	}
	assertNoRows(t, s.db)
}

func assertNoRows(t *testing.T, db *sql.DB) {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM graph_edges").Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("got %d rows, want 0 (rejected write must not insert)", count)
	}
}
