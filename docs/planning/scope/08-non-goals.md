---
type: Decision
title: "Non-goals for phase 2"
description: "What is explicitly out of phase 2, especially tempting-but-deferred items."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 8
slug: non-goals
status: decided
verdict: "A. Full non-goals list"
decided_via: triage
depends_on: [mvp-cut]
---

# Question
What's explicitly excluded from phase 2, so it doesn't creep back in through issues?

# Options
- **A.** `kern-obs` (traced recall) — blocked on a brick that doesn't exist.
  Per-caller/per-session trust separation — no second caller yet (decision 02).
  Cross-layer fan-out query across `.okf`/vector/graph in one call (decision 06, option
  B) — additive polish, not required for the two core journeys.
  Swapping the vector store to pgvector/Qdrant, or the embeddings model — no scale
  pressure observed yet.
  Hot/warm/cold temperature-based memory tiers — considered and rejected in
  `CONCEPT.md`, not this project's model.
- **B.** Narrower list — drop the cross-layer fan-out and store-swap items since nobody
  asked for them, keep only the two blocked-on-external-brick items explicit.

# Recommendation
**A.** Cheap to write down now; each item is something a future issue could plausibly
propose "while we're in there" — naming them explicitly is what keeps this milestone
from absorbing them.

# Verdict
Accepted at triage: A.
