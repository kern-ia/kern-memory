// Package vector is the memory contract's semantic layer (memory.KindVector), backed by
// chromem-go (embedded, zero external service — see ../../../docs/kern-memory-etat-de-lart.md,
// "colle au single-binary") with embeddings from a local Ollama model, souveraineté-first:
// no embedding call ever leaves the machine.
package vector

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	chromem "github.com/philippgille/chromem-go"

	"github.com/yoann/kern-memory/internal/memory"
)

const collectionName = "memory"

// Store is a memory.Store backed by a persistent chromem-go collection.
type Store struct {
	collection *chromem.Collection
}

// Open opens (creating if needed) a persistent chromem-go DB at path, using model via
// Ollama's embedding API. baseURL empty defaults to Ollama's standard local address
// ("http://localhost:11434/api", chromem-go's own default).
func Open(path, model, baseURL string) (*Store, error) {
	db, err := chromem.NewPersistentDB(path, false)
	if err != nil {
		return nil, fmt.Errorf("vector: open %q: %w", path, err)
	}
	embed := chromem.NewEmbeddingFuncOllama(model, baseURL)
	coll, err := db.GetOrCreateCollection(collectionName, nil, embed)
	if err != nil {
		return nil, fmt.Errorf("vector: collection: %w", err)
	}
	return &Store{collection: coll}, nil
}

// Close is a no-op: chromem-go's persistent DB writes synchronously on every call, there
// is nothing left to flush. Kept for symmetry with the other layer's Store (okf.Store),
// both used the same way (defer s.Close()) by callers that don't need to know which.
func (s *Store) Close() error { return nil }

// Write embeds m.Text (a real call to the configured Ollama model) and stores it.
func (s *Store) Write(ctx context.Context, m memory.Memory) (memory.Memory, error) {
	if m.ID == "" {
		m.ID = newID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	meta, err := encodeMetadata(m)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("vector: encode metadata: %w", err)
	}
	if err := s.collection.AddDocument(ctx, chromem.Document{
		ID:       m.ID,
		Metadata: meta,
		Content:  m.Text,
	}); err != nil {
		return memory.Memory{}, fmt.Errorf("vector: write: %w", err)
	}
	m.Kind = memory.KindVector
	return m, nil
}

// Query embeds q.Text and returns the nearest neighbors by cosine similarity. Limit <= 0
// defaults to 5; Limit is clamped to the collection's size — chromem-go's Query requires
// nResults > 0 and errors if it exceeds the document count.
func (s *Store) Query(ctx context.Context, q memory.Query) ([]memory.Recall, error) {
	n := q.Limit
	if n <= 0 {
		n = 5
	}
	if count := s.collection.Count(); n > count {
		n = count
	}
	if n == 0 {
		return nil, nil
	}

	results, err := s.collection.Query(ctx, q.Text, n, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("vector: query: %w", err)
	}

	out := make([]memory.Recall, 0, len(results))
	for _, r := range results {
		m, err := decodeMetadata(r.ID, r.Content, r.Metadata)
		if err != nil {
			return nil, fmt.Errorf("vector: decode metadata: %w", err)
		}
		out = append(out, memory.Recall{Memory: m, Similarity: r.Similarity})
	}
	return out, nil
}

// chromem-go's Metadata is map[string]string only — Tags/Metadata/CreatedAt are folded
// into single JSON-encoded fields rather than widening the schema per-caller.
func encodeMetadata(m memory.Memory) (map[string]string, error) {
	tagsJSON, err := json.Marshal(m.Tags)
	if err != nil {
		return nil, err
	}
	metaJSON, err := json.Marshal(m.Metadata)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"tags":       string(tagsJSON),
		"metadata":   string(metaJSON),
		"created_at": m.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
	}, nil
}

func decodeMetadata(id, content string, meta map[string]string) (memory.Memory, error) {
	m := memory.Memory{ID: id, Text: content, Kind: memory.KindVector}
	if raw, ok := meta["tags"]; ok && raw != "" {
		if err := json.Unmarshal([]byte(raw), &m.Tags); err != nil {
			return memory.Memory{}, err
		}
	}
	if raw, ok := meta["metadata"]; ok && raw != "" {
		if err := json.Unmarshal([]byte(raw), &m.Metadata); err != nil {
			return memory.Memory{}, err
		}
	}
	if raw, ok := meta["created_at"]; ok && raw != "" {
		m.CreatedAt, _ = time.Parse("2006-01-02T15:04:05.999999999Z07:00", raw)
	}
	return m, nil
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
