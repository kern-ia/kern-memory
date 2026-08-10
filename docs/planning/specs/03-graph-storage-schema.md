---
type: Decision
title: "Graph storage schema"
description: "How graph edges are stored in SQLite, and how they reference memories across the .okf and vector layers."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 3
slug: graph-storage-schema
status: decided
verdict: "A. Single flat graph_edges table, (kind, id) composite reference"
decided_via: triage
depends_on: []
---

# Question
The graph layer must store directed relationships between existing memories
([decision](../scope/05-graph-data-model.md): edges reference existing `.okf`/vector
IDs, never duplicate content). `.okf` and vector each generate their own IDs
independently (`internal/memory/okf/store.go`'s `newID()` is a random 8-byte hex, local
to that store) — there's no shared ID namespace across layers, so an edge must
disambiguate which layer an ID belongs to. What's the schema?

# Options
- **A. Single flat `graph_edges` table**, `(id, from_kind, from_id, to_kind, to_id,
  relation, created_at)` — `(kind, id)` pairs disambiguate the cross-layer reference.
  Directed edges only; a caller needing bidirectionality writes two edges. Matches the
  `.okf` layer's own flat-table simplicity (`store.go`'s single `okf_memories` table,
  no joins).
- **B. Normalized `graph_nodes` + `graph_edges`.** Adds a node table so edges reference
  a local node id instead of repeating `(kind, id)` pairs — more relational purity, one
  more join, no clear benefit at phase-2's expected edge volume.
- **C. Edges as a JSON blob column on the existing memory tables**, no new table.
  Cheapest to add, but forces traversal logic into Go (scan + parse JSON per hop)
  instead of an indexed SQL query — defeats the reason to put a graph layer on SQLite
  at all.

# Recommendation
**A.** Consistent with the existing `.okf` store's own flat, unnormalized table, and
keeps multi-hop traversal expressible as a single indexed SQL query (SQLite supports
`WITH RECURSIVE`) rather than N+1 round trips or in-Go graph walking.

# Verdict
Accepted at triage: A.
