package store

import (
	"context"
	"path/filepath"
	"testing"
)

func open(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "kern-memory.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestListIsEmptyOnAFreshStore(t *testing.T) {
	s := open(t)
	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("list = %v, want empty", got)
	}
}

func TestSeedThenListReportsTheWordCount(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	if err := s.Seed(ctx, "doc1", "Compte-rendu", "un deux trois quatre"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := s.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("list = %v, want one document", got)
	}
	if got[0].ID != "doc1" || got[0].Title != "Compte-rendu" || got[0].WordCount != 4 {
		t.Fatalf("list[0] = %+v, want id=doc1 title=Compte-rendu wordCount=4", got[0])
	}
}

func TestGetUnknownDocumentFails(t *testing.T) {
	s := open(t)
	_, err := s.Get(context.Background(), "nope")
	if err != ErrUnknownDocument {
		t.Fatalf("get unknown = %v, want ErrUnknownDocument", err)
	}
}

func TestGetReturnsSuggestionsOrderedByAnchor(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	if err := s.Seed(ctx, "doc1", "Titre", "Une phrase à corriger ici."); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := s.SeedSuggestion(ctx, Suggestion{ID: "s2", DocumentID: "doc1", AnchorStart: 10, AnchorEnd: 20, Text: "second"}); err != nil {
		t.Fatalf("seed suggestion: %v", err)
	}
	if err := s.SeedSuggestion(ctx, Suggestion{ID: "s1", DocumentID: "doc1", AnchorStart: 0, AnchorEnd: 4, Text: "first"}); err != nil {
		t.Fatalf("seed suggestion: %v", err)
	}

	doc, err := s.Get(ctx, "doc1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(doc.Suggestions) != 2 || doc.Suggestions[0].ID != "s1" || doc.Suggestions[1].ID != "s2" {
		t.Fatalf("suggestions = %+v, want [s1 s2] in anchor order", doc.Suggestions)
	}
	if doc.Suggestions[0].Status != StatusPending {
		t.Fatalf("suggestion status = %q, want pending default", doc.Suggestions[0].Status)
	}
}

func TestResolveAcceptsAPendingSuggestion(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_ = s.Seed(ctx, "doc1", "Titre", "corps")
	_ = s.SeedSuggestion(ctx, Suggestion{ID: "s1", DocumentID: "doc1", Text: "x"})

	if err := s.Resolve(ctx, "doc1", "s1", StatusAccepted); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	doc, err := s.Get(ctx, "doc1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if doc.Suggestions[0].Status != StatusAccepted {
		t.Fatalf("status = %q, want accepted", doc.Suggestions[0].Status)
	}
}

func TestResolveUnknownDocumentFails(t *testing.T) {
	s := open(t)
	err := s.Resolve(context.Background(), "nope", "s1", StatusAccepted)
	if err != ErrUnknownDocument {
		t.Fatalf("resolve unknown document = %v, want ErrUnknownDocument", err)
	}
}

func TestResolveUnknownSuggestionFails(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	_ = s.Seed(ctx, "doc1", "Titre", "corps")

	err := s.Resolve(ctx, "doc1", "nope", StatusAccepted)
	if err != ErrUnknownSuggestion {
		t.Fatalf("resolve unknown suggestion = %v, want ErrUnknownSuggestion", err)
	}
}
