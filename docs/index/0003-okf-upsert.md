---
id: okf-0003
feature: okf-upsert
branch: feature/okf-upsert
status: done
files:
  - internal/memory/okf/store.go
tests:
  - internal/memory/okf/store_test.go
decisions:
  - "2026-08-07 : Write upserte (ON CONFLICT(id) DO UPDATE) au lieu d'échouer sur un id déjà présent — besoin découvert en construisant la persistance du calendrier marketing de Kern-UI, qui a une clé stable (l'id du run) et doit pouvoir réécrire le même item plusieurs fois."
---

**Quoi** : `okf.Store.Write` upserte désormais sur l'id plutôt que d'échouer avec une
violation de contrainte UNIQUE à la deuxième écriture du même id.

**Vérifié en réel** : `go test ./...` vert (tout le dépôt), nouveau test couvrant deux
écritures successives sur le même id — confirme une seule ligne en base et le contenu de
la deuxième écriture, pas un doublon ni une erreur.

**Pièges** : aucun.
