# Original raw monster-death callbacks

These references execute original PE32 instructions for `549D80` (124 bytes),
`549E00` (104), `549E70` (27) and `54A250` (27): **282 instruction bytes**.
Alignment padding is excluded. The native engine is unchanged when these
inputs, producers, goldens and manifest seals are added. Native getters,
consumers, actual old-C reproductions and the port remain pending.

The independently generated 305 inputs run under PC53/toward-zero `0xe3f`
and PC53/nearest `0x23f`. Both outputs are byte-identical. All original
instructions execute. Complete image and 64 KiB heap checks admit only
declared world mutations; guarded caller bytes, valid-return stack and
nonvolatile registers, and FPU CW/TOP remain intact. Reentrant cases call
original bodies rather than a replacement callback implementation.

Seven world services are intercepted: flag `40A5C0`, push `52E040`, damage
`4E25B0`, spark `5231B0`, audio `501960`, delete `4E5CC0`, point `522FF0`.
The services' original instruction bodies are outside this execution scope.
Eight inputs stop at an explicitly declared world position-read fault.
There are **zero faults in these original entry instructions**; the eight
service faults are not labelled as original-instruction or actual old-C
crashes. Invalid-point services that do not read the point are retained too.

The original caches its flag-dependent damage choice but keeps a live point
address across push/damage/spark. A declared push mutation is visible to the
later services. The existing ordinary helper's copied point needs its own
consumer and before comparison; no ordinary-gameplay failure is claimed.

Original GAME.EXE SHA-256:
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Input gzip: `acec7841c6a433cfaa36972977c9b471049c1f5780e4605e5cf53521bb39caeb`.
Both golden gzip hashes:
`951ae5cda4a689cd7b717c1479e6a201a6add9380edcbe5104df62008a90f54f`.
Both decoded golden hashes:
`3f3255d2dc3c6c276b735532ed704a67012709d1cdba8cf21fb3341cee21288c`.
The independent generator regenerates the input byte for byte. The producer
reads only GAME.EXE and its adjacent input fixture; it does not read native
source or derive expected values from the new implementation.

With Unicorn 2 and Capstone 5 installed, reproduce each profile into a fresh
output path:

```sh
python3 original_monster_death_explosion_native_549d80.py --exe /path/to/GAME.EXE --control 0xe3f --output /tmp/death-tozero.json.gz --proof /tmp/death-tozero-proof.json
python3 original_monster_death_explosion_native_549d80.py --exe /path/to/GAME.EXE --control 0x23f --output /tmp/death-nearest.json.gz --proof /tmp/death-nearest-proof.json
```

Local complete proofs and preparation files remain under
`ROOT/monster-death-explosion-native-entry-port`. These original seals are
not native ABI, whole-game, campaign or overall 100% completion credit.
