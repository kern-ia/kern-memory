package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yoann/kern-memory/internal/memory"
)

var errBoom = errors.New("boom")

type fakeMemoryStore struct {
	written   []memory.Memory
	recalls   []memory.Recall
	err       error
	lastQuery memory.Query
}

func (f *fakeMemoryStore) Write(_ context.Context, m memory.Memory) (memory.Memory, error) {
	if f.err != nil {
		return memory.Memory{}, f.err
	}
	if m.ID == "" {
		m.ID = "generated-id"
	}
	f.written = append(f.written, m)
	return m, nil
}

func (f *fakeMemoryStore) Query(_ context.Context, q memory.Query) ([]memory.Recall, error) {
	f.lastQuery = q
	if f.err != nil {
		return nil, f.err
	}
	return f.recalls, nil
}

func TestHandleMemoryWriteStoresAndReturnsTheMemory(t *testing.T) {
	mem := &fakeMemoryStore{}
	body, _ := json.Marshal(map[string]any{"kind": "okf", "text": "le taux d'usure change chaque trimestre", "tags": []string{"usure"}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/write", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(mem.written) != 1 || mem.written[0].Text != "le taux d'usure change chaque trimestre" {
		t.Errorf("written = %v", mem.written)
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["id"] != "generated-id" {
		t.Errorf("response id = %v, want generated-id", resp["id"])
	}
}

func TestHandleMemoryWriteRejectsEmptyText(t *testing.T) {
	mem := &fakeMemoryStore{}
	body, _ := json.Marshal(map[string]any{"kind": "okf", "text": ""})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/write", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if len(mem.written) != 0 {
		t.Error("must not delegate an empty write")
	}
}

func TestHandleMemoryQueryReturnsRecalls(t *testing.T) {
	mem := &fakeMemoryStore{recalls: []memory.Recall{
		{Memory: memory.Memory{ID: "a", Text: "critère banque"}, Similarity: 0.9},
	}}
	body, _ := json.Marshal(map[string]any{"text": "quelle banque accepte une SCI"})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/query", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("got %d recalls, want 1", len(resp))
	}
}

func TestHandleMemoryEndpointsRequireAuthWhenATokenIsConfigured(t *testing.T) {
	mem := &fakeMemoryStore{}
	body, _ := json.Marshal(map[string]any{"kind": "okf", "text": "x"})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/write", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "secret").ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 without a bearer token", rec.Code)
	}
}

func TestHandleMemoryWriteThreadsGraphEdgeFields(t *testing.T) {
	mem := &fakeMemoryStore{}
	body, _ := json.Marshal(map[string]any{
		"kind":      "graph",
		"text":      "edge",
		"from_kind": "okf",
		"from_id":   "fact-1",
		"to_kind":   "vector",
		"to_id":     "vec-1",
		"relation":  "supports",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/write", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(mem.written) != 1 {
		t.Fatalf("written = %v, want 1 entry", mem.written)
	}
	got := mem.written[0]
	if got.FromKind != "okf" || got.FromID != "fact-1" || got.ToKind != "vector" || got.ToID != "vec-1" || got.Relation != "supports" {
		t.Errorf("written edge = %+v, want from_kind=okf from_id=fact-1 to_kind=vector to_id=vec-1 relation=supports", got)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["from_kind"] != "okf" || resp["from_id"] != "fact-1" || resp["to_kind"] != "vector" || resp["to_id"] != "vec-1" || resp["relation"] != "supports" {
		t.Errorf("response = %v, want edge fields echoed back", resp)
	}
}

func TestHandleMemoryWriteAllowsAGraphEdgeWithNoText(t *testing.T) {
	mem := &fakeMemoryStore{}
	body, _ := json.Marshal(map[string]any{
		"kind":      "graph",
		"from_kind": "okf",
		"from_id":   "fact-1",
		"to_kind":   "vector",
		"to_id":     "vec-1",
		"relation":  "supports",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/write", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s, want 200 — a graph edge has no text of its own", rec.Code, rec.Body.String())
	}
	if len(mem.written) != 1 {
		t.Fatalf("written = %v, want 1 entry", mem.written)
	}
}

func TestHandleMemoryQueryThreadsGraphTraversalFields(t *testing.T) {
	mem := &fakeMemoryStore{recalls: []memory.Recall{
		{Memory: memory.Memory{ID: "e1", Kind: memory.KindGraph, FromKind: "okf", FromID: "fact-1", ToKind: "vector", ToID: "vec-1", Relation: "supports"}, Similarity: 1},
	}}
	body, _ := json.Marshal(map[string]any{
		"kind":      "graph",
		"from_kind": "okf",
		"from_id":   "fact-1",
		"depth":     2,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/query", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if mem.lastQuery.FromKind != "okf" || mem.lastQuery.FromID != "fact-1" || mem.lastQuery.Depth != 2 {
		t.Errorf("query sent to store = %+v, want from_kind=okf from_id=fact-1 depth=2", mem.lastQuery)
	}

	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("got %d recalls, want 1", len(resp))
	}
	m := resp[0]["memory"].(map[string]any)
	if m["from_kind"] != "okf" || m["to_id"] != "vec-1" || m["relation"] != "supports" {
		t.Errorf("recall memory = %v, want edge fields in response", m)
	}
}

// Resolving known ids (decision 16) travels the same query endpoint.
func TestHandleMemoryQueryThreadsIDs(t *testing.T) {
	mem := &fakeMemoryStore{recalls: []memory.Recall{
		{Memory: memory.Memory{ID: "a", Kind: memory.KindOKF, Text: "x"}, Similarity: 1},
	}}
	body, _ := json.Marshal(map[string]any{
		"kind": "okf",
		"ids":  []string{"a", "b"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/query", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(mem.lastQuery.IDs) != 2 || mem.lastQuery.IDs[0] != "a" || mem.lastQuery.IDs[1] != "b" {
		t.Errorf("query sent to store = %+v, want ids=[a b]", mem.lastQuery)
	}
}

func TestHandleMemoryWritePropagatesAStoreError(t *testing.T) {
	mem := &fakeMemoryStore{err: errBoom}
	body, _ := json.Marshal(map[string]any{"kind": "okf", "text": "x"})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/write", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	NewMemoryRouter(mem, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}
