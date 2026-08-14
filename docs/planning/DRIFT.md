---
type: Drift
title: "kern-memory — Drift"
description: "Standards this codebase has drifted from, and why"
tags: [planning, drift]
timestamp: 2026-08-11T00:00:00Z
---

# Drift

## Epic 1: Graph layer

### 01 — `go.mod`'s PresidioGo pin fell behind Kern-Anon's module rename

- **Decided**: `CONVENTIONS.md` (root) — "if `kern-anon` renames its module, mirror it
  here in the same PR, not afterward."
- **Actual**: `Kern-Anon` renamed its module from `github.com/YoLaub/PresidioGo` to
  `github.com/kern-ia/kern-anon` without a matching `kern-memory` PR landing at the
  same time — `go.mod`'s `require`/`replace` and every import in
  `internal/memory/anon` still pointed at the old name, breaking `go build ./...` on
  the main checkout.
- **Because**: verified by running `go build ./...` from the main checkout during
  issue #9's real-daemon verification — exact errors and the confirming
  `git -C ../Kern-Anon log` check are in the drift record below.
- **Disposition**: resolved (2026-08-11) — PR #17 merged into `dev`: `go.mod` and
  every affected import updated to `github.com/kern-ia/kern-anon`. Verified on the
  main checkout: `go build ./...` and `go test ./...` green across all 8 packages.
- **Revisit when**: n/a — resolved.
- **Evidence**: [drift record](../epics/epic-1-graph-layer/drift/01-presidiogo-module-rename-blocks-local-build.md), PR #17.
