# kern-memory

**kern-memory is where other `kern-*` bricks put things they need to remember.**

One binary, two things it stores today — see [CLAUDE.md](CLAUDE.md) for why they share a
repo:

- **Documents and suggestions** — a minimal store behind `kern-ui`'s Rédaction view: a
  document's text, and suggested edits a person accepts or ignores.
- **Memory** — a general write/query store for any brick that needs to recall something
  across runs: a declarative layer (tagged facts, exact lookup) and a semantic layer
  (free text, recalled by meaning).

Both live in the same process, each with its own storage (SQLite for documents and for
the declarative layer, `chromem-go`'s own files for the semantic layer) — none of the
three reads another's data.

## How it fits into the ecosystem

```
kern-ui  ──── documents & suggestions ───>  kern-memory  <──── kern-anon (Go library,
                                                  │               pseudonymizes on
kern-orch ─── memory write/query ────────────────┘               write, in-process)
                                                  │
                                                  └────>  Ollama (embeddings, local)
```

kern-memory calls no other `kern-*` brick over the network. `kern-anon` is a Go module
dependency (in-process, no HTTP), not a service call; Ollama is the one real outbound
network dependency, and only for the semantic layer.

## What's built

- **Documents and suggestions**: read a document and its suggestions, accept or ignore
  one. The only way content enters is a CLI command (`seed`, below) — there is no HTTP
  path to write a document or generate a suggestion yet. See "What's coming".
- **Declarative memory**: write a fact with tags, recall it by an exact tag match
  (`similarity` always `1`). Meant for things with a natural lookup key — a rule, a
  criterion, an id.
- **Semantic memory**: write free text, recall it by meaning through a local embedding
  model (Ollama) and a cosine-similarity search (`chromem-go`, embedded — no separate
  vector database service to run).
- **Pseudonymization on write**: text is masked (`kern-anon`) before either memory layer
  persists it, on by default. Pattern-based entities (IBAN, email, phone, French
  NIR/SIREN/SIRET…) are always covered; person names are covered too when a downloaded
  ONNX model is configured (opt-in, see Configuration) — organizations and locations are
  deliberately never masked, because a bank's name is exactly the kind of thing this store
  exists to recall, not PII to hide from itself.
- **Editable content loading.** A JSON file holding a batch of memories and graph edges,
  loaded with `load-memory` (below) — the same directness as `seed`, no HTTP round trip,
  so content can be reviewed and re-loaded by editing a file rather than through
  engineering-only tooling.

## What's coming

Stated plainly rather than by an external roadmap's numbering, so this file stays readable
on its own:

- **Suggestion generation.** Nothing today turns a document's text into a suggestion —
  that loop (read → suggest → the person decides) has only its second half built.
- **A write path for documents.** Content only enters through `seed`; there is no
  HTTP equivalent yet, unlike the memory API's `write`.
- **A graph/relationship layer.** Today's two memory layers answer "what fact matches
  this tag" and "what text is close in meaning to this" — neither answers "what is
  connected to what, and since when" (multi-hop reasoning, provenance, recency). Real
  need, not yet designed.
- **Traced recall.** Nothing records *when* or *by whom* a memory was written or recalled.
  A brick for that (audit/observability) does not exist yet anywhere in the ecosystem —
  this is blocked on something outside this repo, not forgotten.
- **Bank-criteria content itself.** The semantic layer works end to end (verified with
  real embeddings and a real query); what it actually knows depends entirely on what gets
  written into it. Populating it with real, current content is separate work from the
  brick that stores it.

## Connection contracts

### Exposed — documents and suggestions

Consumed today by [`kern-ui`](../Kern-UI/README.md)'s Rédaction view, the same way it would
be consumed by any caller: a bearer token, JSON in and out.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/documents` | Every document's summary. |
| `GET` | `/api/v1/documents/{id}` | One document with its suggestions. `404` on an unknown id. |
| `POST` | `/api/v1/documents/{id}/suggestions/{sid}/accept` | Marks a suggestion accepted. |
| `POST` | `/api/v1/documents/{id}/suggestions/{sid}/ignore` | Marks a suggestion ignored. |

```json
// GET /api/v1/documents
[{ "id": "doc1", "title": "Compte-rendu", "word_count": 240, "updated_at": "2026-07-30T10:00:00Z" }]

// GET /api/v1/documents/doc1
{
  "id": "doc1", "title": "Compte-rendu", "body": "...", "word_count": 240,
  "updated_at": "2026-07-30T10:00:00Z",
  "suggestions": [
    { "id": "s1", "anchor_start": 12, "anchor_end": 30, "text": "...", "status": "pending" }
  ]
}
```

### Exposed — memory

Intended for `kern-orch` (or any `kern-*` brick that needs to remember something across
runs) — the same bearer-token convention as above, on the same daemon.

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/memory/write` | Upserts one memory. |
| `POST` | `/api/v1/memory/query` | Recalls memories matching a query. |

```json
// POST /api/v1/memory/write
{"kind": "okf", "text": "Le taux d'usure T3 2026 est 5,92%.", "tags": ["taux-usure"]}
// → 200 { "id": "7a29f286fcdfa7e2", "kind": "okf", "text": "...", "tags": ["taux-usure"] }

// POST /api/v1/memory/query
{"text": "quelle banque accepte une SCI ?", "limit": 5}
// → 200 [{ "memory": {"id": "...", "kind": "vector", "text": "..."}, "similarity": 0.83 }]
```

| Field | Meaning |
|---|---|
| `kind` | `"okf"` (declarative) or `"vector"` (semantic). Empty on write defaults to `"vector"` — most memory has no natural lookup key. Empty on query fans out to both layers, merged by `similarity` descending. |
| `id` | Stable caller-chosen key upserts (overwrites) rather than erroring on a repeat — a caller with a natural key (e.g. a run id) does not have to track existence first. Omitted, a random id is generated. |
| `tags` | Declarative layer: a query with `tags` returns only memories carrying **all** of them. Ignored by the semantic layer. |
| `text` | Pseudonymized (`kern-anon`) before it reaches either layer, unless `KERN_MEMORY_PSEUDONYMIZE=false`. What is written is what stays at rest — there is no round-trip demasking; see `internal/memory/anon`'s own doc for why that differs from `kern-orch`'s courtage-extraction pipeline. |

**What deliberately does not travel**: embeddings themselves (an implementation detail of
the semantic layer, not part of the contract — a caller never sees a raw vector).
`metadata` is free-form and passed through verbatim; kern-memory does not interpret it.

### Consumed — Ollama (embeddings)

The semantic layer calls a local Ollama server's embedding API
(`KERN_MEMORY_OLLAMA_URL`, default `http://localhost:11434`) with `KERN_MEMORY_OLLAMA_MODEL`
(default `nomic-embed-text`) — the one real outbound network call this brick makes, and it
never leaves the machine it runs on. No embedding call is ever made to an external/cloud
API; this is a deliberate sovereignty choice, not a limitation to work around later.

## Configuration

| Variable | Role | Default |
|---|---|---|
| `KERN_MEMORY_ADDR` | Listen address | `127.0.0.1:7080` |
| `KERN_MEMORY_TOKEN` | Bearer token required of every caller | (none, local dev) |
| `KERN_MEMORY_DB` | SQLite file (documents/suggestions) | `kern-memory.db` |
| `KERN_MEMORY_OKF_DB` | SQLite file (declarative memory) | `kern-memory-okf.db` |
| `KERN_MEMORY_VECTOR_DB` | `chromem-go` directory (semantic memory) | `kern-memory-vector.db` |
| `KERN_MEMORY_OLLAMA_MODEL` | Ollama embedding model | `nomic-embed-text` |
| `KERN_MEMORY_OLLAMA_URL` | Ollama API base URL | `chromem-go`'s own default, local |
| `KERN_MEMORY_PSEUDONYMIZE` | Mask PII before writing | `true` (`false` disables) |
| `KERN_ANON_NER_MODEL_DIR` | Path to a downloaded ONNX NER model — enables person-name masking | (unset, NER off) |

The binary refuses to start on a public address without `KERN_MEMORY_TOKEN`. The semantic
layer needs a local Ollama server with the embedding model installed
(`ollama pull nomic-embed-text`).

## Run

```sh
go build -o bin/kern-memory ./cmd/kern-memory
KERN_MEMORY_TOKEN=... ./bin/kern-memory serve
```

## Seed a document

The only way content enters the document store — see "What's coming" for why there is no
HTTP equivalent yet:

```sh
./bin/kern-memory seed "Document title" path/to/body.txt
```

Prints the generated id.

## Load memory from a file

The editable way to update memory content without an HTTP write path: edit a JSON file,
re-run the command. Same category as `seed` (direct store calls, no HTTP round trip).

```sh
./bin/kern-memory load-memory path/to/content.json
```

File format — a batch of memories and the graph edges between them:

```json
{
  "memories": [
    {"id": "criterion-sci-senior", "kind": "vector", "text": "...", "tags": ["bank-criteria"], "metadata": {}}
  ],
  "edges": [
    {"from_kind": "vector", "from_id": "criterion-a", "to_kind": "vector", "to_id": "criterion-b", "relation": "supersedes"}
  ]
}
```

`kind` per memory entry is `"okf"` or `"vector"` (empty defaults to `"vector"`, same as the
HTTP write path). `id` is optional on both memories and edges; when given, all three layers
(`.okf`, vector, and graph) upsert on a repeat id (fixed in #27 — `graph_edges` previously had
no `ON CONFLICT` clause). An id left empty gets a fresh one generated on every write, on every
layer — upserting on re-run requires a stable, caller-chosen id, consistently across all
three, not something re-running the loader on an unchanged file gets you for free.

## Tests

```sh
go test ./...                # default build — no cgo, no NER
CGO_ENABLED=1 go test -tags onnx ./...   # includes the ONNX NER integration test,
                                          # skipped unless KERN_ANON_NER_MODEL_DIR is set
```

## License

MIT — see [LICENSE](LICENSE).
