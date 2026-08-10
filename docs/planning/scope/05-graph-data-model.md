---
type: Decision
title: "Graph data model: reference vs. duplicate"
description: "Do graph nodes reference existing memory records, or store their own copy of content?"
tags: [decision, scope]
timestamp: 2026-08-10T00:00:00Z
phase: scope
decision: 5
slug: graph-data-model
status: decided
verdict: "A. Reference existing memory IDs"
decided_via: triage
depends_on: [graph-layer-approach]
---

# Question
Once the graph layer exists, how does it relate to the two memory layers already
storing content (`.okf`, vector)? This also folds in the checklist's "existing systems
of record" area — the graph must respect what already holds the truth.

# Options
- **A. Reference existing memory IDs.** Graph nodes are pointers into `.okf`/vector
  records already written; the graph itself stores only edges (relationship, since-when)
  — never content.
- **B. Duplicate content into graph nodes.** Graph stores its own copy of the text/facts
  it relates — decoupled from the other layers, but a second place the same
  pseudonymized content lives, with a real risk of drift between copies.

# Recommendation
**A.** The graph answers "what's connected to what, since when" — it doesn't need to
own content to do that. Referencing existing IDs avoids re-pseudonymizing and
duplicating data that's already correctly stored, and keeps a single source of truth
per memory.

# Verdict
Accepted at triage: A. Graph edges reference `.okf`/vector memory IDs; the graph never
stores its own copy of content.
