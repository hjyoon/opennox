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
| `wizard2-eastern-spiders.json` | Healthy FlyingGolem save, HP26/MP88 | Spiders2341 and2255; alternate original obelisks restore full mana |
| `wizard2-necromancer-pit.json` | Healthy eastern-spiders save, HP26/MP88 | Original Necromancer activates; scripted NecroTemp death after the pit scene |
| `wizard2-book-of-oblivion.json` | Completed pit-scene save, HP26/MP88 | Urchin2362 and original BookOfOblivion pickup; final HP26/MP53 |
| `wizard2-book-recovery.json` | Book pickup save, HP26/MP53 | Fresh Load retains the book; native passive mana regeneration returns HP26/MP88 |

Replay all four segments consecutively with one command and the same prebuilt
executable:

```sh
python3 scripts/run-campaign-e2e.py \
  --sequence scripts/e2e/campaign/sequences/wizard2-cave-east.json \
  --binary /path/to/opennox \
  --data /path/to/Nox \
  --seed /path/to/earned/Save \
  --output /path/to/new/sequence-report-directory
```

The initial Save must match the north bookcase contract. The runner validates
every stage before starting the game, then passes each successful stage's exact
earned Save to the next stage automatically. It stops at the first failure,
including an unmet objective with an otherwise healthy player. A passing
sequence report includes `final_seed` and its file hashes. Each stage keeps its
own report and runtime Save. To retry one segment, use `--suite` with that
segment's source Save from `stages.jsonl`; earlier segments need not be replayed.
For a route edit, run that single suite against the supplied executable first.
Use the sequence to check that consecutive checkpoints work together. Native
build and package tests remain separate from these input-only changes.

The eight-stage `sequences/wizard2-cave-north-to-book.json` extends that same
starting checkpoint through the eastern spiders, Necromancer pit scene, book
pickup and full mana recovery. Use an executable with
`pickup-named-object-input` for this longer sequence. Each individual segment
has passed gameplay replay; the eight-stage sequence still needs its own full
consecutive replay before claiming that combined result.

Resume a stopped sequence into a new report directory by adding
`--resume-from /path/to/previous-sequence-report` to the original command.
Keep the original `--seed` and the same executable, data path and config. The
runner validates the passing prefix's suite, scenario, full earned Save,
executable, input assets and read-only observation hashes before starting any
game. It replays the first failed or unattempted stage and everything after it.
Routes in those later stages may be fixed; changing a passed stage or its
evidence rejects reuse. Interrupted runs can resume from complete passing
`stages.jsonl` entries even without a final report. Earlier reports stay intact.
The new report distinguishes `reused_stages` from `executed_stages` and sets
`fresh_replay: false` when it reuses evidence. To perform a fresh full replay
after rebuilding the product, omit `--resume-from`.

Scenario changes reuse the supplied executable. Each run copies the Save and
preserves its executable, executed scenario, suite, logs, screenshots and
read-only world observations, including failures. Source saves and campaign data
inputs are checked for changes. Always use a new output directory.

The GitHub input-validation job checks every published scenario's allowed input
actions and the verifier's outcome checks. Actual gameplay replay requires local
game assets and produces a separate result. Campaign-only edits do not trigger
the Windows native build gate; product-source changes retain that gate.

`pickup-named-object-input` selects an original named map item through the
normal client cursor. It clicks only when that exact drawable has the Pickup
cursor; distant, unavailable and other-cursor targets remain bounded waits.
