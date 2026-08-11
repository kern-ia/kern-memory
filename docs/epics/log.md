# Epics bundle — update log

## 2026-08-11 (7)
* **Reconcile**: Epic 1 complete. All 4 issues merged into `dev`:
  [#6](https://github.com/kern-ia/kern-memory/issues/6) via
  [PR #11](https://github.com/kern-ia/kern-memory/pull/11),
  [#7](https://github.com/kern-ia/kern-memory/issues/7) via
  [PR #12](https://github.com/kern-ia/kern-memory/pull/12),
  [#8](https://github.com/kern-ia/kern-memory/issues/8) via
  [PR #14](https://github.com/kern-ia/kern-memory/pull/14),
  [#9](https://github.com/kern-ia/kern-memory/issues/9) via
  [PR #15](https://github.com/kern-ia/kern-memory/pull/15).
  All GitHub issues closed. Bundle status flipped to `done`.

## 2026-08-11 (6)
* **Reconcile**: [implement-issue] Epic 1 issue 3 (graph store query/traversal,
  [#8](https://github.com/kern-ia/kern-memory/issues/8)) merged via
  [PR #14](https://github.com/kern-ia/kern-memory/pull/14) into `dev` — status flipped
  to `done`. Issue #8 already closed on GitHub.

## 2026-08-11 (5.5)
* **Reconcile**: [implement-issue] Epic 1 issue 4 (HTTP wiring & real-daemon
  verification, [#9](https://github.com/kern-ia/kern-memory/issues/9)) merged via
  [PR #15](https://github.com/kern-ia/kern-memory/pull/15) into `dev` — status flipped
  to `done`. Issue #9 already closed on GitHub.

## 2026-08-11 (5)
* **PR opened**: [implement-issue] Epic 1 issue 3 (graph store query/traversal,
  [#8](https://github.com/kern-ia/kern-memory/issues/8)) — `Query` on
  `internal/memory/graph`: `Depth <= 1` returns direct edges only, `Depth > 1`
  walks a `WITH RECURSIVE` traversal capped at a server-side `maxDepth` (3)
  regardless of caller-requested `Depth`. `memory.Query` gains `FromKind`/
  `FromID` to name the traversal's starting node. `go test -race
  ./internal/memory/graph/...` green (13/13, 5 new); clamp test verified to
  fail when the clamp is removed. [PR #14](https://github.com/kern-ia/kern-memory/pull/14)
  against `dev`.

## 2026-08-11 (5)
* **Reconcile**: [implement-issue] Epic 1 issue 2 (graph store write,
  [#7](https://github.com/kern-ia/kern-memory/issues/7)) merged via
  [PR #12](https://github.com/kern-ia/kern-memory/pull/12) into `dev` — status flipped
  to `done`. Issue #7 already closed on GitHub.

## 2026-08-11 (4)
* **PR opened**: [implement-issue] Epic 1 issue 2 (graph store write,
  [#7](https://github.com/kern-ia/kern-memory/issues/7)) — new
  `internal/memory/graph` package: `Open` creates `graph_edges`
  (idempotent), `Write` persists a directed edge with a generated id and
  validates `Relation` as a label (non-empty, ≤64 bytes, no newline).
  `go test -race ./internal/memory/graph/...` green (7/7).
  [PR #12](https://github.com/kern-ia/kern-memory/pull/12) against `dev`.

## 2026-08-11 (3)
* **Started**: [implement-issue] Epic 1 issue 2 (graph store write,
  [#7](https://github.com/kern-ia/kern-memory/issues/7)) — branch
  `issue-7-graph-store-write`.

## 2026-08-11 (2)
* **Reconcile**: [implement-issue] Epic 1 issue 1 (contract & routing,
  [#6](https://github.com/kern-ia/kern-memory/issues/6)) merged via
  [PR #11](https://github.com/kern-ia/kern-memory/pull/11) into `dev` — status flipped
  to `done`. Issue #6 already closed on GitHub.

## 2026-08-11
* **PR opened**: [implement-issue] Epic 1 issue 1 (contract & routing,
  [#6](https://github.com/kern-ia/kern-memory/issues/6)) — Router gains a `Graph`
  field and `KindGraph` routing on `Write`/`Query`; empty-`Kind` fan-out extends to
  three layers when `Graph` is configured, stays two-layer (backward compatible)
  when it isn't. `go test -race ./internal/memory/...` green (14/14).
  [PR #11](https://github.com/kern-ia/kern-memory/pull/11) against `dev`.

## 2026-08-10 (2)
* **Creation**: [create-issues] Epic 1 (Graph layer) broken into 4 issues:
  [#6](https://github.com/kern-ia/kern-memory/issues/6) contract & routing (S),
  [#7](https://github.com/kern-ia/kern-memory/issues/7) store write (M),
  [#8](https://github.com/kern-ia/kern-memory/issues/8) store query/traversal (M),
  [#9](https://github.com/kern-ia/kern-memory/issues/9) HTTP wiring & real-daemon
  verification (S). All linked as sub-issues of Epic 1
  ([#3](https://github.com/kern-ia/kern-memory/issues/3)).

## 2026-08-10
* **Creation**: Established [Epic 1: Graph layer](/epic-1-graph-layer/EPIC_1.md)
  (issue [#3](https://github.com/kern-ia/kern-memory/issues/3), milestone 1) and
  [Epic 2: AvelFinances content population](/epic-2-avelfinances-content-population/EPIC_2.md)
  (issue [#4](https://github.com/kern-ia/kern-memory/issues/4), milestone 2), split
  from `docs/planning/SCOPE.md`'s two milestones.
