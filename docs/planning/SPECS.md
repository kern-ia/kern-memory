---
type: Technical Specification
title: "kern-memory phase 2 — Technical Specs"
description: "The graph/relationship layer's storage schema and interface contract, on top of the unchanged phase 1 stack."
tags: [planning, specs]
timestamp: 2026-08-10T00:00:00Z
status: final
---

# kern-memory phase 2 — Technical Specs

## Stack
Unchanged from phase 1: Go 1.26.5, `github.com/yoann/kern-memory`, no framework
(`net/http` stdlib), same single binary as the document store
([decision](/specs/01-language-runtime.md), [decision](/specs/02-framework.md),
[decision](/specs/07-hosting-deployment.md)).

## Architecture
A third `memory.Store` implementation (`internal/memory/graph`), composed into the
existing `Router` alongside `.okf` and vector under a new `KindGraph` — the same
"routage par type de requête" pattern phase 1 established, not a parallel system.

## Data model & storage
A single flat SQLite table, `graph_edges`:

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PRIMARY KEY | |
| `from_kind` | TEXT | `okf` or `vector` — disambiguates the ID namespace |
| `from_id` | TEXT | |
| `to_kind` | TEXT | |
| `to_id` | TEXT | |
| `relation` | TEXT | short label, not prose (see Security) |
| `created_at` | TEXT | RFC3339Nano, same convention as `.okf` |

Edges are directed; a caller needing bidirectionality writes two edges. Multi-hop
traversal is a `WITH RECURSIVE` SQL query against this table, not an in-Go graph walk.
Edges reference existing `.okf`/vector memory IDs and never duplicate their content
([decision](/specs/03-graph-storage-schema.md), inherited from
[scope decision 05](/scope/05-graph-data-model.md)).

## Auth
Unchanged: single bearer token (`KERN_MEMORY_TOKEN`), constant-time compare, same
`memoryServer.auth` middleware ([decision](/specs/04-auth.md)).

## Interfaces & integrations
`memory.Memory` and `memory.Query` gain optional fields — `FromKind`, `FromID`,
`ToKind`, `ToID`, `Relation` on `Memory`; `Depth` on `Query` — used only when
`Kind == KindGraph`. `FromKind`/`ToKind` carry the referenced memory's own layer
(`okf` or `vector`), matching the storage schema's `(kind, id)` composite reference
([correction](/specs/05-graph-interface-contract.md), found during issue #7's
implementation — the original recommendation omitted them). No new HTTP routes: the
existing `POST /api/v1/memory/write` and `/query` carry graph edges and traversals the
same way they carry `.okf` facts and vector text today
([decision](/specs/05-graph-interface-contract.md)). No new external integration
([decision](/specs/06-external-integrations.md)).

`Query.Depth` is clamped server-side to a hard maximum regardless of what the caller
requests, the same clamping precedent already established for `chromem-go`'s `Limit`
([decision](/specs/10-traversal-depth-limit.md)).

## Deployment & operations
Unchanged: same process, same deployment shape as phase 1
([decision](/specs/07-hosting-deployment.md)).

## Testing infrastructure
Unchanged: `go test -race ./...`, real-backend integration tests (real SQLite, no
mocked embeddings where the vector layer is touched), the method every prior feature
in this repo has held to ([decision](/specs/08-testing-infrastructure.md)).

## Cross-cutting concerns

**Error handling & observability** — unchanged `fmt.Errorf` wrapping; no observability
layer (`kern-obs` stays deferred, [scope non-goal](/scope/08-non-goals.md))
([decision](/specs/09-error-handling-observability.md)).

**Security & privacy** — `Relation` carries no pseudonymization pass; it's validated as
a constrained label (length-capped, rejected if shaped like prose) rather than run
through `kern-anon`, which is built for scanning free text, not enforcing a field
shape. The memories an edge references are already pseudonymized by their own layer at
write time ([decision](/specs/11-security-privacy.md)).

**Configuration & secrets** — new `KERN_MEMORY_GRAPH_DB` env var, defaulting to
`kern-memory-graph.db`, following the exact pattern already set by `EnvOKFDB` /
`EnvVectorDB` ([decision](/specs/12-configuration-secrets.md)).

**Background work** — none; writes and queries stay synchronous, same as both existing
layers ([decision](/specs/13-background-work.md)).

**Data migrations** — none; additive `CREATE TABLE IF NOT EXISTS` at `Open`, same
pattern as the `.okf` store, no production graph data exists yet
([decision](/specs/14-data-migrations.md)).
