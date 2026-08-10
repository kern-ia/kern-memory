---
type: Decision
title: "Background work"
description: "Async jobs, queues, or schedulers for the graph layer."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 13
slug: background-work
status: na
verdict: "None — synchronous request/response, same as both existing layers"
decided_via: na
depends_on: []
---

# Question
N/A — writes and multi-hop queries stay synchronous within the HTTP request, same as
`.okf` and vector today. Phase 2's scope ([decision](../scope/07-mvp-cut.md)) has no
workload that needs async processing.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A.
