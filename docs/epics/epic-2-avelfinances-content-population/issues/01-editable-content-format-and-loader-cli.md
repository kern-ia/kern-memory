---
type: Issue
title: "Add an editable content format and a load-memory CLI command"
description: "A simple JSON file format for memories + graph edges, and a CLI loader mirroring the existing seed command, so criteria content can be updated by editing a file and re-running one command."
tags: [epic-2]
timestamp: 2026-08-11T21:47:26Z
epic: 2
issue: 1
slug: editable-content-format-and-loader-cli
size: M
status: done
gh_issue: 19
gh_pr: 23
resource: https://github.com/kern-ia/kern-memory/issues/19
depends_on: []
---

# Add an editable content format and a load-memory CLI command

## Summary
AvelFinances' real bank-lending criteria are confidential and not yet available —
this epic populates simulated content now, but the actual point (per the user) is
that the content zone must be **easily editable**, so the agency (or whoever manages
this later) can update it without engineering once they have their own tooling. This
issue builds that mechanism: a plain JSON file format holding a batch of memories
and graph edges, and a `kern-memory load-memory <file>` CLI command that writes them
— mirroring `runSeed`'s directness in `cmd/kern-memory/main.go` (calls the store
directly, no HTTP round trip), not a new pattern.

## Scope
- A JSON file format, e.g.:
  ```json
  {
    "memories": [
      {"id": "criterion-sci-senior", "kind": "vector", "text": "...", "tags": ["bank-criteria"], "metadata": {}}
    ],
    "edges": [
      {"from_kind": "vector", "from_id": "criterion-a", "to_kind": "vector", "to_id": "criterion-b", "relation": "supersedes"}
    ]
  }
  ```
  `kind` per memory entry is `"okf"` or `"vector"` (matches `memory.Kind`); `edges`
  entries match the graph layer's `FromKind`/`FromID`/`ToKind`/`ToID`/`Relation`
  fields exactly ([spec](../../../planning/specs/05-graph-interface-contract.md)).
- `kern-memory load-memory <file>` in `cmd/kern-memory/main.go`: parses the file,
  opens the same `memory.Router` `runServe`/`openMemory` already builds, writes each
  `memories` entry via `Router.Write` with `Kind` set from the entry, then each
  `edges` entry via `Router.Write` with `Kind: memory.KindGraph`.
- Re-running the loader on a file with the same `id`s must **upsert**, not
  duplicate — verify each layer's `Write` actually does this (`.okf` already does,
  per `ON CONFLICT(id) DO UPDATE` in `internal/memory/okf/store.go`; confirm vector
  and graph behave sanely on a repeat `id`, and note in the PR if either doesn't
  upsert cleanly — that's a real finding, not something to paper over).
- Update `README.md`'s "What's built"/command list to document the new command,
  matching how `seed` is already documented there.

## Out of scope
- Any actual bank-criteria content — issue 2.
- An HTTP endpoint for bulk loading — this is a CLI tool, same category as `seed`
  (no HTTP write path for documents either, by design, per `CLAUDE.md`).
- Validating file content against a schema beyond basic JSON parsing and the
  existing per-field validation each store already does (e.g. graph's `Relation`
  length cap) — reuse what exists, don't build new validation.

## Acceptance criteria / Definition of done
- [ ] `kern-memory load-memory <file>` with a file containing 2 memories + 1 edge
      writes all three, verified by querying them back.
- [ ] Re-running the same command with the same file does not create duplicates
      (upsert, or a documented exception if one layer can't).
- [ ] A malformed JSON file produces a clear error, not a panic.
- [ ] `go test -race ./...` green; new tests for the parsing/loading logic (pure
      logic, testable without a real daemon per this repo's TDD convention —
      `CONVENTIONS.md`'s Testing section).
- [ ] `README.md` updated.

## Relevant files / areas
- `cmd/kern-memory/main.go` — `runSeed`, `openMemory` (existing patterns to follow).
- `internal/memory/contract.go` — `memory.Memory`, `memory.Kind`, `Router`.
- `README.md` — command documentation to extend.

## Dependencies
None — first issue in this epic.

## PR size note
Target ~500 changed lines; if this grows past ~1000, split it before opening the PR.
