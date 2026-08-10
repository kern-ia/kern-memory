---
type: Decision
title: "Data migrations"
description: "How the graph schema rolls out."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 14
slug: data-migrations
status: na
verdict: "Additive CREATE TABLE IF NOT EXISTS, same as the .okf layer — no data to migrate yet"
decided_via: na
depends_on: []
---

# Question
N/A — the graph layer is new storage with no existing production data, so there's
nothing to migrate. Same schema-on-open pattern as `internal/memory/okf/store.go`
(`CREATE TABLE IF NOT EXISTS` run at `Open`), no separate migration tool.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A.
