---
type: Decision
title: "Risks & assumptions"
description: "What could sink phase 2, and what's assumed without proof."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 13
slug: risks-assumptions
status: decided
verdict: "A. Named risks and assumptions"
decided_via: triage
depends_on: [mvp-cut]
---

# Question
What are the concrete risks and unproven assumptions behind the phase 2 plan?

# Options
- **A.** Name them:
  - A light Go/SQLite graph may not match Graphiti-style temporal reasoning quality if
    the real workload turns out to need it — assumption: today's need (simple
    relationships, provenance) doesn't require that yet.
  - Content population (milestone 2) depends on AvelFinances actually delivering real,
    current bank-criteria content — an external dependency, not an engineering task;
    the milestone can stall on something outside this team's control.
  - Deferring trust separation (decision 02) assumes no second caller population shows
    up during phase 2 — if one does mid-phase, decision 02 needs reopening, not a
    silent workaround.
- **B.** Skip formal risk-listing, handle issues as they arise.

# Recommendation
**A.** Cheap now, per the checklist's own framing — especially the AvelFinances content
dependency, which is the one risk most likely to actually block milestone 2.

# Verdict
Accepted at triage: A.
