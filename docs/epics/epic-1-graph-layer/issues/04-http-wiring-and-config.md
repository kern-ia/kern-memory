---
type: Issue
title: "Wire the graph store into HTTP, config, and main — verify end to end"
description: "JSON DTOs for the new fields, KERN_MEMORY_GRAPH_DB config, main.go wiring, and a real-daemon verification of Epic 1's acceptance criteria."
tags: [epic-1]
timestamp: 2026-08-10T11:00:00Z
epic: 1
issue: 4
slug: http-wiring-and-config
size: S
status: open
gh_issue: 9
resource: https://github.com/kern-ia/kern-memory/issues/9
depends_on: [graph-contract-and-routing, graph-store-write, graph-store-query-traversal]
---

# Wire the graph store into HTTP, config, and main — verify end to end

## Summary
The last piece: make the graph layer reachable over the existing
`POST /api/v1/memory/write` / `/query` endpoints, configurable, and wired into the
daemon — then verify Epic 1's acceptance criteria against a real running instance, the
same bar EPIC-13 phase 1 held (`docs/index/0002-epic13-phase1.md`'s own "Vérifié en
réel" section is the model to match).

## Scope
- `internal/httpapi/memory.go`: add `FromID`/`ToID`/`Relation`/`Depth` to
  `memoryWriteRequest`, `memoryQueryRequest`, `memoryDTO` (JSON in/out), passed through
  to `memory.Memory`/`memory.Query` — no new routes.
- `internal/config/config.go`: add `EnvGraphDB = "KERN_MEMORY_GRAPH_DB"` and
  `GraphDB` field, defaulting to `"kern-memory-graph.db"` — same pattern as
  `EnvOKFDB`/`EnvVectorDB`.
- `cmd/kern-memory/main.go`'s `openMemory`: open the graph store, set it on
  `Router.Graph`, close it alongside the other two stores.
- Manual verification against a real running `kern-memory serve`: write a graph edge
  between two real memories (one `.okf`, one vector) via `curl`, query it back
  single-hop and multi-hop, confirm the response — document the exact commands and
  output in the PR description, the way phase 1's retro documented its own `curl`
  verification.

## Out of scope
- Any new endpoint — the existing two carry the new fields.
- Populating real content — Epic 2.

## Acceptance criteria / Definition of done
- [ ] `POST /api/v1/memory/write` with `{"kind":"graph","from_id":"...","to_id":"...",
      "relation":"..."}` returns 200 and the edge is queryable.
- [ ] `POST /api/v1/memory/query` with `{"kind":"graph","depth":2,...}` returns a
      multi-hop result for a real 2-hop chain.
- [ ] `KERN_MEMORY_GRAPH_DB` overrides the graph store's file path; unset falls back to
      the default, mirroring `internal/config/config_test.go`'s existing coverage for
      the other two DB paths.
- [ ] `go test -race ./...` (whole module) green.
- [ ] Real daemon started, real `curl` calls made and their output pasted into the PR
      description — not asserted as done without the transcript, consistent with this
      repo's established verification standard.

## Relevant files / areas
- `internal/httpapi/memory.go` — DTOs and handlers (existing file, read in full).
- `internal/config/config.go` — env var pattern.
- `cmd/kern-memory/main.go` — `openMemory`.
- `internal/config/config_test.go` — existing test pattern for env var defaults/
  overrides, to extend.

## Dependencies
Blocked by [issue 1](./01-graph-contract-and-routing.md),
[issue 2](./02-graph-store-write.md), and
[issue 3](./03-graph-store-query-traversal.md) — this is the integration issue that
closes out Epic 1's acceptance criteria. Blocks nothing in this epic; unblocks
[Epic 2](../../epic-2-avelfinances-content-population/EPIC_2.md).

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it before opening the PR.
Expected small — most of the real work already happened in issues 1–3.
