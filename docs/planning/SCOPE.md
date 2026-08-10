---
type: Scope
title: "kern-memory phase 2 — Scope"
description: "Graph/relationship layer and AvelFinances content population on top of the already-shipped EPIC-13 phase 1 memory brick."
tags: [planning, scope]
timestamp: 2026-08-10T00:00:00Z
status: final
---

# kern-memory phase 2 — Scope

## Problem
No `kern-*` brick had cross-run memory before this repo ([concept](../CONCEPT.md)).
Phase 1 closed that for exact-key facts (`.okf`) and semantic recall (vector), routed
by a neutral write/query contract. It left four things open: a relationship/graph layer
("what's connected to what, since when"), the blocked `kern-obs` transverse, no
per-caller trust separation, and an empty AvelFinances semantic layer. Phase 2 takes on
the two that trace directly to the one active client relationship — the graph layer and
real content — and defers the other two, which have no forcing function yet
([decision](/scope/01-problem-scope.md)).

## Users
`kern-orch` only, unchanged from phase 1. No second caller population is assumed for
this phase ([decision](/scope/02-target-users.md)).

## Goals & success criteria
- A caller writes a relationship between two existing memories and later recalls it
  through a multi-hop graph query — verified against a real running daemon, not mocked.
- Real, current AvelFinances bank-criteria content is written into the semantic layer at
  production scale, and a real prospect question correctly recalls the relevant
  criteria — the same rigor phase 1 held for its own end-to-end verification
  ([decision](/scope/11-success-criteria.md)).

## Non-goals (out of scope for v1)
- `kern-obs` (traced recall) — blocked on a brick that doesn't exist anywhere in the
  ecosystem yet.
- Per-caller / per-session trust separation — no second caller population exists.
- A single `query` call fanning out across `.okf`, vector, *and* graph together —
  additive polish once the two core journeys work independently.
- Swapping the vector store (pgvector/Qdrant) or the embeddings model — no scale
  pressure observed.
- Hot/warm/cold temperature-based memory tiers — considered and rejected in
  [`CONCEPT.md`](../CONCEPT.md); this project's model routes by query shape, not
  recency.

([decision](/scope/08-non-goals.md))

## Constraints
- Sovereignty: no external network call for anything touching data, unchanged from
  phase 1 — the argument already sold to AvelFinances.
- Same binary as the existing document store and phase-1 memory API; not a second
  daemon.

([decision](/scope/09-constraints.md))

## Milestone 1: Graph layer
A light, embedded graph on SQLite ([decision](/scope/04-graph-layer-approach.md)),
exposed through the existing `write`/`query` contract via a new `Kind="graph"`. Edges
reference existing `.okf` and vector memory IDs; the graph never stores its own copy of
content ([decision](/scope/05-graph-data-model.md)). Delivers the first core journey:
write a relationship between two memories, query it back through single-hop and
multi-hop traversal ([decision](/scope/06-core-user-journeys.md)).

## Milestone 2: AvelFinances content population
Real, current bank-criteria content written into the semantic layer at production
scale, replacing the three demo memories used to verify phase 1. Depends on Milestone 1
existing but is independently shippable and client-facing on its own — and depends on
an external input (AvelFinances delivering the actual criteria), not just engineering
work ([decision](/scope/12-milestones-phasing.md)).

## Risks & assumptions
- A light Go/SQLite graph may not match richer temporal-graph engines (e.g. Graphiti)
  if the real workload later needs multi-hop reasoning beyond simple relationships —
  assumed sufficient for today's need.
- Milestone 2 depends on AvelFinances actually delivering real criteria content; this
  can stall the milestone on something outside engineering's control.
- Deferring trust separation assumes no second caller population appears during phase
  2. If one does, [decision 02](/scope/02-target-users.md) needs reopening, not a
  silent workaround.

([decision](/scope/13-risks-assumptions.md))
