---
type: Issue
title: "Verify end-to-end recall against the simulated content"
description: "A real prospect-style question against the simulated bank-criteria content, single-hop and multi-hop, with a documented process to repeat once real content replaces it."
tags: [epic-2]
timestamp: 2026-08-11T23:57:00Z
epic: 2
issue: 3
slug: end-to-end-recall-verification
size: S
status: pr-open
gh_issue: 21
gh_pr: 25
resource: https://github.com/kern-ia/kern-memory/issues/21
depends_on: [simulated-bank-criteria-content]
---

# Verify end-to-end recall against the simulated content

## Summary
Closes Epic 2's acceptance criteria: a real prospect-style question, run against the
simulated content from issue 2, correctly recalls the relevant criteria — semantic
recall alone, and, where relevant, through a graph traversal. This also produces a
**repeatable process**, documented once, so the same verification can be re-run
unchanged when real AvelFinances content eventually replaces the simulated set — the
user's stated goal is an easily-updatable content zone, and "how do we know the new
content still works" is part of that, not a one-off check.

## Scope
- Start a real `kern-memory serve` instance with issue 2's content loaded.
- Run at least 2 realistic prospect questions (fictional, matching the fictional
  content) against `POST /api/v1/memory/query`, confirming the correct criteria rank
  above unrelated content — same bar phase 1's own verification held (real
  similarity scores compared, not just "some result came back").
- For at least one pair of criteria connected by a graph edge (from issue 2), run a
  `kind=graph` query and confirm the relationship is recallable.
- Write up the exact commands (curl or equivalent) as a short runbook — e.g.
  `docs/epics/epic-2-avelfinances-content-population/content/verification.md` — so
  re-running this after a future content update is "follow these steps," not
  "re-derive the verification from scratch."

## Out of scope
- Automating this as a CI check — no CI exists yet in this repo
  ([spec](../../../planning/specs/08-testing-infrastructure.md) inherits phase 1's
  real-backend-test approach, not a new CI pipeline).
- Any change to the memory layers' code — this issue verifies, it doesn't implement.

## Acceptance criteria / Definition of done
- [x] At least 2 prospect-style questions correctly recall their matching criteria
      via `kind=vector` (or empty-`kind` fan-out) query, with similarity scores
      pasted into the PR — mirroring phase 1's own retro
      (`docs/index/0002-epic13-phase1.md`'s "0,83 vs 0,56" style evidence).
  - [x] At least 1 graph traversal (`kind=graph`) correctly recalls a relationship
      between two loaded criteria.
- [x] `verification.md` runbook committed, containing the literal commands used.
- [x] Real daemon, real curl — no mocking, consistent with this repo's established
      practice.

## Relevant files / areas
- Issue 2's content file and the running daemon.
- `internal/httpapi/memory.go` — the query endpoint being exercised (read-only for
  this issue, no changes expected).

## Dependencies
Blocked by [issue 2](./02-simulated-bank-criteria-content.md) (needs the content
loaded to query against). Closes out Epic 2.

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it before opening the PR.
Expected small — this is verification + a runbook doc, not new code.
