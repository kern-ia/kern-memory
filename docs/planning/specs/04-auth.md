---
type: Decision
title: "Auth & authorization"
description: "Who can call the graph endpoints and how."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 4
slug: auth
status: na
verdict: "Single bearer token (KERN_MEMORY_TOKEN), unchanged"
decided_via: na
depends_on: []
---

# Question
N/A — not reopened. `kern-orch` remains the only caller for phase 2
([decision](../scope/02-target-users.md)); the graph endpoint(s) sit behind the same
constant-time bearer-token check already in `internal/httpapi/memory.go`.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — inherited unchanged.
