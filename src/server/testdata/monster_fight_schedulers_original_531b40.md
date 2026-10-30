# Original fight scheduler reference

The startup and nearest gzip files contain 4,138 inputs each, independently
executed from the unchanged GAME.EXE with SHA256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The first 3,510 rows retain the complete earlier inputs and outputs.
The startup control word is `0x0e3f` (precision53, toward zero); the separate
`0x023f` nearest profile differs in eighteen melee-scheduler outputs.
Production compares against the startup profile.

The four roots at 531B40, 531C10, 531C60 and 531D50 execute together with
the original PushAction50A260, reset50A110, condition50A010, buff4FF350,
distance4E6C00, combat predicates, hold-position predicate and name getter.
There are three declared service boundaries: canonical related-spell
selection540D90 with specified state changes; diagnostic output5341A0;
and a controlled action-cancel callback at23000500 selected through the
original action table. The producer's introductory two-service description
predates that third boundary; its executed callback contract and proof
records specify the added scope. Its bytes remain those of the executed
producer.

Each profile executes 27,718 original push calls and 215 cancel callbacks.
The 1,562 loaded original bytes include eighteen root padding bytes and one
buff-null return byte unreachable from these non-null roots. The other
1,543 bytes execute. No push success, numeric distance, predicate result,
stack mutation or floating-point outcome is substituted.

Go and actual-public-C comparators were frozen before the engine changes.
They compare guarded native allocations, both update records, all stack
slots and untouched arguments, declared callback effects and changed flags.
The Go comparator additionally records every push attempt; the C comparator
checks the native pointer, zero or integer-one return. The additional
ordinary-runtime C test uses unchanged original rows without a service
provider. Callback cases cover specified effects, not all registered action
bodies or all related spells.

Reproduce either original profile with Capstone and Unicorn installed:

```sh
python3 original_fight_schedulers_native_531b40.py \
  --game /path/to/GAME.EXE --control 0xe3f \
  --fixture /new/output.json.gz --proof /new/proof.json
```

Use `0x23f` for the independent nearest profile. The original executable,
raw before failures and both earlier references remain preserved outside
the repository. These bounded checks do not establish overall porting100%.
