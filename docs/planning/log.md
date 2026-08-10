# Planning bundle — update log

## 2026-08-10
* **Creation**: [SCOPE.md](SCOPE.md) for kern-memory phase 2, via the `scope/` decision
  ledger (13 decisions: 10 decided at triage, 3 n/a). Scopes the graph/relationship
  layer (make, embedded Go/SQLite, edges referencing existing `.okf`/vector memory IDs)
  and AvelFinances content population as two independently shippable milestones.
  Deferred as non-goals: `kern-obs`, per-caller trust separation, cross-layer fan-out
  query, vector-store swap, hot/warm/cold tiers.

## 2026-08-09
* **Creation**: [CONCEPT.md](CONCEPT.md) written retroactively from the already-shipped
  repo (C8 v1 store + EPIC-13 phase 1 memory brick) to give `define-scope` a validated
  intake: the problem (no `kern-*` brick had cross-run memory), who has it (`kern-orch`,
  triggered concretely by AvelFinances), why now (roadmap item + client trigger
  coinciding), and the tranched direction (neutral write/query contract routed by query
  shape — `.okf` declarative, semantic vector, graph layer named but not built).
  Explicitly ruled out during the conversation: framing the layers as hot/warm/cold
  temperature tiers (a candidate model from the état-de-l'art research, never the one
  decided or built).
