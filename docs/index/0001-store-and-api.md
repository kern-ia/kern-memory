---
id: okf-0001
feature: store-and-api
branch: feature/store-and-api
status: done
files:
  - internal/store/model.go
  - internal/store/sqlite.go
  - internal/httpapi/router.go
  - internal/config/config.go
  - internal/config/exposure.go
  - cmd/kern-memory/main.go
tests:
  - internal/store/model_test.go
  - internal/store/sqlite_test.go
  - internal/httpapi/router_test.go
  - internal/config/exposure_test.go
decisions:
  - "2026-07-30 : dépôt bootstrap avec ce seul document + suggestion, pas une RAG — voir CLAUDE.md."
  - "2026-07-30 : SetMaxOpenConns(1) + PRAGMA busy_timeout dès l'ouverture, même leçon que kern-orch/internal/checkpoint, appliquée avant qu'un bug de concurrence ne la réapprenne."
  - "2026-07-30 : pas de chemin HTTP d'écriture de contenu en V1 — seul `kern-memory seed` écrit un document, exactement comme kern-ui useradd est la seule voie pour un compte."
  - "2026-07-30 : Journal des décisions n'a pas de schéma séparé — c'est un Document comme un autre, la distinction reste à faire côté kern-ui si besoin."
---

**Quoi** : `kern-memory` — nouveau dépôt, un daemon HTTP minimal (SQLite, un jeton porteur)
qui sert des documents et leurs suggestions, plus une commande CLI `seed` puisqu'il n'existe
aucun chemin HTTP d'écriture de contenu en V1. Trois endpoints : liste, lecture d'un
document avec ses suggestions, résolution (`accept`/`ignore`) d'une suggestion.

**Vérifié en réel** : `kern-memory serve` + `kern-memory seed` + `curl` — liste, lecture avec
une suggestion insérée directement en SQLite, `accept`, relecture confirmant
`status: accepted`. Jeton porteur vérifié réellement refusé sans en-tête (401) et accepté
avec (200). Refus de démarrage sur une adresse publique sans jeton vérifié en réel (code de
sortie 1, message explicite).

**Pièges** : aucun cette fois — le pattern SQLite et le refus d'exposition publique étaient
déjà rodés sur kern-orch et kern-ui, repris tels quels plutôt que redécouverts.
