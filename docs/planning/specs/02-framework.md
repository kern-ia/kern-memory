---
type: Decision
title: "Framework"
description: "Application framework, if any."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 2
slug: framework
status: na
verdict: "net/http stdlib only, no framework"
decided_via: na
depends_on: []
---

# Question
N/A — not reopened. `internal/httpapi` uses `net/http`'s stdlib `ServeMux`
(`memory.go`), no framework, for both existing memory endpoints. The graph endpoint(s)
follow the same pattern.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — inherited unchanged.
