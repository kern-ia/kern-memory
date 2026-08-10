---
type: Decision
title: "Milestones / phasing"
description: "How phase 2 phases into independently shippable chunks."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 12
slug: milestones-phasing
status: decided
verdict: "A. Two milestones: graph layer, then AvelFinances content population"
decided_via: triage
depends_on: [mvp-cut]
---

# Question
How does the phase 2 MVP cut (decision 07) split into independently shippable
milestones? This shapes exactly what `split-epics` later cuts on.

# Options
- **A. Two milestones.** Milestone 1 — graph layer (write/query API extension,
  `Kind="graph"`, edges referencing existing memory IDs). Milestone 2 — AvelFinances
  content population and validation, which depends on milestone 1 existing but is
  independently shippable and client-facing on its own.
- **B. One milestone.** Ship both together — smaller pipeline overhead, but a
  multi-week combined unit that isn't independently releasable partway through.

# Recommendation
**A.** Matches the "weeks not an afternoon, independently shippable" bar this skill
sets, and separates a purely technical milestone (graph layer) from one with an
external dependency (getting real criteria data from AvelFinances) that could stall on
something outside engineering's control.

# Verdict
Accepted at triage: A.
