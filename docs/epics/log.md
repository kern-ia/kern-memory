# Epics bundle — update log

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
