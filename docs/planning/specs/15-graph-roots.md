---
type: Decision
title: "Graph roots — entry points for a view with no starting node"
description: "How a consumer opens a broad initial graph view when the traversal contract always requires a starting node."
tags: [decision, specs]
timestamp: 2026-08-14T00:00:00Z
phase: specs
decision: 15
slug: graph-roots
status: decided
verdict: "A. A reserved tag on an existing OKF memory — no new field, no new endpoint"
decided_via: triage
depends_on: [graph-interface-contract, traversal-depth-limit]
---

# Question
`graph.Store.Query` (decision 05) always requires `from_kind`/`from_id` — a caller must
already know a node to traverse from. Kern-ui's Cerveau view (`Kern-UI/docs/
expected-contracts.md`, C7) needs to open on a broad view with no single starting node in
mind (the mockup shows *248 souvenirs actifs*, not one node's neighbourhood). What marks
which memories a consumer should start from?

# Options
- **A. A reserved tag on an existing OKF memory.** `POST /api/v1/memory/write
  {"kind":"okf", "tags":["cerveau-racine"], ...}` is already the whole write path — no new
  field, no new storage, no new endpoint. Listing roots is exactly the exact-tag-match
  query `okf.Store.Query` already implements (`hasAllTags`):
  `{"kind":"okf","tags":["cerveau-racine"]}`.
- **B. A new `Root bool` field on `memory.Memory`**, threaded through all three storage
  layers. Closer to a first-class concept, but real plumbing for something a tag already
  expresses, and the vector layer has no way to list rows without an embedding query
  (`chromem-go`'s `Query` requires text) — a vector-kind root would need yet another new
  capability just to be listed.
- **C. A separate roots table/endpoint.** Cleanest conceptual separation, but the same
  "one neutral write/query contract" principle decision 05 already chose over dedicated
  endpoints applies here too.

# Recommendation
**A.** Deliberately restricted to OKF: this repo's own README already frames OKF as the
"curated, auditable, someone chose this" layer — a root is exactly that kind of thing,
chosen explicitly by whoever writes it, not derived. The vector layer's structural
inability to list without an embedding call makes a vector-kind root a real complication
(option B) for a need OKF already covers honestly. Nothing here blocks a future root
concept for other layers if one is ever needed — it would be additive, not a rework.

# Verdict
Accepted at triage: A. `cerveau-racine` is the reserved tag; any caller may write more
than one root, and a query with no `tags` still returns every OKF memory including roots
— they are not a separate kind, only a marked subset of the existing one.
