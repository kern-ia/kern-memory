---
type: Issue
title: "Add graph contract fields and Router routing"
description: "Extend memory.Memory/Query with edge fields and a KindGraph case in Router, with no storage behind it yet."
tags: [epic-1]
timestamp: 2026-08-11T09:00:00Z
epic: 1
issue: 1
slug: graph-contract-and-routing
size: S
status: in-progress
gh_issue: 6
resource: https://github.com/kern-ia/kern-memory/issues/6
depends_on: []
---

# Add graph contract fields and Router routing

## Summary
Lays the typed foundation the rest of Epic 1 builds on: `memory.Kind` gains
`KindGraph`, `memory.Memory` gains optional `FromID`/`ToID`/`Relation` fields, and
`memory.Query` gains an optional `Depth` field — active only when `Kind == KindGraph`
([spec](../../../planning/specs/05-graph-interface-contract.md)). `Router` gets a
`Graph memory.Store` field and routes `KindGraph` to it in both `Write` and `Query`,
matching the existing `KindOKF`/`KindVector` cases exactly.

## Scope
- `internal/memory/contract.go`: add `KindGraph`, the four new struct fields, and the
  `Router.Graph` field + routing cases (write: route on `m.Kind`; query: route on
  `q.Kind`, including the empty-`Kind` fan-out case gaining a third branch).
- Unit tests for the new routing paths using a fake `memory.Store` (same pattern
  `contract_test.go` already uses for `OKF`/`Vector`).

## Out of scope
- The actual graph storage backend (`internal/memory/graph`) — this issue routes to
  whatever `Router.Graph` is set to; a fake is enough here. Storage is issues 2 and 3.
- HTTP layer changes (`internal/httpapi/memory.go`) — issue 4.

## Acceptance criteria / Definition of done
- [ ] `Router.Write` with `Kind: memory.KindGraph` routes to `Router.Graph.Write`
      (test: a fake `Store` records the call).
- [ ] `Router.Query` with `Kind: memory.KindGraph` routes to `Router.Graph.Query`.
- [ ] `Router.Query` with empty `Kind` fans out to `OKF`, `Vector`, *and* `Graph`,
      merged and sorted by `Similarity` descending — extending the existing two-layer
      fan-out test in `internal/memory/contract_test.go`.
- [ ] `Router.Write`/`Query` with an unset `Router.Graph` and `Kind: memory.KindGraph`
      returns a clear error rather than a nil-pointer panic (mirrors how an unknown
      `Kind` already errors today).
- [ ] `go test -race ./internal/memory/...` green.

## Relevant files / areas
- `internal/memory/contract.go` — `Kind`, `Memory`, `Query`, `Router` (existing file,
  read in full before editing).
- `internal/memory/contract_test.go` — existing fake-`Store`-based routing tests to
  extend, not replace.

## Dependencies
None — this is the epic's first issue.

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it before opening the PR.
Expected well under that — this is struct fields plus `switch` cases and their tests.
