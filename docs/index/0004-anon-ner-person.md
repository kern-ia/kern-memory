---
id: okf-0004
feature: anon-ner-person
branch: feature/anon-ner-person
status: done
files:
  - internal/memory/anon/store.go
  - internal/memory/anon/ner_onnx.go
  - internal/memory/anon/ner_noop.go
tests:
  - internal/memory/anon/store_test.go
  - internal/memory/anon/store_onnx_test.go
decisions:
  - "2026-08-07 : LOCATION/ORGANIZATION exclus du masquage même si détectés par NER — plus important encore ici que côté Kern-Orch : la mémoire stocke des critères banques (« La banque Alpha Crédit accepte... ») où le nom de l'organisation EST le contenu recherché, pas du PII. Masquer casserait le RAG banques (besoin #4) plutôt que de simplement dégrader la qualité."
  - "2026-08-07 : même patron singleton + build tag onnx que Kern-Orch (internal/cmd/courtage_ner_onnx.go) — dupliqué, pas partagé (deux modules Go séparés), cohérent avec le choix déjà pris pour piiTokenLabels dans ce même fichier."
---

**Quoi** : le transverse kern-anon d'EPIC-13 masque maintenant les noms de personnes
(PERSON) quand un moteur NER ONNX est configuré (`KERN_ANON_NER_MODEL_DIR`), en plus des
entités à motif fixe déjà masquées. LOCATION et ORGANIZATION, bien que détectées par le
même moteur, sont explicitement exclues.

**Vérifié en réel** : `go test ./...` et `go test -tags onnx ./...` verts. Test
d'intégration réel (modèle ONNX réellement chargé, pas de mock) : un texte contenant un
nom de personne ET un nom de banque confirme le nom masqué (`<PERSONNE_1>`) et le nom de
banque conservé en clair.

**Pièges** : aucun nouveau — mêmes leçons que `Kern-Orch/docs/index/` pour ce même sujet
(voir ce dépôt pour le détail de la mise en place ONNX/macOS).
