---
type: Decision
title: "Traversal depth limit"
description: "How deep a multi-hop graph query is allowed to walk."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 10
slug: traversal-depth-limit
status: decided
verdict: "A. Hard-capped server-side max depth, same clamping precedent as Limit"
decided_via: triage
depends_on: [graph-interface-contract]
---

# Question
A multi-hop query (decision 05) needs a bound, or a caller-supplied depth against a
dense graph becomes an expensive, effectively unbounded walk. `docs/retro.md` already
documents the same class of problem once, on `chromem-go`'s `Query`: an
unclamped caller parameter (`Limit`) broke on a collection smaller than expected, fixed
by having the server clamp it rather than trust the caller. Does the graph query get
the same treatment for depth?

# Options
- **A. Hard-capped server-side max depth** (e.g. 3), regardless of what `Query.Depth`
  requests — the server clamps down to the cap, the same shape as the existing
  `Limit`-clamping fix.
- **B. Caller-specified depth, no cap.** Simplest to implement, no protection as
  real content grows.

# Recommendation
**A.** Directly follows a precedent already established — and already paid for once —
in this exact codebase: a caller-supplied traversal/result-size parameter must be
clamped server-side, not trusted.

# Verdict
Accepted at triage: A.
