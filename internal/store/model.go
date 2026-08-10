// Package store is kern-memory's persistence: documents and the suggestions anchored on
// them. It knows nothing about HTTP or tokens — internal/httpapi is the transport that
// wraps it.
package store

import (
	"strings"
	"time"
)

// Document is a piece of text kern-ui's Rédaction view can show, plus the suggestions
// anchored on it. WordCount is computed server-side so kern-ui never formats domain data.
type Document struct {
	ID          string
	Title       string
	Body        string
	WordCount   int
	UpdatedAt   time.Time
	Suggestions []Suggestion
}

// Summary is what the document list needs — no body, no suggestions.
type Summary struct {
	ID        string
	Title     string
	WordCount int
	UpdatedAt time.Time
}

// Suggestion status values. Pending is the only one a person can still act on.
const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusIgnored  = "ignored"
)

// Suggestion is anchored on a byte span of its document's Body.
type Suggestion struct {
	ID          string
	DocumentID  string
	AnchorStart int
	AnchorEnd   int
	Text        string
	Status      string
}

// wordCount counts whitespace-separated runs — good enough for plain text, and the same
// notion of "word" a person would use counting by eye.
func wordCount(body string) int {
	return len(strings.Fields(body))
}
