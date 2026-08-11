---
type: Conventions
title: "kern-memory — Conventions"
description: "Personal baseline (Go/stdlib, filtered by SPECS.md) + 2 project deviations from root CONVENTIONS.md."
tags: [planning, conventions]
timestamp: 2026-08-10T00:00:00Z
status: final
baseline_version: 2026-07-20T13:00:00Z
---

# kern-memory — Conventions

## Code style & formatting
`gofmt` is law — run in CI, never debated in review. Lint: `golangci-lint`,
`linters.default: standard`, warnings are errors (reference config:
`Kern-Anon/.golangci.yml`). No commented-out code in committed files.

## Naming
Descriptive over short, no non-standard abbreviations. Go community casing —
`PascalCase` for exported identifiers, `camelCase` internal — never a house style.
Files named after the main thing they export/define. Code, identifiers, and comments
are English, no exceptions; internal docs (`README.md`, `CLAUDE.md`, this file) stay in
whatever language the team works in day to day.

## Repository layout
Top level stays small: source, tests, `docs/`, `scripts/`, config. Knowledge-style docs
(plans, references, this bundle) are OKF under `docs/planning/`; `README.md` is for
humans landing on the repo, not a knowledge dump.

## Git & PRs
([decision](/conventions/01-branch-and-commit-model.md))

- `main`: default, always deployable, protected — no direct pushes.
- `dev`: integration branch, protected — no direct pushes.
- Working branches: `feature/<slug>`, `fix/<slug>`, `chore/<slug>`, `docs/<slug>`,
  `test/<slug>`.
- Every change to `main` or `dev` goes through a PR, including solo work.
- Merging into `dev`: `--no-ff` merge commit.
- Conventional Commits: `type(scope): short summary`, imperative mood, no trailing
  period in the subject; body explains the *why*. No tool signature
  (`Co-Authored-By`, `Claude-Session`, or equivalent trailer) — the git author is
  enough.
- One subject per PR, linked to the issue or RFC it resolves; states the semver
  impact; no real personal data in code, fixtures, or the PR description.
- Merge gate: green CI (`gofmt`, `go vet`, `go test -race`, lint). No second human
  approval required — a solo repo, CI is the gate.
- No force-pushes to `main`.

## Testing
Strict red-green test-first wherever correctness is assertable behavior — domain
logic, calculations, data transforms, trust boundaries (APIs, validation, handlers).
This covers essentially all of `kern-memory`'s code, including the phase-2 graph
layer's SQL query correctness. `go test -race ./...` must be green before any PR.
Tests live beside or mirror source; names state the behavior asserted, not the method
called. Prefer integration-level tests over mocking where the real thing is cheap —
this repo's established practice (real SQLite, real Ollama in phase-1 tests); mock
only true externals.

## Error handling
Fail fast, never swallow an error silently — handle it meaningfully or propagate with
context (`fmt.Errorf("<pkg>: <op>: %w", err)`, the pattern already used throughout
`internal/memory/*`). Error messages state what was being attempted. User-facing and
log-facing errors are different audiences — write both where both exist.

## Dependencies
Prefer the standard library; a new dependency needs a one-line justification in the PR
that adds it. Pin or lock everything the ecosystem lets you lock. Go module path:
`github.com/yoann/kern-memory`, depends on `github.com/kern-ia/kern-anon` (replaced
locally via `replace ... => ../Kern-Anon`) — if `kern-anon` renames its module again,
mirror it here in the same PR, not afterward
([drift](/DRIFT.md#01--gomods-presidiogo-pin-fell-behind-kern-anons-module-rename):
missed on the first rename, caught and fixed during Epic 1).

## Documentation & comments
Code comments only for constraints the code can't show (invariants, workarounds,
non-obvious "why") — never narration of what the next line does. Every substantive doc
lands in the OKF bundle (`docs/planning/`), not scattered ad-hoc markdown. `README.md`
and `CLAUDE.md` stay synchronized with what's actually built. No `CHANGELOG.md` —
release notes live in the annotated tag.

## Security & privacy
([decision](/conventions/02-security-privacy-section.md))

`kern-memory` stores searchable, potentially sensitive memory — particular care never
to commit real user content in fixtures or tests. `SECURITY.md` inherited from the
`kern-ia` org once the repo moves under it. A field meant to carry structural data (a
label, an id, a relation type) is validated for shape, not run through `kern-anon` —
`kern-anon` is for scanning free text/prose, not enforcing that a field stays short and
non-prose (the pattern phase 2's `Relation` field follows,
[decision](/specs/11-security-privacy.md)). Any future field with the same
label-vs-prose ambiguity should ask the same question rather than defaulting to either
extreme.
