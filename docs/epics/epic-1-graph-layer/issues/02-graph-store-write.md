---
type: Issue
title: "Add graph store: schema and Write"
description: "New internal/memory/graph package: SQLite schema for graph_edges and a Write implementation with Relation validation."
tags: [epic-1]
timestamp: 2026-08-10T11:00:00Z
epic: 1
issue: 2
slug: graph-store-write
size: M
status: open
gh_issue: 7
resource: https://github.com/kern-ia/kern-memory/issues/7
depends_on: [graph-contract-and-routing]
---

# Add graph store: schema and Write

## Summary
The graph layer's storage half of `memory.Store`. New package
`internal/memory/graph`, following `internal/memory/okf/store.go`'s own shape closely
(`Open`, `SetMaxOpenConns(1)` + `PRAGMA busy_timeout`, schema-on-open) —
[spec](../../../planning/specs/03-graph-storage-schema.md). `Write` inserts a directed
edge referencing existing `.okf`/vector memory IDs; it does not read them back or
validate they exist (the graph layer stores references, it doesn't own referential
integrity across layers it has no query access to).

## Scope
- `internal/memory/graph/store.go`: `Open(path string) (*Store, error)` creating the
  `graph_edges` table if needed:
  ```sql
  CREATE TABLE IF NOT EXISTS graph_edges (
      id         TEXT PRIMARY KEY,
      from_kind  TEXT NOT NULL,
      from_id    TEXT NOT NULL,
      to_kind    TEXT NOT NULL,
      to_id      TEXT NOT NULL,
      relation   TEXT NOT NULL,
      created_at TEXT NOT NULL
  );
  ```
- `Write(ctx, m memory.Memory) (memory.Memory, error)`: generates an id when unset
  (same `crypto/rand` 8-byte-hex pattern as `okf.newID`), validates `Relation`
  ([spec](../../../planning/specs/11-security-privacy.md)): non-empty, under a fixed
  length cap (e.g. 64 bytes), rejected if it contains a newline or exceeds the cap —
  "shaped like a label, not prose." Returns a clear validation error, not a DB error,
  when rejected.
- `Close() error`.

## Out of scope
- `Query` (single-hop or multi-hop) — issue 3.
- Wiring this store into `Router.Graph` or `cmd/kern-memory/main.go` — issue 4.
- Validating that `from_id`/`to_id` actually exist in `.okf`/vector — out of scope for
  the whole epic (`memory.Store` has no cross-layer read access; see EPIC_1.md Notes).

## Acceptance criteria / Definition of done
- [ ] `graph.Open` creates `graph_edges` if it doesn't exist, and is idempotent against
      a database that already has it (matches `okf.Open`'s test pattern).
- [ ] `Write` with `FromID`/`ToID`/`Relation` set persists a row; a re-read (direct SQL
      in the test, since `Query` doesn't exist yet) shows the correct columns.
- [ ] `Write` with an empty `Relation` returns an error, no row inserted.
- [ ] `Write` with a `Relation` longer than the cap returns an error, no row inserted.
- [ ] `Write` with a `Relation` containing a newline returns an error, no row inserted.
- [ ] Real SQLite (`modernc.org/sqlite`, no cgo) — no mocking the database, consistent
      with this repo's established practice.
- [ ] `go test -race ./internal/memory/graph/...` green.

## Relevant files / areas
- `internal/memory/okf/store.go` — the pattern to follow for `Open`/schema/id
  generation (busy_timeout, `SetMaxOpenConns(1)`, `crypto/rand` id).
- `internal/memory/contract.go` — the `memory.Memory`/`memory.Store` types this
  package implements (from issue 1).

## Dependencies
Blocked by [issue 1](./01-graph-contract-and-routing.md) (needs the `FromID`/`ToID`/
`Relation` fields on `memory.Memory`). Blocks [issue 3](./03-graph-store-query-traversal.md)
(needs the schema) and [issue 4](./04-http-wiring-and-config.md).

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it before opening the PR.
