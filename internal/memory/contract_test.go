package memory

import (
	"context"
	"errors"
	"testing"
)

// fakeStore is a minimal in-memory Store for testing Router's routing logic in isolation
// from any real backend (SQLite .okf layer, chromem-go vector layer).
type fakeStore struct {
	written []Memory
	recalls []Recall
	err     error
}

func (f *fakeStore) Write(_ context.Context, m Memory) (Memory, error) {
	if f.err != nil {
		return Memory{}, f.err
	}
	f.written = append(f.written, m)
	return m, nil
}

func (f *fakeStore) Query(_ context.Context, _ Query) ([]Recall, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.recalls, nil
}

func TestRouterWriteRoutesOKFKindToTheOKFStore(t *testing.T) {
	okf, vec := &fakeStore{}, &fakeStore{}
	r := &Router{OKF: okf, Vector: vec}

	if _, err := r.Write(context.Background(), Memory{Kind: KindOKF, Text: "règle métier"}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(okf.written) != 1 || len(vec.written) != 0 {
		t.Errorf("okf.written=%d vec.written=%d, want 1/0", len(okf.written), len(vec.written))
	}
}

func TestRouterWriteRoutesVectorKindToTheVectorStore(t *testing.T) {
	okf, vec := &fakeStore{}, &fakeStore{}
	r := &Router{OKF: okf, Vector: vec}

	if _, err := r.Write(context.Background(), Memory{Kind: KindVector, Text: "un souvenir"}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(vec.written) != 1 || len(okf.written) != 0 {
		t.Errorf("vec.written=%d okf.written=%d, want 1/0", len(vec.written), len(okf.written))
	}
}

func TestRouterWriteDefaultsToVectorWhenKindIsEmpty(t *testing.T) {
	okf, vec := &fakeStore{}, &fakeStore{}
	r := &Router{OKF: okf, Vector: vec}

	if _, err := r.Write(context.Background(), Memory{Text: "sans kind précisé"}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(vec.written) != 1 {
		t.Errorf("vec.written=%d, want 1 (default kind)", len(vec.written))
	}
}

func TestRouterWriteRejectsAnUnknownKind(t *testing.T) {
	r := &Router{OKF: &fakeStore{}, Vector: &fakeStore{}}

	if _, err := r.Write(context.Background(), Memory{Kind: "n_importe_quoi"}); err == nil {
		t.Fatal("expected an error for an unknown kind")
	}
}

func TestRouterQueryRestrictsToTheRequestedKind(t *testing.T) {
	okf := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "fait déclaratif"}, Similarity: 1}}}
	vec := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "souvenir sémantique"}, Similarity: 0.8}}}
	r := &Router{OKF: okf, Vector: vec}

	got, err := r.Query(context.Background(), Query{Kind: KindOKF, Text: "x"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.Text != "fait déclaratif" {
		t.Errorf("got %v, want only the OKF layer's recall", got)
	}
}

func TestRouterQueryFansOutToBothLayersAndMergesByKindWhenUnrestricted(t *testing.T) {
	okf := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "fait"}, Similarity: 1}}}
	vec := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "souvenir haut"}, Similarity: 0.9},
		{Memory: Memory{Text: "souvenir bas"}, Similarity: 0.4}}}
	r := &Router{OKF: okf, Vector: vec}

	got, err := r.Query(context.Background(), Query{Text: "x"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d recalls, want 3 (merged from both layers)", len(got))
	}
	// Sorted by similarity descending — the OKF exact fact (1.0) and the strongest vector
	// recall (0.9) must outrank the weak one (0.4).
	for i := 1; i < len(got); i++ {
		if got[i-1].Similarity < got[i].Similarity {
			t.Errorf("recalls not sorted by similarity desc: %v", got)
		}
	}
}

