package main

import (
	"context"
	"errors"
	"testing"

	"github.com/yoann/kern-memory/internal/memory"
	"github.com/yoann/kern-memory/internal/memory/graph"
	"github.com/yoann/kern-memory/internal/memory/okf"
)

var errBoom = errors.New("boom")

// fakeStore is a minimal in-memory memory.Store for testing loadFile.writeAll's dispatch
// logic in isolation from any real backend — mirrors internal/memory's own fakeStore
// (unexported there, so re-declared here for this package's tests).
type fakeStore struct {
	written []memory.Memory
	err     error
}

func (f *fakeStore) Write(_ context.Context, m memory.Memory) (memory.Memory, error) {
	if f.err != nil {
		return memory.Memory{}, f.err
	}
	f.written = append(f.written, m)
	return m, nil
}

func (f *fakeStore) Query(_ context.Context, _ memory.Query) ([]memory.Recall, error) {
	return nil, nil
}

const exampleLoadFile = `{
  "memories": [
    {"id": "criterion-sci-senior", "kind": "vector", "text": "la banque X accepte les SCI pour les seniors", "tags": ["bank-criteria"], "metadata": {"bank": "X"}},
    {"id": "criterion-a", "kind": "okf", "text": "règle A", "tags": ["bank-criteria"]}
  ],
  "edges": [
    {"from_kind": "vector", "from_id": "criterion-sci-senior", "to_kind": "okf", "to_id": "criterion-a", "relation": "supersedes"}
  ]
}`

func TestParseLoadFileParsesMemoriesAndEdges(t *testing.T) {
	f, err := parseLoadFile([]byte(exampleLoadFile))
	if err != nil {
		t.Fatalf("parseLoadFile: %v", err)
	}
	if len(f.Memories) != 2 {
		t.Fatalf("got %d memories, want 2", len(f.Memories))
	}
	if f.Memories[0].ID != "criterion-sci-senior" || f.Memories[0].Kind != memory.KindVector {
		t.Errorf("memories[0] = %+v, want id=criterion-sci-senior kind=vector", f.Memories[0])
	}
	if f.Memories[1].Kind != memory.KindOKF {
		t.Errorf("memories[1].Kind = %q, want okf", f.Memories[1].Kind)
	}
	if len(f.Edges) != 1 {
		t.Fatalf("got %d edges, want 1", len(f.Edges))
	}
	e := f.Edges[0]
	if e.FromKind != "vector" || e.FromID != "criterion-sci-senior" || e.ToKind != "okf" || e.ToID != "criterion-a" || e.Relation != "supersedes" {
		t.Errorf("edges[0] = %+v, want the parsed edge fields", e)
	}
}

func TestParseLoadFileRejectsMalformedJSONWithAClearError(t *testing.T) {
	_, err := parseLoadFile([]byte(`{"memories": [`))
	if err == nil {
		t.Fatal("expected an error for malformed JSON, got nil")
	}
}

func TestWriteAllWritesEachMemoryWithItsKind(t *testing.T) {
	f, err := parseLoadFile([]byte(exampleLoadFile))
	if err != nil {
		t.Fatalf("parseLoadFile: %v", err)
	}
	okf, vec, graph := &fakeStore{}, &fakeStore{}, &fakeStore{}
	r := &memory.Router{OKF: okf, Vector: vec, Graph: graph}

	if err := f.writeAll(context.Background(), r); err != nil {
		t.Fatalf("writeAll: %v", err)
	}

	if len(vec.written) != 1 || vec.written[0].ID != "criterion-sci-senior" {
		t.Errorf("vec.written = %+v, want the vector-kind memory routed there", vec.written)
	}
	if len(okf.written) != 1 || okf.written[0].ID != "criterion-a" {
		t.Errorf("okf.written = %+v, want the okf-kind memory routed there", okf.written)
	}
}

func TestWriteAllWritesEdgesAsGraphKindWithFieldsIntact(t *testing.T) {
	f, err := parseLoadFile([]byte(exampleLoadFile))
	if err != nil {
		t.Fatalf("parseLoadFile: %v", err)
	}
	okf, vec, graph := &fakeStore{}, &fakeStore{}, &fakeStore{}
	r := &memory.Router{OKF: okf, Vector: vec, Graph: graph}

	if err := f.writeAll(context.Background(), r); err != nil {
		t.Fatalf("writeAll: %v", err)
	}

	if len(graph.written) != 1 {
		t.Fatalf("graph.written = %d, want 1", len(graph.written))
	}
	got := graph.written[0]
	if got.Kind != memory.KindGraph {
		t.Errorf("edge Kind = %q, want graph", got.Kind)
	}
	if got.FromKind != "vector" || got.FromID != "criterion-sci-senior" || got.ToKind != "okf" || got.ToID != "criterion-a" || got.Relation != "supersedes" {
		t.Errorf("graph.written[0] = %+v, want the edge's fields carried through", got)
	}
}

