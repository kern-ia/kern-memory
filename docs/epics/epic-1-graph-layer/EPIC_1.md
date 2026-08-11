---
type: Epic
title: "Graph layer"
description: "An embedded relationship layer on kern-memory's write/query contract: write and traverse relationships between existing memories."
tags: [epic]
timestamp: 2026-08-10T10:00:00Z
epic: 1
slug: graph-layer
status: done
gh_issue: 3
milestone: 1
resource: https://github.com/kern-ia/kern-memory/issues/3
source: docs/planning/SCOPE.md#milestone-1-graph-layer
---

# Epic 1: Graph layer

## Goal
Close the biggest named gap left by EPIC-13 phase 1: `kern-memory` can answer "what
matches this tag" and "what's close in meaning to this," but not "what's connected to
what, and since when." This epic adds that third layer — a light, embedded graph on
SQLite — so a caller can write a relationship between two memories it already stored
and later recall it through single-hop or multi-hop traversal.

## Scope
- A new SQLite-backed `memory.Store` implementation (`internal/memory/graph`), composed
  into the existing `Router` under a new `Kind="graph"`.
- Storage: a single flat `graph_edges` table — `(id, from_kind, from_id, to_kind,
  to_id, relation, created_at)` — edges reference existing `.okf`/vector memory IDs and
  never duplicate their content.
- Contract: `memory.Memory` gains optional `FromID`/`ToID`/`Relation` fields, and
  `memory.Query` gains an optional `Depth` field, both active only when
  `Kind == KindGraph`. No new HTTP routes — the existing `POST /api/v1/memory/write`
  and `/query` carry graph edges and traversals the same way they already carry `.okf`
  facts and vector text.
- Multi-hop traversal via SQL (`WITH RECURSIVE`), not an in-Go graph walk.
- `Query.Depth` clamped to a hard server-side maximum regardless of what the caller
  requests.
- `Relation` validated as a constrained, short label (length-capped, rejected if
  shaped like prose) — not pseudonymized, since it carries no free text.
- New `KERN_MEMORY_GRAPH_DB` env var, following the existing `EnvOKFDB`/`EnvVectorDB`
  pattern.

## Out of scope
- A single `query` call fanning out across `.okf`, vector, *and* graph together in one
  merged response — additive polish once this epic's journey works on its own.
- Populating any real content into the graph — that's Epic 2's concern, and depends on
  this epic existing first.
- `kern-obs` (traced recall) and per-caller trust separation — project-wide non-goals
  for this phase, not specific to the graph layer.

## Acceptance criteria
- A caller writes a relationship between two existing memories (referencing their
  `.okf`/vector IDs) via `POST /api/v1/memory/write` with `Kind="graph"`.
- A caller queries that relationship back, both single-hop and multi-hop, via
  `POST /api/v1/memory/query` with `Kind="graph"`.
- Verified against a real running daemon and a real SQLite database — not mocked, the
  same bar EPIC-13 phase 1 held for its own end-to-end verification.
- `go test -race ./...` green.

## Dependencies
None.

## Context
- [Technical specs](../../planning/SPECS.md) — graph storage schema, interface
  contract, traversal depth cap, and the `Relation` validation rule this epic
  implements exactly.
- [Conventions](../../planning/CONVENTIONS.md) — repo standards (branch/commit model,
  TDD, error handling) this epic's issues follow.

## Notes
Project-wide constraint, relevant here specifically because this epic adds a new
storage layer: **sovereignty** — no external network call for anything touching data.
The graph layer must stay embedded (SQLite, in-process), matching `.okf` and vector,
not a service dependency (this ruled out adopting Graphiti/Neo4j —
[decision](../../planning/specs/03-graph-storage-schema.md)).

Risk carried from `SCOPE.md`: a light Go/SQLite graph may not match richer
temporal-graph engines if the real workload later needs multi-hop reasoning beyond
simple relationships — accepted as sufficient for today's need, revisit if proven
otherwise.
