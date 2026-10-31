# Original missile action and prediction references

These references execute unchanged GAME.EXE bytes, SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The original startup/gameplay environment uses precision53 and ToZero
(`0xe3f`); nearest (`0x23f`) is retained as a diagnostic profile. Runtime
comparators check the startup reference with fixed hashes and dimensions.

`original_monster_missile_actions_532540.py` produces 4,150 contexts per
profile: 1,364 start532540 and 2,786 update532610. It executes all 197 start,
486 update and 64 prediction533080 body bytes; 21 padding bytes are not
credited as executed. Original Push/Pop, RNG, buff and sound lookup helpers
also execute. Allocation, map trace, placement/deletion, publication/audio,
player attack and registered end/cancel callbacks are declared boundaries.
Their event ordering and mutations are checked, while trajectory, prediction,
range spills, stack admission and RNG are never supplied as results.
Initialized type/resource caches are input; full transitive world behavior,
all resource lookup branches and original audio return values are not proved.

`original_monster_missile_lead_533080.py` executes all 64 prediction bytes
without any service substitution for 2,844 inputs per profile. Output and
pointer-return roles, full heap guards, stack, nonvolatile registers and
FPCW are checked. Original FPSW is recorded as diagnostic evidence, without
claiming native status-word equivalence. The two profiles differ in 905
prediction and 829 action outputs.

`original_monster_missile_lead_aliases_533080.py` adds 56 independent inputs
with output overlapping positions/velocity, plus signaling NaNs. Both Go
and actual C calls check the original sequential stores, all remaining
native bytes, native target return and absence of a server/class dependency.

The native FP helpers keep FLD quieting separate from arithmetic memory
operands. NaN selection and floating indefinite follow the masked x87 rules
in [Intel SDM volume1, table4-7 and section4.8.3.7](https://cdrdv2-public.intel.com/782149/253665-sdm-vol-1.pdf).
The original executable results remain the test authority.

One earlier test, `TestMonsterProjectileLeadAndTrajectory532610`, used
nearest-rounding approximations. The exact same coordinates were independently
executed by unchanged original update/prediction code, with observations only.
Gameplay mode returns lead-X `0x431cf1bb`, whereas nearest returns
`0x431cf1bc`. The test now checks exact original lead, spawn, velocity and
trace endpoint words instead of a decimal tolerance and nearest-derived
spawn. Its complete previous bytes and failure are preserved outside the
repository; all other 4,070 preceding inputs/assertions stay byte-identical.
`monster_missile_existing_test_original_532610.json`, both isolated original
profiles/proofs and `original_monster_missile_existing_test_532610.py` record
that correction. The primary 4,150/2,844 frozen inputs, expected values and
comparators are unchanged.

The original allocation controls call no C entry. All three actual C entries
first reproduced native pointer-truncation crashes. Post-port comparators
exercise their real public C symbols and native pointer/byte returns.
Ordinary binding tests use the real type registry, allocator and empty wall
map, with no allocation/trace replacement; create/delete/audio remain
observed world boundaries. Full product campaign, native platform and final
nine-cell audits remain separate requirements. This is not overall100%.
