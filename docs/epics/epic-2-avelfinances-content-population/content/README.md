# Content — epic 2, issue 2

`bank-criteria.json` is **simulated / fictional content**, not real AvelFinances data.

AvelFinances' actual bank-lending criteria are confidential and belong to the agency —
they were not available when this content was authored
([issue #20](https://github.com/kern-ia/kern-memory/issues/20)). Everything in
`bank-criteria.json` — bank names (`Banque Solstice Fictive`, `Crédit Meridian Fictif`,
`Banque Arcadelle Fictive`), thresholds, percentages, and criteria — is invented for
this repo, standing in for the real content so the memory layers can be populated and
queried end-to-end (see [issue #21](https://github.com/kern-ia/kern-memory/issues/21)
for the recall verification that follows). Do not treat any of it as real bank policy
or real client data.

Load it into a running `kern-memory serve` instance with:

```
kern-memory load-memory docs/epics/epic-2-avelfinances-content-population/content/bank-criteria.json
```

Content is in French — the domain content AvelFinances' end users query in, distinct
from this repo's own code/docs, which stay in English per `CLAUDE.md`.
