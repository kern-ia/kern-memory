package vector

import (
	"context"
	"os/exec"
	"testing"

	"github.com/yoann/kern-memory/internal/memory"
)

// requireOllama skips real-embedding tests when Ollama isn't running locally — these are
// integration tests against a real service, not mocked (chromem-go has no fake embedding
// path worth testing against; the value here is the real vector math, not our own code).
func requireOllama(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ollama"); err != nil {
		t.Skip("ollama not installed, skipping real-embedding test")
	}
}

func open(t *testing.T) *Store {
	t.Helper()
	requireOllama(t)
	s, err := Open(t.TempDir()+"/vector.db", "nomic-embed-text", "")
	if err != nil {
		t.Skipf("vector.Open (is Ollama serving nomic-embed-text?): %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestWriteAssignsAnIDWhenNoneIsGiven(t *testing.T) {
	s := open(t)

	out, err := s.Write(context.Background(), memory.Memory{Text: "la banque X accepte les SCI"})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if out.ID == "" {
		t.Error("expected a generated ID")
	}
}

func TestQueryFindsTheMostSemanticallySimilarMemory(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	_, err := s.Write(ctx, memory.Memory{
		ID:   "banque-a",
		Text: "La banque Alpha accepte les prêts hypothécaires pour un bien détenu en SCI, y compris pour des emprunteurs de plus de 70 ans.",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err = s.Write(ctx, memory.Memory{
		ID:   "recette",
		Text: "Recette de tarte aux pommes : préparer la pâte, ajouter les pommes coupées, cuire 40 minutes.",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := s.Query(ctx, memory.Query{Text: "quelle banque accepte un prêt hypothécaire sur un bien en SCI pour un senior", Limit: 2})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one result")
	}
	if got[0].Memory.ID != "banque-a" {
		t.Errorf("top result = %q, want banque-a (the semantically relevant one)", got[0].Memory.ID)
	}
}

// Resolving known ids (decision 16) must not need an embedding call at all — GetByID is a
// direct lookup, not a similarity search, so this works even with an empty Query.Text.
func TestQueryByIDsResolvesWithoutAnEmbeddingCall(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, err := s.Write(ctx, memory.Memory{ID: "a", Text: "x", Tags: []string{"t1"}})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err = s.Write(ctx, memory.Memory{ID: "b", Text: "y"})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := s.Query(ctx, memory.Query{IDs: []string{"a"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" || got[0].Memory.Text != "x" {
		t.Errorf("got %v, want only memory a with its text", got)
	}
	if len(got[0].Memory.Tags) != 1 || got[0].Memory.Tags[0] != "t1" {
		t.Errorf("Tags = %v, want [t1]", got[0].Memory.Tags)
	}
}

// An id that was never written is silently absent — same "best-effort resolve"
// contract as the okf layer, not an error.
func TestQueryByIDsIgnoresAnUnknownID(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_, err := s.Write(ctx, memory.Memory{ID: "a", Text: "x"})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := s.Query(ctx, memory.Query{IDs: []string{"a", "jamais-ecrit"}})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want only a", got)
	}
}

// TestWriteUpsertsOnARepeatedID confirms the vector layer behaves like okf.Store on a
// repeat write with the same id (load-memory's re-run case, epic-2 issue 1): chromem-go's
// AddDocument stores documents in an ID-keyed map (c.documents[doc.ID] = &doc), so writing
// the same ID twice overwrites in place rather than erroring or duplicating.
func TestWriteUpsertsOnARepeatedID(t *testing.T) {
	s := open(t)
	ctx := context.Background()

	if _, err := s.Write(ctx, memory.Memory{ID: "item-1", Text: "version 1", Tags: []string{"a"}}); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	if _, err := s.Write(ctx, memory.Memory{ID: "item-1", Text: "version 2", Tags: []string{"b"}}); err != nil {
		t.Fatalf("second Write (upsert): %v", err)
	}

	got, err := s.Query(ctx, memory.Query{Text: "version 2", Limit: 10})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	count := 0
	for _, r := range got {
		if r.Memory.ID == "item-1" {
			count++
			if r.Memory.Text != "version 2" {
				t.Errorf("got Text=%q, want the second write's content", r.Memory.Text)
			}
		}
	}
	if count != 1 {
		t.Errorf("got %d recalls for id item-1, want 1 (upsert must not duplicate)", count)
	}
}

func TestSurvivesAProcessRestart(t *testing.T) {
	requireOllama(t)
	dir := t.TempDir() + "/vector.db"
	s1, err := Open(dir, "nomic-embed-text", "")
	if err != nil {
		t.Skipf("vector.Open: %v", err)
	}
	if _, err := s1.Write(context.Background(), memory.Memory{ID: "a", Text: "persisté à travers un redémarrage"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = s1.Close()

	s2, err := Open(dir, "nomic-embed-text", "")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()

	got, err := s2.Query(context.Background(), memory.Query{Text: "persisté à travers un redémarrage", Limit: 1})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want the memory written before restart", got)
	}
}
