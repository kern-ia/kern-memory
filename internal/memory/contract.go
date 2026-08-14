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
	// KindGraph is a relationship (edge) between two existing memories — no text of its
	// own, just FromID/ToID/Relation. See docs/planning/specs/05-graph-interface-contract.md.
	KindGraph Kind = "graph"
)

// Memory is one unit of stored memory, either layer. FromID, FromKind, ToID, ToKind and
// Relation are only meaningful when Kind == KindGraph — they express one edge between two
// existing memories (referenced by their .okf/vector IDs, disambiguated by which layer each
// one lives in — an id has no shared namespace across layers), not a new content-bearing
// memory.
type Memory struct {
	ID        string
	Kind      Kind
	Text      string
	Tags      []string
	Metadata  map[string]string
	CreatedAt time.Time
	FromID    string
	FromKind  string
	ToID      string
	ToKind    string
	Relation  string
}

// Query asks for recall. An empty Kind fans out to every layer (Router.Query merges and
// sorts by Similarity); a set Kind restricts to one layer. FromKind, FromID and Depth are
// only meaningful when Kind == KindGraph: FromKind/FromID identify the traversal's
// starting node (the same (kind, id) composite Memory uses to reference a memory across
// layers — see contract 05), and Depth bounds how many hops the traversal walks from
// there (the graph layer clamps it to a hard server-side maximum regardless of what the
// caller requests).
//
// IDs, meaningful for KindOKF and KindVector (decision 16), means "return exactly these
// memories" instead of a tag/text search — mutually exclusive with Tags/Text in practice,
// since a caller resolving known ids has nothing to search for. Ignored by the graph
// layer: an edge is never looked up by id this way, only the memories it references are.
type Query struct {
	Text     string
	Kind     Kind
	Tags     []string
	IDs      []string
	Limit    int
	FromKind string
	FromID   string
	Depth    int
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

// Router composes the phase-1 layers (plus the graph layer added in Epic 1) and routes by
// Kind — the "routage par type de requête" from the ROADMAP entry. It holds no state of its
// own. Graph is optional: callers that haven't wired a graph store yet (e.g. main.go before
// Epic 1's storage lands) leave it nil, and Router keeps behaving exactly as it did before
// KindGraph existed — see the nil checks below.
type Router struct {
	OKF    Store
	Vector Store
	Graph  Store
}

// Write routes m to the layer named by m.Kind; an empty Kind defaults to the vector layer
// (semantic recall is the common case — most memory has no natural tag/lookup key). Writing
// a KindGraph edge with no Graph store configured is a caller/deployment error, not a
// nil-pointer panic — it's reported the same way an unknown Kind is.
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
	case KindGraph:
		if r.Graph == nil {
			return Memory{}, fmt.Errorf("memory: graph write: no graph store configured")
		}
		out, err := r.Graph.Write(ctx, m)
		if err != nil {
			return Memory{}, fmt.Errorf("memory: graph write: %w", err)
		}
		return out, nil
	default:
		return Memory{}, fmt.Errorf("memory: unknown kind %q", m.Kind)
	}
}

// Query routes by q.Kind; an empty Kind fans out to every configured layer and merges the
// results, most similar first — the caller doesn't have to know which layer answered. The
// graph layer only joins that fan-out once Router.Graph is set (see the Router doc comment);
// until then the fan-out stays the original OKF+Vector behavior. Querying KindGraph directly
// with no Graph store configured is reported as a clear error, not a nil-pointer panic.
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
	case KindGraph:
		if r.Graph == nil {
			return nil, fmt.Errorf("memory: graph query: no graph store configured")
		}
		out, err := r.Graph.Query(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("memory: graph query: %w", err)
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
		if r.Graph != nil {
			graph, err := r.Graph.Query(ctx, q)
			if err != nil {
				return nil, fmt.Errorf("memory: graph query: %w", err)
			}
			merged = append(merged, graph...)
		}
		sort.SliceStable(merged, func(i, j int) bool { return merged[i].Similarity > merged[j].Similarity })
		return merged, nil
	default:
		return nil, fmt.Errorf("memory: unknown kind %q", q.Kind)
	}
}
