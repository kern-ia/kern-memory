package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// ErrUnknownDocument is returned when a document id has no row.
var ErrUnknownDocument = errors.New("store: unknown document")

// ErrUnknownSuggestion is returned when a suggestion id has no row on the given document.
var ErrUnknownSuggestion = errors.New("store: unknown suggestion")

const schema = `
CREATE TABLE IF NOT EXISTS documents (
	id         TEXT PRIMARY KEY,
	title      TEXT NOT NULL,
	body       TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS suggestions (
	id           TEXT PRIMARY KEY,
	document_id  TEXT NOT NULL,
	anchor_start INTEGER NOT NULL,
	anchor_end   INTEGER NOT NULL,
	text         TEXT NOT NULL,
	status       TEXT NOT NULL
);`

// busyTimeoutMS mirrors kern-orch/internal/checkpoint: a resolve (accept/ignore) can race a
// concurrent read of the same document, learned the hard way there already.
const busyTimeoutMS = 5000

// SQLiteStore is a Store backed by modernc.org/sqlite (pure Go, no cgo).
type SQLiteStore struct {
	db *sql.DB
}

// Open opens (creating if needed) the database at path and ensures the schema exists.
func Open(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store: open %q: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d;", busyTimeoutMS)); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: set busy_timeout: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: schema: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

// Close releases the database handle.
func (s *SQLiteStore) Close() error { return s.db.Close() }

// Seed writes a new document directly — the only way content enters the store in V1.
func (s *SQLiteStore) Seed(ctx context.Context, id, title, body string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO documents (id, title, body, updated_at) VALUES (?, ?, ?, ?)`,
		id, title, body, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("store: seed: %w", err)
	}
	return nil
}

// List returns one Summary per document, most recently updated first.
func (s *SQLiteStore) List(ctx context.Context) ([]Summary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, body, updated_at FROM documents ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("store: list: %w", err)
	}
	defer rows.Close()

	var out []Summary
	for rows.Next() {
		var (
			sum        Summary
			body       string
			updatedStr string
		)
		if err := rows.Scan(&sum.ID, &sum.Title, &body, &updatedStr); err != nil {
			return nil, fmt.Errorf("store: scan summary: %w", err)
		}
		sum.WordCount = wordCount(body)
		sum.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		out = append(out, sum)
	}
	return out, rows.Err()
}

// Get returns a document with its suggestions, pending first.
func (s *SQLiteStore) Get(ctx context.Context, id string) (Document, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT title, body, updated_at FROM documents WHERE id = ?`, id)
	var (
		doc        Document
		updatedStr string
	)
	doc.ID = id
	switch err := row.Scan(&doc.Title, &doc.Body, &updatedStr); {
	case errors.Is(err, sql.ErrNoRows):
		return Document{}, ErrUnknownDocument
	case err != nil:
		return Document{}, fmt.Errorf("store: get: %w", err)
	}
	doc.WordCount = wordCount(doc.Body)
	doc.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, anchor_start, anchor_end, text, status FROM suggestions
		 WHERE document_id = ? ORDER BY anchor_start ASC`, id)
	if err != nil {
		return Document{}, fmt.Errorf("store: get: suggestions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		sg := Suggestion{DocumentID: id}
		if err := rows.Scan(&sg.ID, &sg.AnchorStart, &sg.AnchorEnd, &sg.Text, &sg.Status); err != nil {
			return Document{}, fmt.Errorf("store: scan suggestion: %w", err)
		}
		doc.Suggestions = append(doc.Suggestions, sg)
	}
	if err := rows.Err(); err != nil {
		return Document{}, err
	}
	return doc, nil
}

// SeedSuggestion writes a suggestion directly, alongside Seed — no HTTP path creates one in
// V1 either.
func (s *SQLiteStore) SeedSuggestion(ctx context.Context, sg Suggestion) error {
	if sg.Status == "" {
		sg.Status = StatusPending
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO suggestions (id, document_id, anchor_start, anchor_end, text, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		sg.ID, sg.DocumentID, sg.AnchorStart, sg.AnchorEnd, sg.Text, sg.Status)
	if err != nil {
		return fmt.Errorf("store: seed suggestion: %w", err)
	}
	return nil
}

// Resolve marks a suggestion accepted or ignored. ErrUnknownDocument if the document does
// not exist, ErrUnknownSuggestion if the suggestion does not exist on it.
func (s *SQLiteStore) Resolve(ctx context.Context, docID, suggestionID, status string) error {
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM documents WHERE id = ?)`, docID).Scan(&exists); err != nil {
		return fmt.Errorf("store: resolve: check document: %w", err)
	}
	if !exists {
		return ErrUnknownDocument
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE suggestions SET status = ? WHERE id = ? AND document_id = ?`,
		status, suggestionID, docID)
	if err != nil {
		return fmt.Errorf("store: resolve: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: resolve: rows affected: %w", err)
	}
	if n == 0 {
		return ErrUnknownSuggestion
	}
	return nil
}
