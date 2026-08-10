---
type: Decision
title: "Delivery form"
description: "What does the caller actually receive?"
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 3
slug: delivery-form
status: na
verdict: "HTTP API in the existing kern-memory binary, unchanged"
decided_via: na
depends_on: []
---

# Question
N/A for phase 2 — already fixed by phase 1 and not reopened. `kern-memory` is an HTTP
API (`POST /api/v1/memory/write`, `/query`) served from the same single binary as the
document store. Phase 2 extends this contract; it does not change the delivery form.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — inherited unchanged from phase 1.
