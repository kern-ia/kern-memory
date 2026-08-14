# Verification — end-to-end recall against the simulated content

This is the repeatable process for confirming that the memory layers actually recall
the content in `bank-criteria.json`. Re-run it unchanged whenever the content changes —
including once real AvelFinances criteria replace the simulated set — to confirm the
new content still recalls correctly. Everything below is a literal command run against
a real `kern-memory serve` instance and real Ollama embeddings — no mocking.

## Prerequisites

- Ollama running locally with the embedding model pulled:
  `ollama pull nomic-embed-text` (see `README.md`'s "Consumed — Ollama").
- A build of the binary: `go build -o bin/kern-memory ./cmd/kern-memory`.

## 1. Load the content before starting the daemon

`chromem-go` (the vector layer) loads its collection into memory once, at `Open` — it
does not live-watch the file. Load first, start the daemon second, so the content the
daemon serves is the content you just loaded.

```sh
rm -f /tmp/kern-memory-verify.db /tmp/kern-memory-verify-okf.db /tmp/kern-memory-verify-graph.db
rm -rf /tmp/kern-memory-verify-vector.db

export KERN_MEMORY_DB=/tmp/kern-memory-verify.db
export KERN_MEMORY_OKF_DB=/tmp/kern-memory-verify-okf.db
export KERN_MEMORY_VECTOR_DB=/tmp/kern-memory-verify-vector.db
export KERN_MEMORY_GRAPH_DB=/tmp/kern-memory-verify-graph.db

./bin/kern-memory load-memory docs/epics/epic-2-avelfinances-content-population/content/bank-criteria.json
# → loaded 10 memories and 4 edges
```

## 2. Start the daemon against the same store paths

```sh
export KERN_MEMORY_ADDR=127.0.0.1:7080
export KERN_MEMORY_TOKEN=verify-token-epic2
./bin/kern-memory serve &
```

If the daemon was already running when you loaded content in step 1 (or you're
re-verifying after an edit to `bank-criteria.json`), restart it now — same reason as
above, the vector layer only reads the file at `Open`.

## 3. Run prospect-style questions (semantic recall)

Each of these is a question an AvelFinances agent could plausibly field from a
prospect, run against `POST /api/v1/memory/query` with `kind: "vector"`. Below are the
runs from the 2026-08-11 verification, with real similarity scores — the matching
criterion should rank first, clearly ahead of unrelated content, the same bar phase 1's
own verification held (see `docs/index/0002-epic13-phase1.md`).

### Question 1 — age limit exception for a SCI-held loan

```sh
curl -s -X POST http://127.0.0.1:7080/api/v1/memory/query \
  -H "Authorization: Bearer verify-token-epic2" \
  -H "Content-Type: application/json" \
  -d '{"text": "Mon client a 72 ans et veut emprunter via une SCI, est-ce que la limite d'\''âge de 70 ans s'\''applique ?", "kind": "vector", "limit": 5}'
```

Top result: `criterion-age-senior-sci` at **0.803**, next relevant result
`criterion-age-senior-general` at 0.679, unrelated criteria (guarantee, income) at
0.60–0.63. Correct criterion recalled, clearly separated from the rest.

### Question 2 — minimum shareholders for a SCI

```sh
curl -s -X POST http://127.0.0.1:7080/api/v1/memory/query \
  -H "Authorization: Bearer verify-token-epic2" \
  -H "Content-Type: application/json" \
  -d '{"text": "Une SCI avec un seul associé peut-elle emprunter ?", "kind": "vector", "limit": 3}'
```

Top result: `criterion-sci-min-shareholders` at **0.766**, next result
`criterion-age-senior-sci` at 0.678. Correct criterion recalled.

### Question 3 — VEFA fund disbursement

```sh
curl -s -X POST http://127.0.0.1:7080/api/v1/memory/query \
  -H "Authorization: Bearer verify-token-epic2" \
  -H "Content-Type: application/json" \
  -d '{"text": "Comment sont débloqués les fonds pour un achat en VEFA ?", "kind": "vector", "limit": 3}'
```

Top result: `criterion-property-new-construction` at **0.730**, next result
`criterion-age-senior-general` at 0.599. Correct criterion recalled.

## 4. Run a graph traversal

For at least one pair of criteria connected by a graph edge (from `bank-criteria.json`'s
`edges`), confirm the relationship is recallable via `kind: "graph"`.

### Single-hop — age exception supersedes the general rule

```sh
curl -s -X POST http://127.0.0.1:7080/api/v1/memory/query \
  -H "Authorization: Bearer verify-token-epic2" \
  -H "Content-Type: application/json" \
  -d '{"kind": "graph", "from_kind": "vector", "from_id": "criterion-age-senior-sci", "depth": 1}'
```

Returns the `edge-age-sci-supersedes-age-general` edge
(`criterion-age-senior-sci` → `relation: supersedes-when-borrower-is-sci` →
`criterion-age-senior-general`) at `similarity: 1` (graph lookups are exact, not ranked).

### Multi-hop — guarantee tightens through the rental/self-employed chain

```sh
curl -s -X POST http://127.0.0.1:7080/api/v1/memory/query \
  -H "Authorization: Bearer verify-token-epic2" \
  -H "Content-Type: application/json" \
  -d '{"kind": "graph", "from_kind": "vector", "from_id": "criterion-guarantee-caution-vs-hypotheque", "depth": 2}'
```

Returns both edges in the chain: `criterion-guarantee-caution-vs-hypotheque` →
(`stricter-requirement-for`) → `criterion-property-rental-investment` →
(`combines-with`) → `criterion-income-self-employed`.

## 5. Stop the daemon and clean up

```sh
kill %1   # or the PID printed by `serve &`
rm -f /tmp/kern-memory-verify*.db
rm -rf /tmp/kern-memory-verify-vector.db
```

## 6. Full test suite

Not specific to this content, but run alongside any content-loading change as a
sanity check that nothing in the storage layers regressed:

```sh
go test -race -count=1 ./...
```

All 9 packages green as of this verification.
