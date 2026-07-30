package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yoann/kern-memory/internal/store"
)

type fakeStore struct {
	docs map[string]store.Document
}

func (f *fakeStore) List(ctx context.Context) ([]store.Summary, error) {
	var out []store.Summary
	for _, d := range f.docs {
		out = append(out, store.Summary{ID: d.ID, Title: d.Title, WordCount: d.WordCount, UpdatedAt: d.UpdatedAt})
	}
	return out, nil
}

func (f *fakeStore) Get(ctx context.Context, id string) (store.Document, error) {
	d, ok := f.docs[id]
	if !ok {
		return store.Document{}, store.ErrUnknownDocument
	}
	return d, nil
}

func (f *fakeStore) Resolve(ctx context.Context, docID, suggestionID, status string) error {
	d, ok := f.docs[docID]
	if !ok {
		return store.ErrUnknownDocument
	}
	for i, sg := range d.Suggestions {
		if sg.ID == suggestionID {
			d.Suggestions[i].Status = status
			f.docs[docID] = d
			return nil
		}
	}
	return store.ErrUnknownSuggestion
}

func sample() *fakeStore {
	return &fakeStore{docs: map[string]store.Document{
		"doc1": {
			ID: "doc1", Title: "Compte-rendu", Body: "un deux trois", WordCount: 3,
			UpdatedAt:   time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
			Suggestions: []store.Suggestion{{ID: "s1", DocumentID: "doc1", Text: "x", Status: store.StatusPending}},
		},
	}}
}

func TestListRequiresTheTokenWhenOneIsConfigured(t *testing.T) {
	r := NewRouter(sample(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestListReturnsSummariesWithTheToken(t *testing.T) {
	r := NewRouter(sample(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var got []summaryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].ID != "doc1" || got[0].WordCount != 3 {
		t.Fatalf("got %+v", got)
	}
}

func TestGetReturnsTheDocumentWithItsSuggestions(t *testing.T) {
	r := NewRouter(sample(), "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/doc1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got documentDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Title != "Compte-rendu" || len(got.Suggestions) != 1 || got.Suggestions[0].ID != "s1" {
		t.Fatalf("got %+v", got)
	}
}

func TestGetUnknownDocumentIs404(t *testing.T) {
	r := NewRouter(sample(), "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/nope", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestAcceptResolvesTheSuggestion(t *testing.T) {
	s := sample()
	r := NewRouter(s, "")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc1/suggestions/s1/accept", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if s.docs["doc1"].Suggestions[0].Status != store.StatusAccepted {
		t.Fatalf("status = %q, want accepted", s.docs["doc1"].Suggestions[0].Status)
	}
}

func TestIgnoreUnknownSuggestionIs404(t *testing.T) {
	r := NewRouter(sample(), "")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc1/suggestions/nope/ignore", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
