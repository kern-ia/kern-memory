---
type: Decision
title: "Phase 2 problem scope"
description: "Which of CONCEPT.md's four open questions does phase 2 actually tackle, and in what priority?"
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 1
slug: problem-scope
status: decided
verdict: "C. Graph layer + AvelFinances content population"
decided_via: triage
depends_on: []
---

# Question
`CONCEPT.md` names four things phase 1 left open: the graph/relationship layer, the
blocked `kern-obs` transverse, missing per-caller trust separation, and populating the
AvelFinances semantic layer with real content. Phase 2 can't sensibly chase all four at
once — which does it actually take on, and in what priority?

# Options
- **A. All four in one phase.** Complete closure of every open item at once — largest
  scope, slowest to ship, and two items (`kern-obs`, trust separation) have no forcing
  function behind them yet.
- **B. Graph layer only.** The single biggest named technical gap (multi-hop,
  provenance, "what's connected to what, since when") — defer everything else.
- **C. Graph layer + AvelFinances content population.** Both serve the one active
  client relationship directly — the graph layer gives it real relational recall, and
  the content population is what makes the semantic layer actually useful to them today,
  independent of the graph work.

# Recommendation
**C.** `kern-obs` is blocked on a brick that doesn't exist anywhere in the ecosystem —
not this team's to build speculatively. Trust separation has no second caller yet, so
building it now is designing for a population that may never show up. The graph layer
and content population both trace directly to the one real, dated client need.

# Verdict
Accepted at triage: C. Phase 2 tackles the graph layer and AvelFinances content
population; `kern-obs` and trust separation stay deferred.
