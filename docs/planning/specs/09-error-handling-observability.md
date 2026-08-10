---
type: Decision
title: "Error handling & observability"
description: "Logging, error reporting, and monitoring for the graph layer."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 9
slug: error-handling-observability
status: na
verdict: "fmt.Errorf wrapping, no observability layer — unchanged, kern-obs stays deferred"
decided_via: na
depends_on: []
---

# Question
N/A — not reopened. Same `fmt.Errorf("<pkg>: <op>: %w", err)` wrapping pattern used
throughout `internal/memory/*`. Traced recall (`kern-obs`) remains an explicit
non-goal for phase 2 ([decision](../scope/08-non-goals.md)) — there is still no
observability brick anywhere in the ecosystem to integrate with.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — inherited unchanged.
