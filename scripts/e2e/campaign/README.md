# Campaign regression replay

Build the game once, then replay a saved segment with Python 3.11 or newer:

```sh
python3 scripts/run-campaign-e2e.py \
  --suite scripts/e2e/campaign/wizard2-north-bookcase.json \
  --binary /path/to/opennox \
  --data /path/to/Nox \
  --seed /path/to/earned/Save \
  --output /path/to/new/report-directory
```

Read the suite's `seed_contract` first. The north bookcase case starts from an
earned Wizard save at the eastern Wiz02c obelisks with the listed spells and
uncleared northern enemies. It checks three original enemy deaths, both original
bookcase endpoints, a passage crossing and a healthy manual save. This segment
does not certify the entire chapter.

The subsequent segments have their own seed contracts:

| Suite | Earned source checkpoint | Required original enemies |
| --- | --- | --- |
| `wizard2-north-remaining-spider.json` | Healthy north-bookcase save | AlbinoSpider2345 |
| `wizard2-northeast.json` | Healthy remaining-spider save | AlbinoSpiders2346 and2347 |
| `wizard2-flying-golem.json` | Healthy northeast save, HP20/MP78 | FlyingGolem2348; earned level gain to HP26/MP88 |

Scenario changes reuse the supplied executable. Each run copies the Save and
preserves its executable, executed scenario, suite, logs, screenshots and
read-only world observations, including failures. Source saves and campaign data
inputs are checked for changes. Always use a new output directory.

The GitHub input-validation job checks every published scenario's allowed input
actions and the verifier's outcome checks. Actual gameplay replay requires local
game assets and produces a separate result. Campaign-only edits do not trigger
the Windows native build gate; product-source changes retain that gate.
