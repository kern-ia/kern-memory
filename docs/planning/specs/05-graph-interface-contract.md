---
type: Decision
title: "Graph interface contract"
description: "How write/query express an edge and a traversal, on top of the existing neutral contract."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 5
slug: graph-interface-contract
status: decided
verdict: "A. Extend memory.Memory/Query with optional edge fields, same two endpoints"
decided_via: triage
depends_on: [graph-storage-schema]
---

# Question
`memory.Memory` (`internal/memory/contract.go`) is a single-text unit with no natural
place for `from`/`to`/`relation`. How does writing an edge and running a multi-hop
query fit the existing `write`/`query` contract?

# Options
- **A. Extend `memory.Memory` with optional edge fields** (`FromID`, `ToID`, `Relation`)
  used only when `Kind == KindGraph`, and `memory.Query` with an optional `Depth` field
  used the same way. Same two HTTP endpoints (`POST /api/v1/memory/write`, `/query`),
  same `Router`, one more case in its `switch`.
- **B. Dedicated `/api/v1/memory/graph/write` and `/graph/query` endpoints** with their
  own request/response types — cleaner separation, but doubles the API surface for one
  layer and breaks the "one neutral contract" principle `CONCEPT.md` names as the point
  of this brick.
- **C. Encode an edge as `memory.Memory.Text`** (e.g. JSON `{"from":...,"to":...}` as
  the text body) — zero contract changes, but abuses `Text` (meant for
  pseudonymizable prose) to carry structured data, and pseudonymization would run over
  it meaninglessly.

# Recommendation
**A.** Keeps the "neutral write/query contract routed by kind" principle intact — the
same shape, one more `Kind`, explicit typed fields instead of a smuggled encoding.

# Verdict
Accepted at triage: A.

**Correction (2026-08-11, discovered during issue #7's implementation):** the
recommendation above under-specified option A — it named `FromID`/`ToID`/`Relation`
but omitted `FromKind`/`ToKind`, even though decision 03's whole point is a
`(kind, id)` composite reference to disambiguate `.okf` vs. vector IDs. Without them,
the fields the interface carries can't populate the columns the storage schema
requires, and issue #7's implementer correctly left `from_kind`/`to_kind` empty rather
than invent a value. Corrected: `memory.Memory` carries `FromKind`, `FromID`,
`ToKind`, `ToID`, `Relation` — all five, not three. The rest of the verdict (extend
the existing contract, no new routes) stands unchanged.
