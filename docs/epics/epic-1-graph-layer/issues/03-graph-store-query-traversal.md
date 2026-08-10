---
type: Issue
title: "Add graph store Query: single-hop and clamped multi-hop traversal"
description: "Query on the graph store, backed by a recursive SQL traversal, with a server-side depth cap."
tags: [epic-1]
timestamp: 2026-08-10T11:00:00Z
epic: 1
issue: 3
slug: graph-store-query-traversal
size: M
status: open
gh_issue: 8
resource: https://github.com/kern-ia/kern-memory/issues/8
depends_on: [graph-store-write]
---

# Add graph store Query: single-hop and clamped multi-hop traversal

## Summary
Completes `internal/memory/graph`'s `memory.Store` implementation. `Depth <= 1` (or
unset) returns direct edges only; `Depth > 1` walks the graph via SQLite's
`WITH RECURSIVE` up to a hard server-side cap, regardless of what the caller requested
— the same clamping precedent already set for `chromem-go`'s `Limit`
([spec](../../../planning/specs/10-traversal-depth-limit.md), retro.md's own
documented piège on trusting a caller-supplied size parameter).

## Scope
- `Query(ctx, q memory.Query) ([]memory.Recall, error)`:
  - Filters by `q.Tags`-equivalent inputs the graph layer actually has (`FromID`/
    `ToID` on `q`, or however issue 1 shaped the query-side fields — reconcile
    against issue 1's actual field names, which may need one additional field on
    `memory.Query` beyond `Depth` to specify the traversal's starting node; flag this
    to the reviewer if issue 1 didn't anticipate it).
  - `Depth <= 1`: direct edges from/to the given node.
  - `Depth > 1`: `WITH RECURSIVE` traversal, capped at a constant max depth (e.g. 3)
    defined in this package — `min(q.Depth, maxDepth)`, never trust `q.Depth` alone.
  - `Similarity` on each `Recall` is `1` (exact structural match, no ranking) — same
    convention `okf.Store.Query` already uses.
- A named constant for the max depth, with a comment explaining why it's capped
  (same style as the `busyTimeoutMS` comment in `okf/store.go`).

## Out of scope
- Cross-layer fan-out (a single query merging `.okf`/vector/graph results into one
  ranked list) — explicit non-goal for this phase
  ([scope decision](../../../planning/scope/06-core-user-journeys.md)).
- Any caching of traversal results.

## Acceptance criteria / Definition of done
- [ ] Querying a direct edge (`Depth` unset or `1`) returns it.
- [ ] Querying with `Depth: 2` against a 2-hop chain (A→B→C) returns C when starting
      from A; `Depth: 1` from the same starting point does not.
- [ ] Querying with `Depth` set higher than the package's max (e.g. `Depth: 100`)
      behaves identically to querying at the max — proving the clamp, not just that a
      large value doesn't crash.
- [ ] A query against a node with no edges returns an empty result, not an error.
- [ ] Real SQLite, no mocking. `go test -race ./internal/memory/graph/...` green.

## Relevant files / areas
- `internal/memory/graph/store.go` (created in issue 2 — this issue extends it, same
  file).
- `internal/memory/vector/store.go` — the existing `Limit`-clamping precedent to match
  in spirit (comment style, defensive-clamp-not-trust pattern).

## Dependencies
Blocked by [issue 2](./02-graph-store-write.md) (needs the schema and `Write` to seed
test data). Blocks [issue 4](./04-http-wiring-and-config.md).

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it further — e.g. single-hop
`Query` and the recursive multi-hop path could become two PRs if the recursive SQL
proves fiddly enough to want its own focused review.
