// Package httpapi is kern-memory's HTTP transport: routing, authentication, and JSON in and
// out. It knows nothing about SQLite — internal/store is the Store it is handed, so this
// package is testable with a fake.
package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/yoann/kern-memory/internal/store"
)

// Store is what the router needs from persistence.
type Store interface {
	List(ctx context.Context) ([]store.Summary, error)
	Get(ctx context.Context, id string) (store.Document, error)
	Resolve(ctx context.Context, docID, suggestionID, status string) error
}

// NewRouter builds kern-memory's HTTP handler. An empty token leaves every endpoint open,
// the local-development case; the binary refuses to bind a public address without one.
func NewRouter(s Store, token string) http.Handler {
	srv := &server{store: s, token: token}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /api/v1/documents", srv.auth(srv.handleList))
	mux.HandleFunc("GET /api/v1/documents/{id}", srv.auth(srv.handleGet))
	mux.HandleFunc("POST /api/v1/documents/{id}/suggestions/{sid}/accept", srv.auth(srv.handleResolve(store.StatusAccepted)))
	mux.HandleFunc("POST /api/v1/documents/{id}/suggestions/{sid}/ignore", srv.auth(srv.handleResolve(store.StatusIgnored)))
	return mux
}

type server struct {
	store Store
	token string
}

func (s *server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next(w, r)
			return
		}
		presented := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if presented == "" || subtle.ConstantTimeCompare([]byte(s.token), []byte(presented)) != 1 {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r)
	}
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type summaryDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	WordCount int    `json:"word_count"`
	UpdatedAt string `json:"updated_at"`
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]summaryDTO, len(summaries))
	for i, sum := range summaries {
		out[i] = summaryDTO{ID: sum.ID, Title: sum.Title, WordCount: sum.WordCount, UpdatedAt: sum.UpdatedAt.Format(rfc3339)}
	}
	writeJSON(w, http.StatusOK, out)
}

type suggestionDTO struct {
	ID          string `json:"id"`
	AnchorStart int    `json:"anchor_start"`
	AnchorEnd   int    `json:"anchor_end"`
	Text        string `json:"text"`
	Status      string `json:"status"`
}

type documentDTO struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Body        string          `json:"body"`
	WordCount   int             `json:"word_count"`
	UpdatedAt   string          `json:"updated_at"`
	Suggestions []suggestionDTO `json:"suggestions"`
}

const rfc3339 = "2006-01-02T15:04:05.999999999Z07:00"

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	doc, err := s.store.Get(r.Context(), r.PathValue("id"))
	switch {
	case errors.Is(err, store.ErrUnknownDocument):
		writeError(w, http.StatusNotFound, "unknown document")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sugs := make([]suggestionDTO, len(doc.Suggestions))
	for i, sg := range doc.Suggestions {
		sugs[i] = suggestionDTO{ID: sg.ID, AnchorStart: sg.AnchorStart, AnchorEnd: sg.AnchorEnd, Text: sg.Text, Status: sg.Status}
	}
	writeJSON(w, http.StatusOK, documentDTO{
		ID: doc.ID, Title: doc.Title, Body: doc.Body, WordCount: doc.WordCount,
		UpdatedAt: doc.UpdatedAt.Format(rfc3339), Suggestions: sugs,
	})
}

func (s *server) handleResolve(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.store.Resolve(r.Context(), r.PathValue("id"), r.PathValue("sid"), status)
		switch {
		case errors.Is(err, store.ErrUnknownDocument):
			writeError(w, http.StatusNotFound, "unknown document")
		case errors.Is(err, store.ErrUnknownSuggestion):
			writeError(w, http.StatusNotFound, "unknown suggestion")
		case err != nil:
			writeError(w, http.StatusInternalServerError, err.Error())
		default:
			writeJSON(w, http.StatusOK, map[string]string{"status": status})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
