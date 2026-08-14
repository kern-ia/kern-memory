---
type: Decision
title: "Resolving a memory's own content by id"
description: "A graph traversal returns edges, never the text/tags of the memories at either end — how does a caller get them back?"
tags: [decision, specs]
timestamp: 2026-08-14T00:00:00Z
phase: specs
decision: 16
slug: resolve-by-id
status: decided
verdict: "A. An optional IDs field on the existing Query, one call per layer"
decided_via: triage
depends_on: [graph-interface-contract]
---

# Question
`graph.Store.Query` (decision 05) returns edges — `(from_kind, from_id, to_kind, to_id,
relation)` — never the text/tags of the memories those ids reference. A consumer
rendering a graph (e.g. kern-ui's Cerveau, C7) needs to show a label for each node it
reaches, not just how they connect. Neither `okf.Store.Query` nor `vector.Store.Query`
supports looking a memory up by id today — only tag match and text-similarity search.
How does a caller resolve a set of ids to their content?

# Options
- **A. An optional `IDs []string` field on the existing `memory.Query`.** Set, it means
  "return exactly these" instead of searching: `okf.Store.Query` gains a
  `WHERE id IN (...)` branch alongside its tag filter (mutually exclusive with a text/tag
  search, same request shape); `vector.Store.Query` gains a branch that calls
  `chromem-go`'s `Collection.GetByID` per id instead of embedding `q.Text` — a real,
  already-available capability of the dependency this repo has never called. Same "extend
  the existing neutral contract" precedent as `Depth` (decision 05) and `Kind == KindGraph`
  itself.
- **B. A dedicated `POST /api/v1/memory/resolve` endpoint.** Cleaner intent signalling,
  but the same objection decision 05 already raised against dedicated endpoints: it
  doubles the API surface for what the existing contract can express with one more field.
- **C. Leave it to the caller** — issue N separate tag/similarity queries and hope they
  land on the right memory. Not reliable (a tag query can match more than one memory; a
  similarity query is a *search*, not a lookup, and offers no guarantee the intended
  memory is even in the top results) — rejected outright, not a real option.

# Recommendation
**A.** Same shape as every other extension this contract has taken (`Kind`, `Depth`) —
one more optional field, routed per layer, no new endpoint. `graph.Store` is untouched:
an edge is never looked up by id this way, only the memories it references are.

# Verdict
Accepted at triage: A. A batched call — `{"kind":"okf","ids":[...]}` /
`{"kind":"vector","ids":[...]}` — resolves many ids in one request; a caller with ids
split across layers makes one call per layer, not one call per id.
