---
id: okf-0002
feature: epic13-phase1
branch: feature/epic13-phase1
status: done
files:
  - internal/memory/contract.go
  - internal/memory/okf/store.go
  - internal/memory/vector/store.go
  - internal/memory/anon/store.go
  - internal/httpapi/memory.go
  - internal/httpapi/combined.go
  - internal/config/config.go
  - cmd/kern-memory/main.go
  - go.mod
tests:
  - internal/memory/contract_test.go
  - internal/memory/okf/store_test.go
  - internal/memory/vector/store_test.go
  - internal/memory/anon/store_test.go
  - internal/httpapi/memory_test.go
  - internal/httpapi/combined_test.go
decisions:
  - "2026-08-06 : construit EPIC-13 phase 1 en entier (contrat neutre + .okf + vecteur + kern-anon), pas seulement le sous-ensemble nécessaire au besoin #4 (RAG banques) — choix explicite de l'utilisateur, portée plus large que le strict besoin courtage."
  - "2026-08-06 : kern-obs (transverse « rappels tracés ») NON construit — la brique n'existe pas du tout comme repo (EPIC-11), documenté comme bloqué plutôt qu'improvisé avec une observabilité maison non demandée."
  - "2026-08-06 : même binaire que le store C8 v1 déjà construit (cmd/kern-memory), nouveaux endpoints /api/v1/memory/* à côté de /api/v1/documents/* — cohérent avec CLAUDE.md ('ce dépôt tient la place d'une tranche de cette brique'), pas un second daemon à faire tourner."
  - "2026-08-06 : embeddings via Ollama local (nomic-embed-text, 768 dims) — 100% self-host, cohérent avec l'argumentaire de souveraineté déjà vendu à AvelFinances (aucun appel réseau externe pour vectoriser une donnée)."
  - "2026-08-06 : couche vecteur = chromem-go (embarqué, persistance synchrone par écriture, zéro service externe) — suit directement la décision brainstorm 2026-07-22 du fichier docs/kern-memory-etat-de-lart.md."
  - "2026-08-06 : transverse kern-anon masque au moment de l'écriture SEULEMENT, aucune démasquage prévu — contrairement au pipeline courtage-extraction de kern-orch (masque -> modèle -> démasque dans UN run), la mémoire est longue durée et interrogée par des appels futurs sans lien entre eux ; le texte masqué est ce qui reste au repos, définitivement."
  - "2026-08-06 : Router.Write route par Kind vers UNE seule couche (okf OU vector), jamais les deux — chaque mémoire appartient à une seule couche par construction, contrairement à Query qui peut fan-out sur les deux quand Kind est vide."
---

**Quoi** : EPIC-13 phase 1 — brique mémoire agnostique, contrat neutre
`write(mémoire)`/`query(contexte)` routé par type (`internal/memory.Router`), deux couches
(`.okf` déclarative sur SQLite, vecteur sémantique via `chromem-go` + embeddings Ollama
locaux), transverse `kern-anon` (pseudonymise avant écriture, activé par défaut). Exposée
via `POST /api/v1/memory/write` et `/query` dans le MÊME binaire que le store C8 v1 déjà
construit — pas un second daemon.

**Vérifié en réel** : `go test ./...` vert sur tout le dépôt (7 packages), y compris des
tests d'intégration réels contre un vrai Ollama (`nomic-embed-text` installé cette
session) — pas de mock sur le calcul d'embeddings. Un vrai daemon lancé, testé via `curl` :
- Écriture d'une mémoire contenant un email réel → confirmé masqué (`<EMAIL_1>`) dans la
  réponse ET dans ce qui est effectivement persisté.
- Le cas d'usage réel du besoin #4 (RAG critères banques) : trois mémoires écrites (deux
  critères bancaires réalistes SCI/senior, un texte non lié — une recette), puis la
  question exacte du prospect (« Quelle banque accepte un prêt hypothécaire sur un bien en
  SCI avec un emprunteur de 74 ans ? ») a correctement classé les deux critères bancaires
  devant le texte non lié (similarité 0,83 vs 0,56).
- Couche `.okf` : écriture taguée + requête par tag, résultat exact, `similarity: 1`.
- Le store C8 v1 pré-existant vérifié intact (`GET /api/v1/documents` répond toujours).

**Pièges** :
- `chromem-go`'s `Query` exige `nResults > 0` et échoue si `nResults` dépasse le nombre de
  documents de la collection — `vector.Store.Query` clampe `Limit` à `collection.Count()`
  avant d'appeler, sinon une collection avec peu de mémoires ferait échouer toute requête
  avec un `limit` par défaut trop généreux.
- `chromem-go`'s `Metadata` n'accepte que `map[string]string` — `Tags`/`Metadata`/
  `CreatedAt` de `memory.Memory` sont repliés en champs JSON-encodés plutôt que d'élargir
  le schéma chromem-go par appelant.
