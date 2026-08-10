---
type: Decision
title: "Branch, merge & commit-trailer model"
description: "Deviate from the personal baseline's branch naming / review-gate defaults in favor of the project's existing root CONVENTIONS.md."
tags: [decision, conventions]
timestamp: 2026-08-10T00:00:00Z
phase: conventions
decision: 1
slug: branch-and-commit-model
status: decided
verdict: "A. Adopt root CONVENTIONS.md's model"
decided_via: triage
depends_on: []
---

# Question
The personal baseline defaults to `issue-<n>-<slug>` branch names and is silent on
commit trailers. This repo's own (pre-existing, uncommitted) root `CONVENTIONS.md`
already decides both differently — and this session has already committed three times
following it (`docs/concept-kern-memory` branch, no tool co-author trailer). Does
`docs/planning/CONVENTIONS.md` adopt the baseline defaults, or formalize what
`CONVENTIONS.md` (and this session's actual git history) already does?

# Options
- **A. Adopt root `CONVENTIONS.md`'s model as the project's convention.** Typed-prefix
  branches (`feature/`, `fix/`, `chore/`, `docs/`, `test/`) off protected `main`/`dev`,
  PR with `--no-ff` merge into `dev`; commits are Conventional Commits with no tool
  signature (`Co-Authored-By` or equivalent) — "the git author is enough."
- **B. Adopt the baseline's `issue-<n>-<slug>` naming and drop the no-trailer rule.**
  Consistent with other projects using this baseline, but contradicts the root doc and
  the git history already produced this session.

# Recommendation
**A.** Changing now would mean renaming an active branch and rewriting already-pushed
commits for no functional gain — the baseline's alternative isn't wrong, it's just not
what this specific repo already committed to, in writing and in git history, before
this skill ran.

# Verdict
Accepted at triage: A. Not project-specific — this is really "the personal baseline's
branch-naming/no-trailer defaults should probably just say this," worth a look via
the improve-skill loop rather than re-deciding it per project.
