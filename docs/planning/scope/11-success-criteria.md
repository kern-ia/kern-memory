---
type: Decision
title: "Success criteria"
description: "How you'd know phase 2 worked."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 11
slug: success-criteria
status: decided
verdict: "A. Real, unmocked end-to-end verification of both core journeys"
decided_via: triage
depends_on: [mvp-cut]
---

# Question
What checkable outcome tells you phase 2 succeeded — mirroring how phase 1 was
verified with a real daemon, real `curl` calls, and a real prospect question?

# Options
- **A.** A real multi-hop graph query, against real relationships written by a real
  caller (not mocked), returns the correct chained result — same verification bar phase
  1 held for the semantic layer (real Ollama, real embeddings, real similarity
  comparison). And: a real AvelFinances prospect question, run against production-scale
  criteria content (not the three demo memories), correctly recalls the relevant
  criteria.
- **B.** Test coverage alone (unit + integration tests green) without a real end-to-end
  demonstration.

# Recommendation
**A.** Consistent with the standard phase 1 already set and with the repo's own method
(`CLAUDE.md`: "vérifié en réel" for every prior feature) — B would be a regression in
rigor from what's already been the norm in this repo.

# Verdict
Accepted at triage: A.
