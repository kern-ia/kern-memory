# kern-memory

Deux choses dans un binaire — voir `CLAUDE.md` pour le découpage complet :
- **C8 v1** : stockage minimal de documents/suggestions pour la vue Rédaction de kern-ui.
- **EPIC-13 phase 1** : la brique mémoire agnostique de `kern-orch` (`.okf` déclaratif +
  vecteur sémantique `chromem-go`/Ollama, pseudonymisation `kern-anon` par défaut).

## API — C8 v1 (documents)

```
GET  /api/v1/documents                          → liste (id, titre, nombre de mots, mise à jour)
GET  /api/v1/documents/{id}                      → document + suggestions
POST /api/v1/documents/{id}/suggestions/{sid}/accept
POST /api/v1/documents/{id}/suggestions/{sid}/ignore
```

## API — EPIC-13 phase 1 (mémoire)

```
POST /api/v1/memory/write   {"kind":"okf"|"vector","text":"...","tags":[...],"metadata":{...}}
POST /api/v1/memory/query   {"text":"...","kind":"okf"|"vector"|"","tags":[...],"limit":N}
```

`kind` vide sur `write` route vers la couche vecteur (défaut) ; sur `query`, interroge les
deux couches et fusionne par similarité décroissante. Le texte écrit est pseudonymisé
(kern-anon) avant persistance, sauf `KERN_MEMORY_PSEUDONYMIZE=false`.

Authentification par jeton porteur (`KERN_MEMORY_TOKEN`) sur les deux API, même appelant.

## Configuration

| Variable                   | Rôle                                          | Défaut                     |
|-----------------------------|------------------------------------------------|-----------------------------|
| `KERN_MEMORY_ADDR`          | Adresse d'écoute                                | `127.0.0.1:7080`            |
| `KERN_MEMORY_TOKEN`         | Jeton porteur exigé de l'appelant               | (aucun, dev local)          |
| `KERN_MEMORY_DB`            | Chemin du fichier SQLite (C8 v1)                | `kern-memory.db`            |
| `KERN_MEMORY_OKF_DB`        | Chemin du fichier SQLite (couche `.okf`)        | `kern-memory-okf.db`        |
| `KERN_MEMORY_VECTOR_DB`     | Chemin du dossier `chromem-go` (couche vecteur) | `kern-memory-vector.db`     |
| `KERN_MEMORY_OLLAMA_MODEL`  | Modèle d'embeddings Ollama                      | `nomic-embed-text`          |
| `KERN_MEMORY_OLLAMA_URL`    | URL de l'API Ollama                             | (défaut chromem-go, local)  |
| `KERN_MEMORY_PSEUDONYMIZE`  | Masquer le PII avant écriture                   | `true` (`false` désactive)  |

Le binaire refuse de démarrer sur une adresse publique sans `KERN_MEMORY_TOKEN`. La couche
vecteur nécessite un serveur Ollama local avec le modèle d'embeddings installé
(`ollama pull nomic-embed-text`).

## Lancer

```sh
go build -o bin/kern-memory ./cmd/kern-memory
KERN_MEMORY_TOKEN=... ./bin/kern-memory serve
```

## Semer un document

Seule voie d'entrée pour du contenu en V1 — pas de chemin HTTP d'écriture :

```sh
./bin/kern-memory seed "Titre du document" chemin/vers/corps.txt
```

Affiche l'id généré.

## Tests

```sh
go test ./...
```
