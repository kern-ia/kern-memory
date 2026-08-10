---
type: Decision
title: "Security & privacy — pseudonymization on graph writes"
description: "Does kern-anon pseudonymization apply to anything the graph layer writes?"
tags: [decision, specs]
timestamp: 2026-08-10T00:00:00Z
phase: specs
decision: 11
slug: security-privacy
status: decided
verdict: "A. No pseudonymization on Relation; validate it as a constrained label instead"
decided_via: triage
depends_on: [graph-storage-schema, graph-interface-contract]
---

# Question
`.okf` and vector both pseudonymize `Text` before persisting. A graph edge
(decision 05: `FromID`/`ToID`/`Relation`) carries no `Text` of its own — the memories it
connects are already pseudonymized by their own layer at write time. Does the graph
layer need its own pseudonymization step, or a different safeguard?

# Options
- **A. No pseudonymization on the edge itself; validate `Relation` as a constrained
  label instead** (e.g. length cap, rejects anything shaped like free prose). The
  memories the edge references were already pseudonymized when *they* were written; the
  edge adds no new prose surface, so there's nothing for the pseudonymizer to act on —
  the actual risk is a caller stuffing sentence-length text into `Relation` where a
  short label belongs, which validation catches and pseudonymization does not.
- **B. Run `Relation` through kern-anon defensively anyway**, in case a caller ignores
  the intended shape and writes prose there.

# Recommendation
**A.** `Relation` is a label, not prose — the right tool for "this field must stay a
label" is validation, not the pseudonymizer built for scanning free text. Running
kern-anon over a field that's not supposed to carry sensitive content adds cost without
addressing the actual failure mode (a caller misusing the field), which validation
catches directly.

# Verdict
Accepted at triage: A.
