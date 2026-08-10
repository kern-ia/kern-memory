---
type: Decision
title: "Core user journeys for phase 2"
description: "The 2-5 end-to-end flows phase 2 must support."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 6
slug: core-user-journeys
status: decided
verdict: "A. Two journeys: multi-hop relationship write/query; real AvelFinances content recall"
decided_via: triage
depends_on: [problem-scope, graph-layer-approach, graph-data-model]
---

# Question
What are the concrete flows phase 2 must make work end-to-end, that become
acceptance-criteria seeds for its epics?

# Options
- **A. Two journeys, matching decision 01(C):**
  1. A caller writes a relationship between two existing memories (e.g. "criterion X
     supersedes criterion Y") and later queries it, getting a traversal result — not
     just a single record.
  2. Real AvelFinances bank-criteria content is written into the semantic layer at
     production scale (not the three demo memories from phase 1's verification) and a
     real prospect question correctly recalls the relevant criteria.
- **B. Add a third: cross-layer query.** A single `query` call that fans out across
  `.okf`, vector, *and* graph, merging results — mirrors how `query` already fans out
  across `.okf`/vector today.

# Recommendation
**A.** B is a real extension of the existing fan-out pattern, but it's additive
polish once A's two flows work independently — including it now risks the same
scope-creep the checklist warns against. Worth a non-goal note (see decision 08)
rather than silently dropped.

# Verdict
Accepted at triage: A. Cross-layer fan-out query (option B) stays out — named in
non-goals (decision 08) rather than silently dropped.
