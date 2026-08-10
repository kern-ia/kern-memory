---
type: Epic
title: "AvelFinances content population"
description: "Write real, current bank-lending-criteria content into the semantic (and where relevant, graph) memory layers at production scale."
tags: [epic]
timestamp: 2026-08-10T10:00:00Z
epic: 2
slug: avelfinances-content-population
status: open
gh_issue: 4
milestone: 2
resource: https://github.com/kern-ia/kern-memory/issues/4
source: docs/planning/SCOPE.md#milestone-2-avelfinances-content-population
---

# Epic 2: AvelFinances content population

## Goal
EPIC-13 phase 1 proved the semantic layer works end to end (real embeddings, real
similarity ranking), verified with three demo memories. What it actually knows depends
entirely on what's written into it — and today that's still demo content, not the real
bank-lending criteria AvelFinances (the client whose need triggered this whole brick)
needs their agents to recall. This epic is what makes the two shipped layers, and the
new graph layer from Epic 1, actually useful to the one real client waiting on them.

## Scope
- Write real, current AvelFinances bank-lending-criteria content into the semantic
  layer, at production scale — not a handful of demo entries.
- Where a criterion has a natural relationship to another (supersedes, depends on,
  applies-to), record it via Epic 1's graph layer rather than leaving the relationship
  implicit in prose.
- Verify recall end-to-end: a real prospect question, run against the populated
  content, correctly returns the relevant criteria.

## Out of scope
- Any change to the storage layers themselves (`.okf`, vector, graph) — this epic is
  content, not code, except where a real gap surfaces while populating (handled as
  drift, not silently patched).
- Ongoing content maintenance/refresh process — this epic delivers the initial
  population, not a recurring pipeline.

## Acceptance criteria
- Real, current bank-lending-criteria content (not demo data) is written into the
  semantic layer.
- A real prospect question — of the kind AvelFinances agents actually field — correctly
  recalls the relevant criteria, verified against a real running daemon.
- Where relationships between criteria exist, they're recorded via the graph layer and
  recallable through it.

## Dependencies
[Epic 1: Graph layer](/epic-1-graph-layer/EPIC_1.md) — this epic uses the graph layer
to record relationships between criteria; it can start once Epic 1's write/query
contract exists, even before Epic 1's own acceptance criteria are fully closed out, but
cannot finish before it.

## Context
- [Technical specs](../../planning/SPECS.md) — the write/query contract this epic's
  content goes through.
- [Conventions](../../planning/CONVENTIONS.md) — repo standards this epic's issues
  follow.

## Notes
Risk carried from `SCOPE.md`: this epic depends on AvelFinances actually delivering
real, current criteria content — an external dependency, not an engineering task. It
can stall on something outside this team's control; that's expected, not a signal to
substitute placeholder content to look finished.

Pseudonymization applies here as it does to every write: content goes through
`kern-anon` before persisting, same as it did for phase 1's own demo verification.
