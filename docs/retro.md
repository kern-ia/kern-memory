# Rétro — kern-memory

## 2026-07-30 — Bootstrap + store-and-api

Ce qui a marché : reprendre tel quel le pattern SQLite de `kern-orch/internal/checkpoint`
(`SetMaxOpenConns(1)` + `PRAGMA busy_timeout`) et le refus de démarrage sur adresse publique
sans jeton, plutôt que de les redécouvrir. Les deux avaient coûté un vrai bug la première
fois (voir `Kern-Orch/docs/retro.md`, C6) ; aucun cette fois.

Rien à signaler côté pièges — première feature du dépôt, tout vérifié en réel dès le départ
(binaire réel, `curl` réel, refus d'exposition testé en le déclenchant pour de vrai).

## 2026-08-06 — EPIC-13 phase 1 (mémoire agnostique)

Ce qui a marché : réutiliser tel quel le pattern SQLite du store C8 v1 (`SetMaxOpenConns(1)`
+ `PRAGMA busy_timeout`) pour la couche `.okf`, plutôt que de le redécouvrir une deuxième
fois dans ce même dépôt. Vérifier qu'Ollama tournait ET qu'un modèle d'embeddings était
installé AVANT d'écrire le moindre code (aucun n'était disponible au départ — `ollama pull
nomic-embed-text` a suffi, mais ça aurait pu ne pas suffire).

Piège trouvé (pas un bug, une contrainte d'API découverte en écrivant les tests) :
`chromem-go`'s `Collection.Query` exige `nResults > 0` et échoue si `nResults` dépasse le
nombre de documents dans la collection — n'importe quel appelant avec un `Limit` par défaut
généreux (5, disons) casserait sur une collection avec 1 ou 2 mémoires seulement. Corrigé en
clampant `Limit` à `collection.Count()` avant l'appel. Généralisable : toute lib de
recherche/top-K doit être vérifiée sur le cas "moins de résultats que demandé", pas
supposée le gérer silencieusement.

kern-obs (transverse « rappels tracés » du ROADMAP EPIC-13) documenté comme bloqué plutôt
que construit — vérifié qu'aucun repo `kern-obs` n'existe avant de improviser quoi que ce
soit à sa place.
