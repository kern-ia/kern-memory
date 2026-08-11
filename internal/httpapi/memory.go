package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yoann/kern-memory/internal/memory"
)

// MemoryStore is what the memory endpoints need — memory.Router (or any memory.Store)
// satisfies it without this package importing memory.Router's concrete backends.
type MemoryStore interface {
	Write(ctx context.Context, m memory.Memory) (memory.Memory, error)
	Query(ctx context.Context, q memory.Query) ([]memory.Recall, error)
}

// NewMemoryRouter builds the standalone /api/v1/memory/* handler — kept separate from
// NewRouter (documents/suggestions) so each transport concern stays independently
// testable; cmd/kern-memory/main.go composes both into the one served binary (EPIC-13
// phase 1 lives in the same daemon as the C8 v1 slice, see CLAUDE.md).
func NewMemoryRouter(s MemoryStore, token string) http.Handler {
	srv := &memoryServer{store: s, token: token}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/memory/write", srv.auth(srv.handleWrite))
	mux.HandleFunc("POST /api/v1/memory/query", srv.auth(srv.handleQuery))
	return mux
}

type memoryServer struct {
	store MemoryStore
	token string
}

func (s *memoryServer) auth(next http.HandlerFunc) http.HandlerFunc {
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

type memoryWriteRequest struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"`
	Text     string            `json:"text"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	// FromKind/FromID/ToKind/ToID/Relation (Epic 1) are only meaningful when
	// Kind == "graph" — they describe one directed edge between two existing
	// memories. See memory.Memory's doc comment.
	FromKind string `json:"from_kind"`
	FromID   string `json:"from_id"`
	ToKind   string `json:"to_kind"`
	ToID     string `json:"to_id"`
	Relation string `json:"relation"`
}

type memoryDTO struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"`
	Text     string            `json:"text"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	FromKind string            `json:"from_kind,omitempty"`
	FromID   string            `json:"from_id,omitempty"`
	ToKind   string            `json:"to_kind,omitempty"`
	ToID     string            `json:"to_id,omitempty"`
	Relation string            `json:"relation,omitempty"`
}

func (s *memoryServer) handleWrite(w http.ResponseWriter, r *http.Request) {
	var req memoryWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "malformed body: want {\"kind\",\"text\",...}")
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	out, err := s.store.Write(r.Context(), memory.Memory{
		ID: req.ID, Kind: memory.Kind(req.Kind), Text: req.Text, Tags: req.Tags, Metadata: req.Metadata,
		FromKind: req.FromKind, FromID: req.FromID, ToKind: req.ToKind, ToID: req.ToID, Relation: req.Relation,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, memoryDTO{
		ID: out.ID, Kind: string(out.Kind), Text: out.Text, Tags: out.Tags, Metadata: out.Metadata,
		FromKind: out.FromKind, FromID: out.FromID, ToKind: out.ToKind, ToID: out.ToID, Relation: out.Relation,
	})
}

type memoryQueryRequest struct {
	Text  string   `json:"text"`
	Kind  string   `json:"kind"`
	Tags  []string `json:"tags"`
	Limit int      `json:"limit"`
	// FromKind/FromID/Depth (Epic 1) are only meaningful when Kind == "graph" — they
	// name the traversal's starting node and how many hops to walk. See memory.Query's
	// doc comment.
	FromKind string `json:"from_kind"`
	FromID   string `json:"from_id"`
	Depth    int    `json:"depth"`
}

type recallDTO struct {
	Memory     memoryDTO `json:"memory"`
	Similarity float32   `json:"similarity"`
}

func (s *memoryServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	var req memoryQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "malformed body: want {\"text\",...}")
		return
	}

	recalls, err := s.store.Query(r.Context(), memory.Query{
		Text: req.Text, Kind: memory.Kind(req.Kind), Tags: req.Tags, Limit: req.Limit,
		FromKind: req.FromKind, FromID: req.FromID, Depth: req.Depth,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]recallDTO, len(recalls))
	for i, rec := range recalls {
		out[i] = recallDTO{
			Memory: memoryDTO{
				ID: rec.Memory.ID, Kind: string(rec.Memory.Kind), Text: rec.Memory.Text,
				Tags: rec.Memory.Tags, Metadata: rec.Memory.Metadata,
				FromKind: rec.Memory.FromKind, FromID: rec.Memory.FromID,
				ToKind: rec.Memory.ToKind, ToID: rec.Memory.ToID, Relation: rec.Memory.Relation,
			},
			Similarity: rec.Similarity,
		}
	}
	writeJSON(w, http.StatusOK, out)
}
