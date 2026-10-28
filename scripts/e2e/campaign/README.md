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
| `video-book-save-load-smoke.json` | Healthy book-recovery save, HP26/MP88 | Fresh Load, retained book and spell bindings, original road99 movement and full-resource manual Save; no new enemy objective |
| `wizard2-upper-urchins.json` | Healthy book-recovery save, HP26/MP88 | Original Urchins2363/2364, retained book, HP26/poison0 manual Save; the qualified run ends at MP2 and needs normal recovery before further spell-heavy combat |
| `wizard2-upper-mana-recovery.json` | Upper-Urchins save, HP26/MP2, book1 | Southern floor bypass, original BluePotion1940 automatic +50 mana and Meat1939/1938 pickups; retained book, poison0 and manual Save with HP at least20 and MP at least52 |
| `wizard2-swap-escape.json` | Earned Wiz02b Swap-book save, HP20/MP78, F=Swap115, player at2822.5457,2758.513 | One ordinary F cast at the original moving SmallAlbinoSpider; reciprocal native position exchange in one frame, full resources and living manual Save |

For a short Load, movement, pickup and manual Save regression, use that last
suite with its earned upper-Urchins checkpoint:

```sh
python3 scripts/run-campaign-e2e.py \
  --suite scripts/e2e/campaign/wizard2-upper-mana-recovery.json \
  --binary /path/to/opennox \
  --data /path/to/Nox \
  --seed /path/to/earned/upper-urchins/Save \
  --output /path/to/new/recovery-report-directory
```

This segment has taken about 30 seconds on the ARM64 validation machine. It
checks the listed outcomes automatically and preserves evidence on failure.
It does not exercise every enemy, an obelisk interaction or an entire chapter.
Use the suite that exercises the changed behavior for gameplay coverage.

The Swap case uses its own earlier earned checkpoint, before the barred-room
escape. Run it with `--suite scripts/e2e/campaign/wizard2-swap-escape.json`
and that Save. Its native position requirement rejects an uneventful cast or
a Save that already starts outside the room. The quickbar input observer copies
both original positions once per frame; the verifier requires the pinned pair
to exchange positions between adjacent frames after the actual F key input.
The moving spider has no fixed landing coordinate. Ordinary movement, a change
in just one object, missing frames or a different target cannot pass. The route
has no movement input after Load. This case does not
certify Magic Missile, later cave combat or the entire chapter.

An input-route edit uses the existing executable and the affected suite.
For a native gameplay port, run the final normal package checks once, focus
strict pointer checks on the changed paths, build one clean normal product,
and replay the relevant saved segments. Record native platform CI separately;
an observation timeout does not require restarting an existing CI job.

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

For a short combat and resource regression, use
`--sequence scripts/e2e/campaign/sequences/wizard2-upper-cave.json` with the
earned BookOfOblivion recovery Save (HP26/MP88, both upper Urchins alive).
It reuses one executable for the upper Urchin fights and the following original
BluePotion1940 and Meat1939/1938 pickups, passing the actual manual Save between
them. Recovery leaves the lower room along its southern floor before road154;
the direct road99 goal can overlap the moving Troll group and trap the player.
Replaying this pair does not require rebuilding or repeating the earlier cave route.

Scenario changes reuse the supplied executable. Each run copies the Save and
preserves its executable, executed scenario, suite, logs, screenshots and
read-only world observations, including failures. Source saves and campaign data
inputs are checked for changes. Always use a new output directory.

Case and sequence summaries include `failed_checks`. Resource failures report
the actual value and required value; missing enemy deaths report their original
script IDs. A healthy Save with an unmet objective still fails and stops the
sequence. These summaries let route retries target the failing segment without
rebuilding the game or manually searching all captured observations.

For consumed map items, `required_map_pickups` lists each original `extent`
and `item` type. The verifier requires initial presence, actual ordinary Pickup
cursor input for that exact identity, and removal from a final live scene.
The observed map must match the single expected map. Missing food, a different
cursor or item, another map, or an unloaded scene fails the case even if its
final health and mana are sufficient.

Cases can require `minimum_health` and `minimum_mana` for segments where
ordinary combat or passive regeneration varies the exact final resources.
Existing `expected_health` and `expected_mana` checks remain exact when present.
The upper mana recovery segment requires a reserve of HP20/MP52; it does not
claim full resources or completion of the northern route. The original potion
is consumed immediately when its class-scaled recovery would stay below maximum
mana, so the scenario checks zero potion inventory and the recovered mana.

The GitHub input-validation job checks every published scenario's allowed input
actions and the verifier's outcome checks. Actual gameplay replay requires local
game assets and produces a separate result. Campaign-only edits do not trigger
the Windows native build gate; product-source changes retain that gate.

`pickup-named-object-input` selects an original named map item through the
normal client cursor. It clicks only when that exact drawable has the Pickup
cursor; distant, unavailable and other-cursor targets remain bounded waits.

`pickup-map-object-input` uses the original map object's `extent` and expected
`item` type for unnamed items. It uses the same live drawable and near Pickup
cursor gate, followed by an ordinary mouse press and release. It neither moves
the item nor inserts it into inventory.
