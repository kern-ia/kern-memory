---
type: Decision
title: "Phase 2 MVP cut"
description: "Precisely what ships in phase 2."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 7
slug: mvp-cut
status: decided
verdict: "A. Graph layer (Kind=\"graph\") + AvelFinances content population, same milestone track"
decided_via: triage
depends_on: [problem-scope, graph-layer-approach, graph-data-model, core-user-journeys]
---

# Question
What exactly is in phase 2 v1, stated as concrete, buildable features?

# Options
- **A. Graph layer (write + query, single-hop and multi-hop) + AvelFinances content
  population.** A new `Kind="graph"` on the existing `write`/`query` contract, backed by
  SQLite, storing edges that reference existing `.okf`/vector memory IDs (decision 05);
  plus writing real, current bank-criteria content into the semantic layer at scale.
- **B. Graph layer only, defer content population to a separate initiative.** Ships the
  technical capability without the client-facing payoff in the same milestone.

# Recommendation
**A.** Splitting the two (as in B) ships a capability nobody's using yet — the content
population is what makes the graph (and the existing semantic layer) actually valuable
to the one real client waiting on it. Keeping them in the same v1 keeps the milestone
honestly "shippable and useful," not just "code complete."

# Verdict
Accepted at triage: A. See decision 12 for how this splits into two shippable
milestones.
