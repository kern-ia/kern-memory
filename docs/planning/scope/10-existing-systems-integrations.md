---
type: Decision
title: "Existing systems of record & integrations"
description: "What already holds the truth phase 2 must respect."
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 10
slug: existing-systems-integrations
status: na
verdict: "Covered by decision 05 (graph-data-model)"
decided_via: na
depends_on: [graph-data-model]
---

# Question
N/A as a separate decision — the only existing systems of record phase 2 must respect
are the two memory layers it's built on top of (`.okf` SQLite store, `chromem-go`
vector store), plus `kern-anon` pseudonymization and Ollama embeddings. How the new
graph layer relates to them is exactly what decision 05 (graph-data-model) settles.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — folded into decision 05 rather than duplicated here.
