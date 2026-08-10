---
type: Decision
title: "Testing infrastructure"
description: "Frameworks and harness for the graph layer."
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 8
slug: testing-infrastructure
status: na
verdict: "go test -race ./..., real-backend integration tests, no mocks — unchanged"
decided_via: na
depends_on: []
---

# Question
N/A — not reopened. Same TDD/`go test` method as every prior feature in this repo
(`CLAUDE.md`), same "verified in real" bar EPIC-13 phase 1 held (real SQLite, real
Ollama, no mocked embeddings) — applied to the new graph package the same way.

# Options
(not applicable)

# Recommendation
(not applicable)

# Verdict
N/A — inherited unchanged.
