---
type: Issue
title: "Author and load simulated bank-lending criteria content"
description: "Realistic but fictional bank-criteria memories and their relationships, loaded via the issue 1 CLI tool, standing in for AvelFinances' real (confidential, not-yet-available) content."
tags: [epic-2]
timestamp: 2026-08-12T09:00:00Z
epic: 2
issue: 2
slug: simulated-bank-criteria-content
size: S
status: open
gh_issue: 20
resource: https://github.com/kern-ia/kern-memory/issues/20
depends_on: [editable-content-format-and-loader-cli]
---

# Author and load simulated bank-lending criteria content

## Summary
AvelFinances' actual bank-lending criteria belong to the agency and aren't available
yet — this issue populates the semantic and graph layers with **simulated, fictional**
content standing in for it: realistic mortgage-lending criteria (age limits, property
types like SCI, income requirements, etc.) with no real bank names, no real client
data, nothing confidential — written using the issue 1 loader so updating it later is
"edit the file, re-run the command," not an engineering task.

## Scope
- A content file (using issue 1's format) with a realistic-but-fictional set of
  bank-lending criteria as `vector`-kind memories — similar shape to phase 1's own
  verification content (`docs/index/0002-epic13-phase1.md` mentions "deux critères
  bancaires réalistes SCI/senior" as its phase-1 demo — this issue produces a larger,
  more structured set in that same spirit, clearly fictional).
- Where criteria genuinely relate to each other (e.g., a stricter rule superseding a
  general one, or one criterion applying only when another condition holds), record
  that relationship as a graph edge in the same file, per Epic 1's graph layer.
- Load the file into a running `kern-memory` instance via `load-memory`.
- A short note in the PR (or a comment in the content file itself) stating plainly
  that this content is simulated/fictional, not real AvelFinances data — so nobody
  downstream mistakes it for production content.

## Out of scope
- Any real AvelFinances data — not available, and out of scope for this issue by
  design.
- Changes to the loader tool itself (issue 1) — if the format proves insufficient
  while authoring content, note it as a follow-up rather than reopening issue 1's
  scope here.

## Acceptance criteria / Definition of done
- [ ] Content file committed under the epic folder (e.g.
      `docs/epics/epic-2-avelfinances-content-population/content/bank-criteria.json`),
      containing at least 8-10 distinct criteria memories and at least 2 graph edges
      between related ones.
- [ ] Loaded into a real running `kern-memory serve` instance via `load-memory`,
      verified by querying at least 3 of the loaded memories back.
- [ ] Content is clearly marked as simulated/fictional (file header comment or
      adjacent doc note).
- [ ] No real personal or organizational data anywhere in the file (per
      `CONVENTIONS.md`'s Security & privacy section).

## Relevant files / areas
- New: `docs/epics/epic-2-avelfinances-content-population/content/bank-criteria.json`
  (or similar path — no existing convention for this in the repo, this issue
  establishes it).
- The `load-memory` CLI command from issue 1.

## Dependencies
Blocked by [issue 1](./01-editable-content-format-and-loader-cli.md) (needs the
loader and file format). Blocks [issue 3](./03-end-to-end-recall-verification.md).

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it before opening the PR.
Expect this to be small in code, larger in data (the content file itself) — the
~500-line guidance is about reviewable diff, and a data file of criteria is easy to
skim even if it runs longer than typical code.
