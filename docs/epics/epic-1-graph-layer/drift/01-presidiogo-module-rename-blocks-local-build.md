---
type: Drift
title: "Epic 1 issue 4 — kern-memory's PresidioGo pin no longer matches Kern-Anon's module"
description: "go.mod's replace directive for github.com/YoLaub/PresidioGo cannot resolve against Kern-Anon's current main/dev — its module was renamed upstream."
tags: [epic-1, drift]
timestamp: 2026-08-11T18:55:00Z
epic: 1
issue: 4
gh_issue: 9
---

# Drift: PresidioGo pin vs. Kern-Anon's renamed module

- **Decided**: `go.mod` requires `github.com/YoLaub/PresidioGo`, resolved locally via
  `replace github.com/YoLaub/PresidioGo => ../Kern-Anon` (a sibling checkout next to
  `kern-memory`, per CLAUDE.md's cross-repo layout).
- **Actual**: `../Kern-Anon` on disk (`/Users/yoann/Developer/SERENIS_PROJET/Kern-Anon`,
  `main`/`dev`) has renamed its module to `github.com/kern-ia/kern-anon`
  (commits `b86b6a5`/`e787491`/`97a6c57` in that repo). Building this worktree's
  `cmd/kern-memory` (which imports `internal/memory/anon`, which imports
  `github.com/YoLaub/PresidioGo/*`) fails two ways in sequence:
  1. From a nested `.claude/worktrees/<id>` checkout, the relative path `../Kern-Anon`
     doesn't even point at the real sibling repo (it resolves one level too shallow) —
     `replacement directory ../Kern-Anon does not exist`.
  2. Once that path is bridged (a symlink or a `git worktree add` at the expected
     relative location), the build fails differently: `no required module provides
     package github.com/kern-ia/kern-anon/contextaware` (and 7 similar), because
     Kern-Anon's current default branches no longer use the `PresidioGo` module name
     `go.mod` still pins.
- **Because**: verified by running `go build ./...` from this worktree's root at each
  stage — exact errors captured above. Confirmed via `git -C ../Kern-Anon log --oneline`
  that the module rename is real and lands on both `main` and `dev`, not a stray branch.
- **Workaround used for this issue's real-daemon verification only (not committed)**:
  `git -C /Users/yoann/Developer/SERENIS_PROJET/Kern-Anon worktree add --detach
  <path-matching-../Kern-Anon> e9784f0` — `e9784f0` is Kern-Anon's last commit still
  under the `github.com/YoLaub/PresidioGo` module name, one commit before the rename.
  That let `cmd/kern-memory` build and the real daemon run for this issue's curl
  verification; the temporary worktree was removed afterward and nothing in
  `kern-memory`'s `go.mod`/`go.sum` changed.
- **Disposition**: deferred — out of scope for issue 4 (HTTP wiring), and not this
  repo's decision to make unilaterally: bumping the `replace` target and
  `github.com/YoLaub/PresidioGo` import paths to `github.com/kern-ia/kern-anon` touches
  `internal/memory/anon` across the board, a separate, focused change.
- **Revisit when**: before Epic 2 (or any future issue) needs a real daemon build in a
  *fresh* worktree — every such run will hit the same two-stage failure until
  `kern-memory` either updates its `PresidioGo` import paths to
  `github.com/kern-ia/kern-anon` and re-points the `replace` directive, or Kern-Anon
  publishes a tagged release under its new module name that `go.mod` can pin to instead
  of a local `replace`.
- **Evidence**: this record; PR for issue 9 (#9, Epic 1 issue 4).
