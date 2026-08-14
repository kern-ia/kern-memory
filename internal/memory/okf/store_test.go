package okf

import (
	"context"
	"testing"

	"github.com/yoann/kern-memory/internal/memory"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir() + "/okf.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestWriteAssignsAnIDWhenNoneIsGiven(t *testing.T) {
	s := open(t)

	out, err := s.Write(context.Background(), memory.Memory{Text: "le taux d'usure change chaque trimestre"})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if out.ID == "" {
		t.Error("expected a generated ID")
	}
}

func TestWriteKeepsAGivenID(t *testing.T) {
	s := open(t)

	out, err := s.Write(context.Background(), memory.Memory{ID: "règle-usure", Text: "x"})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if out.ID != "règle-usure" {
		t.Errorf("ID = %q, want the caller's own", out.ID)
	}
}

func TestWriteUpsertsOnAnExistingID(t *testing.T) {
	// A caller with a stable id (e.g. Kern-UI's marketing calendar, keyed by run id) must
	// be able to re-write the same memory as its content changes over time — a plain
	// INSERT would fail on the PRIMARY KEY the second time.
	s := open(t)
	ctx := context.Background()
	_, err := s.Write(ctx, memory.Memory{ID: "item-1", Text: "version 1", Tags: []string{"a"}})
	if err != nil {
		t.Fatalf("first Write: %v", err)
	}

	_, err = s.Write(ctx, memory.Memory{ID: "item-1", Text: "version 2", Tags: []string{"b"}})
	if err != nil {
		t.Fatalf("second Write (upsert): %v", err)
	}

	got, err := s.Query(ctx, memory.Query{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d memories, want 1 (upsert must not duplicate the row)", len(got))
	}
	if got[0].Memory.Text != "version 2" || got[0].Memory.Tags[0] != "b" {
		t.Errorf("got %+v, want the second write's content", got[0].Memory)
	}
}

func TestQueryByTagReturnsOnlyMatchingMemories(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "critère banque A", Tags: []string{"banque-a", "sci"}})
	_, _ = s.Write(ctx, memory.Memory{ID: "b", Text: "critère banque B", Tags: []string{"banque-b"}})

	got, err := s.Query(ctx, memory.Query{Tags: []string{"sci"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want only memory a", got)
	}
}

func TestQueryWithoutTagsReturnsEverything(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "x", Tags: []string{"t1"}})
	_, _ = s.Write(ctx, memory.Memory{ID: "b", Text: "y", Tags: []string{"t2"}})

	got, err := s.Query(ctx, memory.Query{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d recalls, want 2", len(got))
	}
}

func TestQueryResultsCarryFullSimilarityAndTheirTags(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "x", Tags: []string{"sci", "hypothécaire"}})

	got, err := s.Query(ctx, memory.Query{Tags: []string{"sci"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1", len(got))
	}
	if got[0].Similarity != 1 {
		t.Errorf("Similarity = %v, want 1 (exact lookup, no embedding)", got[0].Similarity)
	}
	if len(got[0].Memory.Tags) != 2 {
		t.Errorf("Tags = %v, want both tags preserved", got[0].Memory.Tags)
	}
}

// Resolving known ids (decision 16) — a lookup, not a search.
func TestQueryByIDsReturnsExactlyThoseMemories(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "x", Tags: []string{"t1"}})
	_, _ = s.Write(ctx, memory.Memory{ID: "b", Text: "y", Tags: []string{"t2"}})
	_, _ = s.Write(ctx, memory.Memory{ID: "c", Text: "z", Tags: []string{"t3"}})

	got, err := s.Query(ctx, memory.Query{IDs: []string{"a", "c"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2", len(got))
	}
	ids := map[string]bool{got[0].Memory.ID: true, got[1].Memory.ID: true}
	if !ids["a"] || !ids["c"] {
		t.Errorf("got %v, want a and c", got)
	}
}

// An id that does not exist is silently absent from the result, not an error — the same
// "best-effort resolve" a caller batching several ids across an evolving graph needs.
func TestQueryByIDsIgnoresAnUnknownID(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "x"})

	got, err := s.Query(ctx, memory.Query{IDs: []string{"a", "jamais-ecrit"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want only a", got)
	}
}

// IDs takes precedence over Tags when both are set — a resolve call has nothing to
// search for, it already knows what it wants.
func TestQueryByIDsIgnoresTagsWhenBothAreSet(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "x", Tags: []string{"sci"}})

	got, err := s.Query(ctx, memory.Query{IDs: []string{"a"}, Tags: []string{"aucun-rapport"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want a regardless of the unrelated tag filter", got)
	}
}

func TestQueryByMultipleTagsRequiresAllOfThem(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, _ = s.Write(ctx, memory.Memory{ID: "a", Text: "x", Tags: []string{"sci", "senior"}})
	_, _ = s.Write(ctx, memory.Memory{ID: "b", Text: "y", Tags: []string{"sci"}})

	got, err := s.Query(ctx, memory.Query{Tags: []string{"sci", "senior"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want only memory a (has both tags)", got)
	}
}

func TestSurvivesAProcessRestart(t *testing.T) {
	dir := t.TempDir() + "/okf.db"
	s1, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, _ = s1.Write(context.Background(), memory.Memory{ID: "a", Text: "persisté", Tags: []string{"t"}})
	_ = s1.Close()

	s2, err := Open(dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()

	got, err := s2.Query(context.Background(), memory.Query{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.Text != "persisté" {
		t.Errorf("got %v, want the memory written before restart", got)
	}
}
