package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yoann/kern-memory/internal/store"
)

func TestCombinedRouterServesDocumentEndpoints(t *testing.T) {
	docs := &fakeStore{docs: map[string]store.Document{}}
	mem := &fakeMemoryStore{}
	h := NewCombinedRouter(docs, mem, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestCombinedRouterServesMemoryEndpoints(t *testing.T) {
	docs := &fakeStore{docs: map[string]store.Document{}}
	mem := &fakeMemoryStore{recalls: nil}
	h := NewCombinedRouter(docs, mem, "")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/query", strings.NewReader(`{"text":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
}

func TestCombinedRouterServesHealthz(t *testing.T) {
	docs := &fakeStore{docs: map[string]store.Document{}}
	mem := &fakeMemoryStore{}
	h := NewCombinedRouter(docs, mem, "")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
