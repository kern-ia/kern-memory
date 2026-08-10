---
type: Decision
title: "Configuration & secrets"
description: "How the graph layer's config reaches the app."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 12
slug: configuration-secrets
status: na
verdict: "New KERN_MEMORY_GRAPH_DB env var, same pattern as EnvOKFDB/EnvVectorDB"
decided_via: na
depends_on: []
---

# Question
N/A as a new decision — `internal/config/config.go` already establishes the pattern
(one env var per storage file, a local-development default beside the binary). The
graph layer's DB path follows it exactly (`KERN_MEMORY_GRAPH_DB`, defaulting to
`kern-memory-graph.db`) rather than introducing a new configuration mechanism.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — repeats an established pattern, not a new decision.
