# Rétro — kern-memory

## 2026-07-30 — Bootstrap + store-and-api

Ce qui a marché : reprendre tel quel le pattern SQLite de `kern-orch/internal/checkpoint`
(`SetMaxOpenConns(1)` + `PRAGMA busy_timeout`) et le refus de démarrage sur adresse publique
sans jeton, plutôt que de les redécouvrir. Les deux avaient coûté un vrai bug la première
fois (voir `Kern-Orch/docs/retro.md`, C6) ; aucun cette fois.

Rien à signaler côté pièges — première feature du dépôt, tout vérifié en réel dès le départ
(binaire réel, `curl` réel, refus d'exposition testé en le déclenchant pour de vrai).
