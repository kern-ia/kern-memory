# kern-memory

**One binary, two contracts** — see [CLAUDE.md](CLAUDE.md) for the split in full:

- **C8 v1** — a minimal document/suggestion store for `kern-ui`'s Rédaction view.
- **EPIC-13 phase 1** — `kern-orch`'s agnostic memory brick: a declarative `.okf` layer
  plus a semantic vector layer (`chromem-go` + Ollama embeddings), with `kern-anon`
  pseudonymization on write by default.

The two coexist in the same process, each with its own storage (SQLite for C8 and `.okf`,
`chromem-go`'s own files for the vector layer) — none of the three reads another's data.

## How it fits into the ecosystem

```
kern-ui  ──── C8 (documents/suggestions) ────>  kern-memory  <──── kern-anon (Go library,
                                                      │               pseudonymizes on
kern-orch ─── EPIC-13 (memory write/query) ──────────┘               write, in-process)
                                                      │
                                                      └────>  Ollama (embeddings, local)
```

kern-memory calls no other `kern-*` brick over the network. `kern-anon` is a Go module
dependency (in-process, no HTTP), not a service call; Ollama is the one real outbound
network dependency, and only for the vector layer.

## Connection contracts

### Exposed — C8 v1 (documents and suggestions)

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

**What v1 deliberately does not do**: generate suggestions, or accept content over HTTP.
The only way content enters the store is the `seed` CLI command — see below — exactly like
`kern-ui useradd` is the only way an account is created. There is no path yet from "a
document changed" to "a new suggestion appeared"; that is future work, not an oversight.

### Exposed — EPIC-13 phase 1 (memory)

Intended for `kern-orch` (or any `kern-*` brick that needs to remember something across
runs) — the same bearer-token convention as C8, on the same daemon.

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
| `kind` | `"okf"` (declarative, tag-lookup, `similarity` always `1`) or `"vector"` (semantic, `chromem-go`). Empty on write defaults to `"vector"` — most memory has no natural lookup key. Empty on query fans out to both layers, merged by `similarity` descending. |
| `id` | Stable caller-chosen key upserts (overwrites) rather than erroring on a repeat — a caller with a natural key (e.g. a run id) does not have to track existence first. Omitted, a random id is generated. |
| `tags` | `.okf` layer: a query with `tags` returns only memories carrying **all** of them. Ignored by the vector layer. |
| `text` | Pseudonymized (`kern-anon`) before it reaches either layer, unless `KERN_MEMORY_PSEUDONYMIZE=false`. What is written is what stays at rest — there is no round-trip demasking; see `internal/memory/anon`'s own doc for why that differs from `kern-orch`'s courtage-extraction pipeline. |

**Person names are masked only when `kern-anon`'s ONNX NER engine is configured**
(`KERN_ANON_NER_MODEL_DIR`, a build with `-tags onnx`) — the default build masks only
pattern-based entities (IBAN, email, phone, French NIR/SIREN/SIRET…). Organization and
location names are never masked even with NER enabled: a bank's name is exactly the kind
of thing this store is meant to recall, not PII to hide from itself.

**What deliberately does not travel**: embeddings themselves (an implementation detail of
the vector layer, not part of the contract — a caller never sees a raw vector). `metadata`
is free-form and passed through verbatim; kern-memory does not interpret it.

### Consumed — Ollama (embeddings)

The vector layer calls a local Ollama server's embedding API
(`KERN_MEMORY_OLLAMA_URL`, default `http://localhost:11434`) with `KERN_MEMORY_OLLAMA_MODEL`
(default `nomic-embed-text`) — the one real outbound network call this brick makes, and it
never leaves the machine it runs on. No embedding call is ever made to an external/cloud
API; this is a deliberate sovereignty choice, not a limitation to work around later.

## Configuration

| Variable | Role | Default |
|---|---|---|
| `KERN_MEMORY_ADDR` | Listen address | `127.0.0.1:7080` |
| `KERN_MEMORY_TOKEN` | Bearer token required of every caller | (none, local dev) |
| `KERN_MEMORY_DB` | SQLite file (C8 v1) | `kern-memory.db` |
| `KERN_MEMORY_OKF_DB` | SQLite file (`.okf` layer) | `kern-memory-okf.db` |
| `KERN_MEMORY_VECTOR_DB` | `chromem-go` directory (vector layer) | `kern-memory-vector.db` |
| `KERN_MEMORY_OLLAMA_MODEL` | Ollama embedding model | `nomic-embed-text` |
| `KERN_MEMORY_OLLAMA_URL` | Ollama API base URL | `chromem-go`'s own default, local |
| `KERN_MEMORY_PSEUDONYMIZE` | Mask PII before writing | `true` (`false` disables) |
| `KERN_ANON_NER_MODEL_DIR` | Path to a downloaded ONNX NER model — enables person-name masking | (unset, NER off) |

The binary refuses to start on a public address without `KERN_MEMORY_TOKEN`. The vector
layer needs a local Ollama server with the embedding model installed
(`ollama pull nomic-embed-text`).

## Run

```sh
go build -o bin/kern-memory ./cmd/kern-memory
KERN_MEMORY_TOKEN=... ./bin/kern-memory serve
```

## Seed a document (C8 v1)

The only way content enters the C8 v1 store — no HTTP write path exists for it:

```sh
./bin/kern-memory seed "Document title" path/to/body.txt
```

Prints the generated id.

## Tests

```sh
go test ./...                # default build — no cgo, no NER
CGO_ENABLED=1 go test -tags onnx ./...   # includes the ONNX NER integration test,
                                          # skipped unless KERN_ANON_NER_MODEL_DIR is set
```
