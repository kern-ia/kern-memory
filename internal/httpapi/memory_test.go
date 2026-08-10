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
	written []memory.Memory
	recalls []memory.Recall
	err     error
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

func (f *fakeMemoryStore) Query(_ context.Context, _ memory.Query) ([]memory.Recall, error) {
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
