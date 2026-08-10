package anon

import (
	"context"
	"strings"
	"testing"

	"github.com/YoLaub/PresidioGo/pii"
	"github.com/yoann/kern-memory/internal/memory"
)

// fakeStore records exactly what Write was called with, so tests can assert on what
// actually reached the wrapped layer, not what the caller passed to the decorator.
type fakeStore struct {
	written []memory.Memory
	recalls []memory.Recall
}

func (f *fakeStore) Write(_ context.Context, m memory.Memory) (memory.Memory, error) {
	f.written = append(f.written, m)
	return m, nil
}

func (f *fakeStore) Query(_ context.Context, _ memory.Query) ([]memory.Recall, error) {
	return f.recalls, nil
}

func TestWriteMasksPIIBeforeDelegating(t *testing.T) {
	inner := &fakeStore{}
	s := Wrap(inner)

	_, err := s.Write(context.Background(), memory.Memory{
		Text: "Contact du client : jean.dupont@example.com, IBAN FR7630006000011234567890189.",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(inner.written) != 1 {
		t.Fatalf("inner.written = %d, want 1", len(inner.written))
	}
	got := inner.written[0].Text
	if strings.Contains(got, "jean.dupont@example.com") {
		t.Errorf("text still contains the raw email: %q", got)
	}
	if strings.Contains(got, "FR7630006000011234567890189") {
		t.Errorf("text still contains the raw IBAN: %q", got)
	}
}

func TestWriteLeavesOrdinaryTextIntact(t *testing.T) {
	inner := &fakeStore{}
	s := Wrap(inner)

	_, err := s.Write(context.Background(), memory.Memory{Text: "Le taux d'usure évolue chaque trimestre."})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if inner.written[0].Text != "Le taux d'usure évolue chaque trimestre." {
		t.Errorf("text with no PII should pass through unchanged, got %q", inner.written[0].Text)
	}
}

func TestQueryPassesThroughWithoutTransformation(t *testing.T) {
	want := []memory.Recall{{Memory: memory.Memory{ID: "a", Text: "déjà masqué au repos"}, Similarity: 1}}
	inner := &fakeStore{recalls: want}
	s := Wrap(inner)

	got, err := s.Query(context.Background(), memory.Query{Text: "x"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Errorf("got %v, want the inner store's recalls unchanged", got)
	}
}

func TestFilterNerScopeKeepsPersonAndRegexEntitiesButDropsLocationAndOrganization(t *testing.T) {
	in := []pii.Result{
		{EntityType: "PERSON", Start: 0, End: 4},
		{EntityType: "LOCATION", Start: 5, End: 9},
		{EntityType: "ORGANIZATION", Start: 10, End: 14},
		{EntityType: "IBAN_CODE", Start: 15, End: 20},
	}

	got := filterNerScope(in)

	var types []string
	for _, r := range got {
		types = append(types, r.EntityType)
	}
	if len(types) != 2 || types[0] != "PERSON" || types[1] != "IBAN_CODE" {
		t.Errorf("got %v, want [PERSON IBAN_CODE]", types)
	}
}

func TestWritePreservesEveryOtherField(t *testing.T) {
	inner := &fakeStore{}
	s := Wrap(inner)

	_, err := s.Write(context.Background(), memory.Memory{
		ID: "x", Kind: memory.KindOKF, Text: "sans PII", Tags: []string{"t"},
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := inner.written[0]
	if got.ID != "x" || got.Kind != memory.KindOKF || len(got.Tags) != 1 {
		t.Errorf("got %+v, want ID/Kind/Tags preserved", got)
	}
}
