# CLAUDE.md — kern-memory

## Contexte

Ce dépôt porte DEUX choses distinctes, dans le même binaire (`cmd/kern-memory`) :

1. **C8 v1** (`internal/store`, `internal/httpapi/router.go`) : stockage minimal
   document/suggestion pour la vue Rédaction de kern-ui — voir
   `../Kern-UI/docs/expected-contracts.md`. Décidé le 2026-07-30 pour livrer C8 v1 sans
   construire la RAG en même temps.
2. **EPIC-13 phase 1** (`internal/memory/*`, `internal/httpapi/memory.go`) : la vraie brique
   mémoire agnostique décrite dans `../Kern-Orch/docs/ROADMAP.md` — contrat neutre
   `write`/`query`, couche `.okf` (SQLite, déclarative) + couche vecteur (`chromem-go` +
   Ollama, sémantique), transverse `kern-anon` (pseudonymise avant écriture, activé par
   défaut). Construite le 2026-08-06 pour servir le besoin #4 de l'agence de courtage
   (RAG critères banques, `../Kern-Orch/skills/specs.md`). **kern-obs (transverse "rappels
   tracés") non construit** — la brique n'existe pas du tout (EPIC-11), documenté comme
   bloqué plutôt qu'improvisé.

Les deux coexistent dans le même processus (`api/v1/documents/*` et `api/v1/memory/*`),
chacun avec son propre stockage (SQLite pour C8 et `.okf`, fichiers `chromem-go` pour le
vecteur) — aucun des trois ne lit les données d'un autre.

## Rôle

**C8 v1**, lu par kern-ui et rien d'autre :
- `GET /api/v1/documents` — la liste (id, titre, nombre de mots, dernière mise à jour).
- `GET /api/v1/documents/{id}` — un document et ses suggestions.
- `POST /api/v1/documents/{id}/suggestions/{sid}/accept` et `.../ignore`.

Ce qu'il ne fait PAS : générer des suggestions, éditer un document, faire de la recherche
sémantique. V1 est lecture + décision sur des suggestions qui existent déjà. La seule voie
d'entrée pour du contenu est la commande CLI `seed` — pas de chemin HTTP d'écriture de
contenu en V1, exactement comme `kern-ui useradd` est la seule voie pour créer un compte.

**EPIC-13 phase 1**, lu par kern-orch (ou tout appelant du même écosystème) :
- `POST /api/v1/memory/write` — `{"kind":"okf"|"vector","text":"...","tags":[...],"metadata":{...}}`.
  Kind vide route vers la couche vecteur (défaut : la plupart des mémoires n'ont pas de clé
  de recherche naturelle). Le texte est pseudonymisé avant écriture sauf
  `KERN_MEMORY_PSEUDONYMIZE=false`.
- `POST /api/v1/memory/query` — `{"text":"...","kind":"okf"|"vector"|"","tags":[...],"limit":N}`.
  Kind vide interroge les deux couches et fusionne par similarité décroissante ; la couche
  `.okf` répond en similarité 1 (lookup exact par tag, pas un classement).

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