// TestLoadFileRoundTripsThroughRealOKFAndGraphStores is the acceptance criterion's own
// scenario ("a file containing 2 memories + 1 edge writes all three, verified by querying
// them back"), run against the real okf and graph SQLite layers (pure Go, no external
// service) — the vector layer is faked here only to keep this test independent of a running
// Ollama; real vector-layer behavior is covered directly in internal/memory/vector.
func TestLoadFileRoundTripsThroughRealOKFAndGraphStores(t *testing.T) {
	okfStore, err := okf.Open(t.TempDir() + "/okf.db")
	if err != nil {
		t.Fatalf("okf.Open: %v", err)
	}
	defer okfStore.Close()
	graphStore, err := graph.Open(t.TempDir() + "/graph.db")
	if err != nil {
		t.Fatalf("graph.Open: %v", err)
	}
	defer graphStore.Close()
	r := &memory.Router{OKF: okfStore, Vector: &fakeStore{}, Graph: graphStore}

	const file = `{
		"memories": [
			{"id": "criterion-a", "kind": "okf", "text": "règle A", "tags": ["bank-criteria"]},
			{"id": "criterion-b", "kind": "okf", "text": "règle B", "tags": ["bank-criteria"]}
		],
		"edges": [
			{"from_kind": "okf", "from_id": "criterion-a", "to_kind": "okf", "to_id": "criterion-b", "relation": "supersedes"}
		]
	}`

	f, err := parseLoadFile([]byte(file))
	if err != nil {
		t.Fatalf("parseLoadFile: %v", err)
	}
	if err := f.writeAll(context.Background(), r); err != nil {
		t.Fatalf("writeAll: %v", err)
	}

	memories, err := r.Query(context.Background(), memory.Query{Kind: memory.KindOKF})
	if err != nil {
		t.Fatalf("Query okf: %v", err)
	}
	if len(memories) != 2 {
		t.Fatalf("got %d okf memories back, want 2", len(memories))
	}

	edges, err := r.Query(context.Background(), memory.Query{Kind: memory.KindGraph, FromKind: "okf", FromID: "criterion-a"})
	if err != nil {
		t.Fatalf("Query graph: %v", err)
	}
	if len(edges) != 1 || edges[0].Memory.ToID != "criterion-b" || edges[0].Memory.Relation != "supersedes" {
		t.Errorf("got %+v, want one edge criterion-a->criterion-b (supersedes)", edges)
	}
}

// TestLoadFileRoundTripDoesNotDuplicateOKFMemoriesOnARerun is the acceptance criterion's
// upsert requirement for the layer that actually supports it (okf) — re-running the loader
// on an unchanged file must not create duplicate rows.
func TestLoadFileRoundTripDoesNotDuplicateOKFMemoriesOnARerun(t *testing.T) {
	okfStore, err := okf.Open(t.TempDir() + "/okf.db")
	if err != nil {
		t.Fatalf("okf.Open: %v", err)
	}
	defer okfStore.Close()
	r := &memory.Router{OKF: okfStore, Vector: &fakeStore{}}

	const file = `{"memories": [{"id": "criterion-a", "kind": "okf", "text": "règle A"}]}`
	f, err := parseLoadFile([]byte(file))
	if err != nil {
		t.Fatalf("parseLoadFile: %v", err)
	}

	if err := f.writeAll(context.Background(), r); err != nil {
		t.Fatalf("first writeAll: %v", err)
	}
	if err := f.writeAll(context.Background(), r); err != nil {
		t.Fatalf("second writeAll (rerun): %v", err)
	}

	got, err := r.Query(context.Background(), memory.Query{Kind: memory.KindOKF})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d okf memories after two identical loads, want 1 (upsert, no duplicate)", len(got))
	}
}

func TestWriteAllPropagatesTheUnderlyingStoreError(t *testing.T) {
	f, err := parseLoadFile([]byte(exampleLoadFile))
	if err != nil {
		t.Fatalf("parseLoadFile: %v", err)
	}
	r := &memory.Router{OKF: &fakeStore{err: errBoom}, Vector: &fakeStore{}, Graph: &fakeStore{}}

	if err := f.writeAll(context.Background(), r); err == nil {
		t.Fatal("expected an error to propagate from the underlying store, got nil")
	}
}
