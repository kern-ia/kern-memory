// Package memory is kern-memory's neutral contract (EPIC-13 phase 1, see
// ../../docs/kern-memory-etat-de-lart.md): write(memoire)/query(contexte), routed by kind
// to one of two backends composed here, not reimplemented. Router owns no storage itself —
// it is pure routing logic, deliberately kept separate from the SQLite/chromem-go backends
// so it stays trivially testable with fakes.
package memory

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// Kind selects which layer a Memory belongs to, or a Query should hit.
type Kind string

const (
	// KindOKF is declarative, versionable memory (a fact, a rule) — exact/tag lookup, no
	// embedding. The layer's own package (internal/memory/okf) is the differentiator noted
	// in the état-de-l'art doc: relisible, auditable, diff-able, unlike a vector's opaque
	// float array.
	KindOKF Kind = "okf"
	// KindVector is free-text memory recalled by semantic similarity.
	KindVector Kind = "vector"
)

// Memory is one unit of stored memory, either layer.
type Memory struct {
	ID        string
	Kind      Kind
	Text      string
	Tags      []string
	Metadata  map[string]string
	CreatedAt time.Time
}

// Query asks for recall. An empty Kind fans out to every layer (Router.Query merges and
// sorts by Similarity); a set Kind restricts to one layer.
type Query struct {
	Text  string
	Kind  Kind
	Tags  []string
	Limit int
}

// Recall is one Query result. Similarity is the vector layer's cosine similarity ([-1,1]);
// the OKF layer, having no embedding, reports 1 for a lookup match — comparable enough for
// Router's merge-and-sort, not a claim of semantic equivalence.
type Recall struct {
	Memory     Memory
	Similarity float32
}

// Store is what each layer (okf, vector) implements — and what any transverse decorator
// (e.g. a future pseudonymizing wrapper around kern-anon) can wrap without knowing which
// concrete layer sits behind it.
type Store interface {
	Write(ctx context.Context, m Memory) (Memory, error)
	Query(ctx context.Context, q Query) ([]Recall, error)
}

// Router composes the two phase-1 layers and routes by Kind — the "routage par type de
// requête" from the ROADMAP entry. It holds no state of its own.
type Router struct {
	OKF    Store
	Vector Store
}

// Write routes m to the layer named by m.Kind; an empty Kind defaults to the vector layer
// (semantic recall is the common case — most memory has no natural tag/lookup key).
func (r *Router) Write(ctx context.Context, m Memory) (Memory, error) {
	switch m.Kind {
	case KindOKF:
		out, err := r.OKF.Write(ctx, m)
		if err != nil {
			return Memory{}, fmt.Errorf("memory: okf write: %w", err)
		}
		return out, nil
	case KindVector, "":
		out, err := r.Vector.Write(ctx, m)
		if err != nil {
			return Memory{}, fmt.Errorf("memory: vector write: %w", err)
		}
		return out, nil
	default:
		return Memory{}, fmt.Errorf("memory: unknown kind %q", m.Kind)
	}
}

// Query routes by q.Kind; an empty Kind fans out to both layers and merges the results,
// most similar first — the caller doesn't have to know which layer answered.
func (r *Router) Query(ctx context.Context, q Query) ([]Recall, error) {
	switch q.Kind {
	case KindOKF:
		out, err := r.OKF.Query(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("memory: okf query: %w", err)
		}
		return out, nil
	case KindVector:
		out, err := r.Vector.Query(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("memory: vector query: %w", err)
		}
		return out, nil
	case "":
		okf, err := r.OKF.Query(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("memory: okf query: %w", err)
		}
		vec, err := r.Vector.Query(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("memory: vector query: %w", err)
		}
		merged := append(okf, vec...)
		sort.SliceStable(merged, func(i, j int) bool { return merged[i].Similarity > merged[j].Similarity })
		return merged, nil
	default:
		return nil, fmt.Errorf("memory: unknown kind %q", q.Kind)
	}
}
