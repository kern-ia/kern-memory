---
type: Concept
title: kern-memory
description: Persistent, agnostic memory brick for the Kern ecosystem — no kern-* brick had memory across runs before this.
tags: [concept, kern-memory, kern-ecosystem]
timestamp: 2026-08-09T00:00:00Z
---

# kern-memory — concept

## The problem

Before this repo existed, no brick in the Kern ecosystem had memory across runs. Every
run of `kern-orch` (and any other `kern-*` brick) started from nothing — nothing written
during one run was available to the next, nothing survived a rebuild. Whatever an agent
needed to know, it needed to be told again, every time.

`kern-memory` exists to close that gap: a place any `kern-*` brick can write something to
and later recall it, across runs, without carrying it itself.

## Who has the problem

Any `kern-*` brick that runs more than once and needs continuity — `kern-orch` first and
foremost, as the orchestrator whose agents are the ones re-deriving context on every run.

The first concrete instance of the problem was AvelFinances (a mortgage brokerage
client): their agents need to recall bank lending criteria across separate, unrelated
runs — a criterion learned or written once should be found again weeks later by an
unrelated query, not re-entered. This was both a pre-existing roadmap item (EPIC-13) and
a client-driven trigger that made building it now rather than later the obvious call —
the two reasons coincided, neither alone would be the full story.

## Why this, why now

`kern-memory` was scoped and built to serve that AvelFinances need directly (see EPIC-13
phase 1), but deliberately built as the general-purpose brick the roadmap already called
for, not a bespoke bank-criteria store. The alternative — a one-off store shaped only
around that first use case — would have left the second `kern-*` brick that needs memory
to either reinvent it or awkwardly repurpose something not meant to be general.

A non-negotiable constraint shaped the direction as much as the problem did:
**sovereignty**. AvelFinances was sold on a no-external-network-call argument for
anything touching their data. That ruled out any cloud-hosted or cloud-dependent memory
framework (Zep, for instance) regardless of how well it otherwise fit, and pushed the
embeddings choice to a locally-run model (Ollama) rather than an API call.

## What it is

A **write/query** memory store, callable by any `kern-*` brick, with a neutral contract
that doesn't assume what's being remembered:

- `write` — persist a piece of memory (free text, or a fact), pseudonymized before it
  touches disk (`kern-anon`, in-process, on by default) so what's retained at rest is
  never raw personal data.
- `query` — recall it later, by exact match on a tag or by semantic closeness to a
  question, from a caller with no relationship to whoever wrote it.

Two layers exist today, and the contract routes each request to the one it fits rather
than forcing every caller to pick a backend:

- **Declarative** (`.okf`, SQLite) — for facts with a natural lookup key: a rule, a
  criterion, an id. Recalled by exact tag match.
- **Semantic** (vector, `chromem-go` + local Ollama embeddings) — for free text with no
  natural key, recalled by meaning. This is the default when a caller doesn't specify
  which layer it wants — most memories don't have a natural lookup key.

A third layer — a **graph/relationship** layer, answering "what's connected to what, and
since when" (multi-hop reasoning, provenance, recency) — is a real, named need that
neither existing layer answers, and is deliberately not built yet.

## What it deliberately is not

- **Not a cloud service.** Every layer runs in-process or as a local dependency (Ollama).
  No memory ever leaves the machine it's written on to be stored or embedded.
- **Not a general PII vault.** Pseudonymization masks pattern-based entities (IBAN,
  email, phone, French NIR/SIREN/SIRET) and, when configured, person names — but never
  organizations or locations, because a bank's name is exactly the kind of thing this
  store exists to recall, not something to hide from itself.
- **Not a single-tenant memory per caller.** There's one bearer token today
  (`kern-ui` is presently the only caller of the unrelated document store in this same
  binary); `kern-memory`'s own memory API has no per-caller or per-session separation
  built in, unlike the producer/session split `kern-orch` needed for its two distinct
  caller populations. If a second population of memory callers with different trust
  levels shows up, that's a decision still to make, not one made here.
- **Not built around temperature/recency tiers.** Frameworks like Letta/MemGPT organize
  memory into hot/warm/cold-like paging (main/recall/archival). `kern-memory` considered
  and explicitly did not adopt that model — its layers are chosen by **query shape**
  (exact-key lookup vs. semantic recall vs., later, relationship traversal), not by how
  recently or often a memory is touched.

## What people do today instead

Nothing, inside the Kern ecosystem — this is the first attempt. Outside it, the
alternative is what any team defaults to without a memory brick: re-deriving context
by hand each run, or reaching for a cloud-hosted framework (Zep, Mem0's managed tier)
that would violate the sovereignty constraint outright.

## Open questions (inherited by `define-scope`)

- The graph/relationship layer: make (light graph on SQLite/Postgres) vs. adopt
  (Graphiti in-subprocess) — not decided.
- `kern-obs` (traced recall — who read/wrote what, when) is a named transverse need with
  no brick behind it anywhere in the ecosystem yet. Blocked on that brick existing, not
  forgotten.
- No per-caller/per-session trust separation exists yet in the memory API. Fine while
  `kern-orch` is the only caller; unresolved for a second one.
- Populating the AvelFinances semantic layer with real, current bank-criteria content is
  separate work from the brick itself, and untouched by this concept.
