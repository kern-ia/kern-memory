package graph

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

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

// TestWriteWithARepeatedGivenIDFailsRatherThanUpserting documents a real gap found while
// building epic-2 issue 1's load-memory loader: unlike okf.Store (ON CONFLICT DO UPDATE)
// and vector.Store (ID-keyed map overwrite), graph_edges' schema has no ON CONFLICT clause,
// so a second Write with the same explicit ID hits the PRIMARY KEY constraint and errors —
// loudly, not silently, but it is not an upsert.
func TestWriteWithARepeatedGivenIDFailsRatherThanUpserting(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	if _, err := s.Write(ctx, memory.Memory{ID: "edge-1", FromID: "a", ToID: "b", Relation: "rel"}); err != nil {
		t.Fatalf("first Write: %v", err)
	}

	_, err := s.Write(ctx, memory.Memory{ID: "edge-1", FromID: "a", ToID: "c", Relation: "rel2"})
	if err == nil {
		t.Fatal("expected an error on a repeated explicit ID (graph_edges has no ON CONFLICT clause), got nil")
	}
}

// TestRepeatedWriteWithNoIDCreatesASeparateEdgeNotAnUpsert documents the other half of the
// same gap: the load-memory file format's edge entries carry no id field (see the issue's
// example), so a caller re-running the loader on an unchanged file leaves every edge's ID
// empty — Write then generates a fresh random ID each time, and the same logical edge
// accumulates as multiple rows rather than upserting.
func TestRepeatedWriteWithNoIDCreatesASeparateEdgeNotAnUpsert(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	if _, err := s.Write(ctx, memory.Memory{FromID: "a", ToID: "b", Relation: "rel"}); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	if _, err := s.Write(ctx, memory.Memory{FromID: "a", ToID: "b", Relation: "rel"}); err != nil {
		t.Fatalf("second Write: %v", err)
	}

	out, err := s.Query(ctx, memory.Query{FromKind: "", FromID: "a"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	count := 0
	for _, r := range out {
		if r.Memory.ToID == "b" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("got %d edges a->b after two identical writes, want 2 (no ID means no upsert, this is a real gap, not a bug in this test)", count)
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

func writeEdge(t *testing.T, s *Store, fromKind, fromID, toKind, toID, relation string) {
	t.Helper()
	if _, err := s.Write(context.Background(), memory.Memory{
		FromKind: fromKind, FromID: fromID, ToKind: toKind, ToID: toID, Relation: relation,
	}); err != nil {
		t.Fatalf("seed Write(%s->%s): %v", fromID, toID, err)
	}
}

func hasEdgeTo(recalls []memory.Recall, toID string) bool {
	for _, r := range recalls {
		if r.Memory.ToID == toID {
			return true
		}
	}
	return false
}

func TestQueryDirectEdgeIsReturnedForDepthUnsetOrOne(t *testing.T) {
	s := open(t)
	writeEdge(t, s, "okf", "a", "okf", "b", "supports")

	for _, depth := range []int{0, 1} {
		out, err := s.Query(context.Background(), memory.Query{
			Kind: memory.KindGraph, FromKind: "okf", FromID: "a", Depth: depth,
		})
		if err != nil {
			t.Fatalf("Query(Depth=%d): %v", depth, err)
		}
		if !hasEdgeTo(out, "b") {
			t.Errorf("Depth=%d: expected an edge to %q, got %+v", depth, "b", out)
		}
	}
}

func TestQueryDepth2ReachesTwoHopsButDepth1DoesNot(t *testing.T) {
	s := open(t)
	writeEdge(t, s, "okf", "a", "okf", "b", "rel")
	writeEdge(t, s, "okf", "b", "okf", "c", "rel")

	out1, err := s.Query(context.Background(), memory.Query{FromKind: "okf", FromID: "a", Depth: 1})
	if err != nil {
		t.Fatalf("Query(Depth=1): %v", err)
	}
	if hasEdgeTo(out1, "c") {
		t.Errorf("Depth=1: expected NOT to reach c, got %+v", out1)
	}

	out2, err := s.Query(context.Background(), memory.Query{FromKind: "okf", FromID: "a", Depth: 2})
	if err != nil {
		t.Fatalf("Query(Depth=2): %v", err)
	}
	if !hasEdgeTo(out2, "c") {
		t.Errorf("Depth=2: expected to reach c, got %+v", out2)
	}
}

// TestQueryDepthAboveMaxIsClampedToTheServerSideMax proves the server-side depth clamp
// itself, not just that a large Depth doesn't crash: it builds a chain one hop longer than
// maxDepth and shows that asking for a Depth far beyond maxDepth (100) returns exactly the
// same edges as asking for maxDepth — the extra hop past the cap is never reached. If the
// clamp (min(q.Depth, maxDepth)) were removed, Depth: 100 would walk the whole chain and
// this test would fail by finding the extra node.
func TestQueryDepthAboveMaxIsClampedToTheServerSideMax(t *testing.T) {
	s := open(t)
	// Chain of maxDepth+1 hops: a -> b -> c -> d -> e (maxDepth == 3, so "e" sits 4 hops
	// from "a" — one hop past the cap).
	writeEdge(t, s, "okf", "a", "okf", "b", "rel")
	writeEdge(t, s, "okf", "b", "okf", "c", "rel")
	writeEdge(t, s, "okf", "c", "okf", "d", "rel")
	writeEdge(t, s, "okf", "d", "okf", "e", "rel")

	atMax, err := s.Query(context.Background(), memory.Query{FromKind: "okf", FromID: "a", Depth: maxDepth})
	if err != nil {
		t.Fatalf("Query(Depth=maxDepth): %v", err)
	}
	aboveMax, err := s.Query(context.Background(), memory.Query{FromKind: "okf", FromID: "a", Depth: 100})
	if err != nil {
		t.Fatalf("Query(Depth=100): %v", err)
	}

	if len(atMax) != len(aboveMax) {
		t.Fatalf("clamp not applied: Depth=maxDepth got %d edges, Depth=100 got %d edges", len(atMax), len(aboveMax))
	}
	if hasEdgeTo(aboveMax, "e") {
		t.Errorf("clamp not applied: Depth=100 reached %q, which is one hop past maxDepth=%d", "e", maxDepth)
	}
}

// TestQueryOnACyclicGraphTerminates proves the recursive traversal is cycle-safe: a naive
// WITH RECURSIVE walk with no depth bound would loop forever on this 2-node cycle. The
// depth-bounded recursion (WHERE hop < depth in the recursive term) guarantees termination
// after a fixed number of steps regardless of graph structure — this test fails by timeout
// (not by a wrong answer) if that guarantee is ever lost.
func TestQueryOnACyclicGraphTerminates(t *testing.T) {
	s := open(t)
	writeEdge(t, s, "okf", "a", "okf", "b", "rel")
	writeEdge(t, s, "okf", "b", "okf", "a", "rel")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	var out []memory.Recall
	var err error
	go func() {
		out, err = s.Query(ctx, memory.Query{FromKind: "okf", FromID: "a", Depth: 1000})
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatalf("Query on cyclic graph: %v", err)
		}
		if !hasEdgeTo(out, "b") {
			t.Errorf("expected to reach b, got %+v", out)
		}
	case <-ctx.Done():
		t.Fatal("Query on a cyclic graph did not terminate within 5s — depth clamp/cycle-safety broken")
	}
}

func TestQueryOnANodeWithNoEdgesReturnsEmptyNotError(t *testing.T) {
	s := open(t)
	out, err := s.Query(context.Background(), memory.Query{FromKind: "okf", FromID: "lonely", Depth: 3})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty result, got %+v", out)
	}
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
