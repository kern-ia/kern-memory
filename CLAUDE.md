# CLAUDE.md — kern-memory

## Contexte

`kern-memory` porte une tranche du contrat **C8** (voir `../Kern-UI/docs/expected-contracts.md`)
côté stockage : la vue Rédaction a besoin qu'un document et ses suggestions voyagent du
stockage au navigateur, et qu'une décision (accepter/ignorer) voyage en retour. Ce dépôt
n'est **pas** le kern-memory décrit dans `../Kern-Orch/docs/ROADMAP.md` (EPIC-13 : RAG,
embeddings, `chromem-go`, couche `.okf`) — c'est un stockage minimal qui tient la place d'une
tranche de cette brique, décidé ainsi le 2026-07-30 pour livrer C8 v1 sans construire la RAG
en même temps.

## Rôle

Un daemon HTTP minimal, lu par kern-ui et rien d'autre :
- `GET /api/v1/documents` — la liste (id, titre, nombre de mots, dernière mise à jour).
- `GET /api/v1/documents/{id}` — un document et ses suggestions.
- `POST /api/v1/documents/{id}/suggestions/{sid}/accept` et `.../ignore`.

**Ce qu'il ne fait PAS** : générer des suggestions, éditer un document, faire de la
recherche sémantique. V1 est lecture + décision sur des suggestions qui existent déjà. La
seule voie d'entrée pour du contenu est la commande CLI `seed` — pas de chemin HTTP
d'écriture de contenu en V1, exactement comme `kern-ui useradd` est la seule voie pour créer
un compte.

## Indépendance des briques

`kern-memory` ne connaît rien des internes de kern-ui ni de kern-orch. Un seul jeton porteur
(`KERN_MEMORY_TOKEN`) : kern-ui est son unique appelant, donc pas de séparation
producteur/session comme kern-orch en a besoin pour deux populations d'appelants distinctes.

## Vocabulaire affiché

Ce dépôt ne sert que du JSON — aucun texte destiné à l'utilisateur ne vit ici. Le
vocabulaire affiché (mission, compétence, etc.) est fixé dans `Kern-UI/web/src/i18n/fr.ts`.

## Méthode

- **TDD**, `go test`. Le calcul du nombre de mots et le découpage du corps autour d'une
  ancre de suggestion sont de la logique pure, testée sans base ni réseau.
- **SQLite** (`modernc.org/sqlite`, pur Go, pas de cgo) : `SetMaxOpenConns(1)` +
  `PRAGMA busy_timeout` dès l'ouverture — leçon apprise sur kern-orch (`internal/checkpoint`)
  cette même session, appliquée ici avant qu'un vrai bug de concurrence ne la réapprenne.
- **Le binaire refuse de démarrer** sur une adresse publique sans jeton — même règle que
  kern-ui et kern-orch, redérivée ici, pas partagée.
- **Git** : `main` ← `dev` ← `feature/xx`. Jamais de commit direct sur main/dev.
- **Code et documentation en anglais**, comme le reste de l'écosystème Kern.
- **Secrets** : jamais dans le dépôt. `KERN_MEMORY_TOKEN` lu depuis l'environnement.
- **Index OKF** : `docs/index/<n>-<feature>.md` par feature, `docs/retro.md` en continu.

## Commandes

- Build : `go build ./...`
- Tests : `go test ./...`
- Servir : `kern-memory serve` (voir README.md pour les variables d'environnement).
- Semer un document : `kern-memory seed <title> <body-file>`
