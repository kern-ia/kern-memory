---
type: Decision
title: "Graph layer: make vs. adopt"
description: "Build a light graph in Go, or adopt an existing graph/temporal-KG engine?"
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 4
slug: graph-layer-approach
status: decided
verdict: "A. Make - light graph on SQLite"
decided_via: triage
depends_on: [problem-scope]
---

# Question
`docs/kern-memory-etat-de-lart.md` left this explicitly "à trancher": make a light
graph/DAG in Go on top of existing storage, or adopt an external graph engine
(Graphiti, Neo4j)?

# Options
- **A. Make — light graph on SQLite.** Go-native, embedded, zero new ops — same pattern
  as the `.okf` layer already shipped.
- **B. Adopt Graphiti in-subprocess.** Best-in-class temporal reasoning (per the
  état-de-l'art research), but a Python subprocess dependency — breaks the 100%-Go-native
  single-binary direction already decided 2026-07-22.
- **C. Adopt Neo4j via Go driver.** Mature, capable graph tooling, but a service to run
  — breaks the "zero external service" pattern chromem-go and SQLite both follow.

# Recommendation
**A.** Both B and C reopen a direction already tranched in the état-de-l'art decision
(100% Go-native, embedded, no new service) for gains (Graphiti's temporal modeling,
Neo4j's scale) not needed at today's data volume. Revisit only if a real multi-hop
workload later proves a light graph insufficient.

# Verdict
Accepted at triage: A. Embedded Go/SQLite graph, consistent with the existing
single-binary, no-external-service direction.
