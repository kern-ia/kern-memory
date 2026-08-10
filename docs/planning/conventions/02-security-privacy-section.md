---
type: Decision
title: "Security & privacy section"
description: "The baseline has no security section at all — what fills that gap for this project."
tags: [decision, conventions]
timestamp: 2026-08-10T00:00:00Z
phase: conventions
decision: 2
slug: security-privacy-section
status: decided
verdict: "A. Merge root CONVENTIONS.md's policy with the SPECS.md graph-field rule"
decided_via: triage
depends_on: []
---

# Question
`assets/baseline.md` has no Security & privacy section. This project needs one: it's a
memory store handling pseudonymizable content, root `CONVENTIONS.md` already has a
short policy, and phase 2's `SPECS.md` just added a specific rule (`Relation` is
validated as a label, not pseudonymized — [decision](/specs/11-security-privacy.md)).
What goes in the gap?

# Options
- **A. Merge root `CONVENTIONS.md`'s existing policy with the phase-2 SPECS.md rule.**
  No real user content in fixtures/tests, `SECURITY.md` org-inherited once the repo
  moves under `kern-ia` (root's existing text), plus a pointer to SPECS.md's graph-field
  validation rule so a future contributor adding a fourth memory layer knows to ask the
  same "does this field carry prose or a label" question.
- **B. Leave it uncovered** — rely on SPECS.md alone for the graph-specific rule and
  root `CONVENTIONS.md` for the general one, don't consolidate into this doc.

# Recommendation
**A.** This doc's whole point is to be the one place a reader (or `implement-issue`)
checks for standards — leaving security split across two other documents defeats that,
and the merge costs one section.

# Verdict
Accepted at triage: A. Worth promoting a generic "Security & privacy" section shape
into the baseline itself — every project handling any sensitive data hits this same
gap; the specific content here (kern-anon, `SECURITY.md` inheritance) stays
project-specific.