func TestRouterWriteRoutesGraphKindToTheGraphStore(t *testing.T) {
	okf, vec, graph := &fakeStore{}, &fakeStore{}, &fakeStore{}
	r := &Router{OKF: okf, Vector: vec, Graph: graph}

	if _, err := r.Write(context.Background(), Memory{Kind: KindGraph, FromID: "a", ToID: "b", Relation: "cites"}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(graph.written) != 1 || len(okf.written) != 0 || len(vec.written) != 0 {
		t.Errorf("graph.written=%d okf.written=%d vec.written=%d, want 1/0/0", len(graph.written), len(okf.written), len(vec.written))
	}
}

func TestRouterWriteWithUnsetGraphReturnsAClearErrorNotAPanic(t *testing.T) {
	r := &Router{OKF: &fakeStore{}, Vector: &fakeStore{}}

	_, err := r.Write(context.Background(), Memory{Kind: KindGraph, FromID: "a", ToID: "b", Relation: "cites"})
	if err == nil {
		t.Fatal("expected an error for an unset Router.Graph, got nil")
	}
}

func TestRouterQueryRoutesGraphKindToTheGraphStore(t *testing.T) {
	okf := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "fait déclaratif"}, Similarity: 1}}}
	vec := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "souvenir sémantique"}, Similarity: 0.8}}}
	graph := &fakeStore{recalls: []Recall{{Memory: Memory{FromID: "a", ToID: "b", Relation: "cites"}, Similarity: 1}}}
	r := &Router{OKF: okf, Vector: vec, Graph: graph}

	got, err := r.Query(context.Background(), Query{Kind: KindGraph, Text: "x"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Memory.Relation != "cites" {
		t.Errorf("got %v, want only the graph layer's recall", got)
	}
}

func TestRouterQueryWithUnsetGraphReturnsAClearErrorNotAPanic(t *testing.T) {
	r := &Router{OKF: &fakeStore{}, Vector: &fakeStore{}}

	_, err := r.Query(context.Background(), Query{Kind: KindGraph, Text: "x"})
	if err == nil {
		t.Fatal("expected an error for an unset Router.Graph, got nil")
	}
}

func TestRouterQueryFansOutToAllThreeLayersAndMergesByKindWhenUnrestricted(t *testing.T) {
	okf := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "fait"}, Similarity: 1}}}
	vec := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "souvenir haut"}, Similarity: 0.9},
		{Memory: Memory{Text: "souvenir bas"}, Similarity: 0.4}}}
	graph := &fakeStore{recalls: []Recall{{Memory: Memory{Relation: "cites"}, Similarity: 0.6}}}
	r := &Router{OKF: okf, Vector: vec, Graph: graph}

	got, err := r.Query(context.Background(), Query{Text: "x"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d recalls, want 4 (merged from all three layers)", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Similarity < got[i].Similarity {
			t.Errorf("recalls not sorted by similarity desc: %v", got)
		}
	}
}

func TestRouterQueryFanOutStaysTwoLayerWhenGraphIsUnset(t *testing.T) {
	okf := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "fait"}, Similarity: 1}}}
	vec := &fakeStore{recalls: []Recall{{Memory: Memory{Text: "souvenir"}, Similarity: 0.5}}}
	r := &Router{OKF: okf, Vector: vec}

	got, err := r.Query(context.Background(), Query{Text: "x"})
	if err != nil {
		t.Fatalf("Query: %v (unset Router.Graph must not break the empty-Kind fan-out)", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d recalls, want 2 (Graph unset, backward compatible with the two-layer fan-out)", len(got))
	}
}

func TestRouterQueryRejectsAnUnknownKind(t *testing.T) {
	r := &Router{OKF: &fakeStore{}, Vector: &fakeStore{}}

	if _, err := r.Query(context.Background(), Query{Kind: "n_importe_quoi"}); err == nil {
		t.Fatal("expected an error for an unknown kind")
	}
}

func TestRouterPropagatesTheUnderlyingStoreError(t *testing.T) {
	boom := errors.New("boom")
	r := &Router{OKF: &fakeStore{err: boom}, Vector: &fakeStore{}}

	if _, err := r.Write(context.Background(), Memory{Kind: KindOKF}); !errors.Is(err, boom) {
		t.Errorf("Write err = %v, want to wrap %v", err, boom)
	}
}
