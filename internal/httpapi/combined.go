package httpapi

import (
	"net/http"
	"strings"
)

// NewCombinedRouter serves both the C8 v1 document/suggestion API and the EPIC-13 phase 1
// memory API from one process (CLAUDE.md: "ce dépôt tient la place d'une tranche de cette
// brique" — same daemon, growing set of endpoints, not a second binary). Dispatch is by
// path prefix rather than merging both muxes into one: NewRouter and NewMemoryRouter each
// own their own exact-match patterns (including NewRouter's /healthz), and neither needs
// to know the other exists.
func NewCombinedRouter(docs Store, mem MemoryStore, token string) http.Handler {
	docsHandler := NewRouter(docs, token)
	memHandler := NewMemoryRouter(mem, token)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/memory/") {
			memHandler.ServeHTTP(w, r)
			return
		}
		docsHandler.ServeHTTP(w, r)
	})
}
