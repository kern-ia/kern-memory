# kern-memory

Stockage minimal de documents et de suggestions pour la vue Rédaction de kern-ui (C8 v1).
Ce n'est pas le `kern-memory` de l'EPIC-13 de `kern-orch` (RAG, embeddings, `.okf`) — voir
`CLAUDE.md` pour le découpage.

## API

```
GET  /api/v1/documents                          → liste (id, titre, nombre de mots, mise à jour)
GET  /api/v1/documents/{id}                      → document + suggestions
POST /api/v1/documents/{id}/suggestions/{sid}/accept
POST /api/v1/documents/{id}/suggestions/{sid}/ignore
```

Authentification par jeton porteur (`KERN_MEMORY_TOKEN`), kern-ui est l'unique appelant.

## Configuration

| Variable            | Rôle                                              | Défaut               |
|---------------------|----------------------------------------------------|-----------------------|
| `KERN_MEMORY_ADDR`  | Adresse d'écoute                                    | `127.0.0.1:7080`      |
| `KERN_MEMORY_TOKEN` | Jeton porteur exigé de l'appelant                   | (aucun, dev local)    |
| `KERN_MEMORY_DB`    | Chemin du fichier SQLite                            | `kern-memory.db`      |

Le binaire refuse de démarrer sur une adresse publique sans `KERN_MEMORY_TOKEN`.

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
