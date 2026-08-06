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
