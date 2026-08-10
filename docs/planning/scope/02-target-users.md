---
type: Decision
title: "Target users for phase 2"
description: "Who is phase 2 built for, specifically?"
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 2
slug: target-users
status: decided
verdict: "A. kern-orch only, unchanged"
decided_via: triage
depends_on: [problem-scope]
---

# Question
Phase 1 has one caller (`kern-orch`). Does phase 2 stay scoped to that same caller, or
does it design for a second population of callers?

# Options
- **A. `kern-orch` only, unchanged.** No new caller assumed; the single bearer token
  stays sufficient.
- **B. Design for a second caller now.** Anticipates a future population (e.g. a second
  `kern-*` brick, or direct AvelFinances access) and builds the trust separation for it.

# Recommendation
**A.** Consistent with decision 01 (C) deferring trust separation — there is no second
caller today, and building for one that doesn't exist yet is speculative scope.

# Verdict
Accepted at triage: A. No new caller population assumed for phase 2.
