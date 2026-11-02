# Original ROAM start/cancel public entries 545790 / 5457C0

This pre-engine reference executes the original GAME.EXE identified by
SHA-256 `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The retained producer is `ROOT/monster-roam-lifecycle-native-entry-port/original-roam-lifecycle-v1.py`.
It executes both leaf functions directly, with no callee services, replaced
instructions, arithmetic or substituted results.

| Entry | Complete instruction bytes | SHA-256 |
| --- | ---: | --- |
| 545790 | 43 | `8935856131ee4c84722f66f35cb2d17ea9ebcf248c65bad7583e046f455daa6c` |
| 5457C0 | 32 | `5e58e34d04aa27066ba2172b237397a038c814a1a69f29306d9aec24f75f76ff` |

All 75 instruction bytes execute. The five alignment bytes after 545790 earn
no instruction credit. There are no constant-data ranges. Start loads the
update pointer, performs the original unused waypoint read and sixteen-step
register loop, clears waypoint zero and its cursor, then returns the update
pointer. Cancel loads a signed byte stack index, clears the corresponding
first argument word and returns that update pointer. Neither raw entry
checks class, null pointers or index bounds.

Each PC53/ToZero startup `0xe3f` and nearest `0x23f` profile executes 1,321
contexts, including six faults: null unit, null update and unmapped update
for each entry. Start covers seven signed indices and five whole-word seed
patterns/classes. Cancel covers every signed-byte index and all five seed
patterns/classes. Whole original image and 64 KiB heap, guarded caller bytes,
valid-return nonvolatile registers, CW and TOP are checked. Every original
store is observed only at its original instruction address, destination and
width. All other heap bytes remain exact. Both complete outputs are identical.

Frozen gzip SHA-256:
`d01324ba2502d77649e7892c9be0572044a05f08edc3404cdc8e4ef02fde6552`.
Decoded SHA-256:
`63613298e0debb443b96e078362b444bbcd411302f6042b242754c0c801f8680`.
Both profiles and their independent execution proofs remain outside the repo.

The consumer checks the actual C entry and a raw Go contract against the
unchanged original write history, complete native unit and guarded 64 KiB
native record. An independent C mirror checks object/MUD/action offsets,
native stride and argument padding. Pointer stores and numeric AI argument
stores clear a whole native slot; untouched upper uintptr bits are guarded.
Every successful result must retain the complete native update pointer.

Valid indices 0 through 23 refer to actual AI stack slots. Negative and
out-of-range indices also retain signed, unchecked native array arithmetic
inside the guarded allocation. Those diagnostic projections do not establish
equivalence of named fields outside the array: native pointer expansion
changes such overlapping layouts. They do not certify malformed game states.
Null/unmapped raw faults and ordinary admitted behavior are distinct checks.

The getter supports both legacy integer signatures and native pointer
signatures without editing the comparison during the engine port. The before
test first checks independent C/Go layouts, ordinary native start/cancel,
typed sound/action/health controls and whole initial records; each old C
entry is then invoked in its own process above 4 GiB. Exact executed test
binaries and all failure output are retained. This reference does not claim
campaign progress, the complete ROAM action graph or overall 100% porting.
