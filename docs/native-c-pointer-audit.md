# Native C entry pointer audit

This inventory prioritizes remaining public C ABIs after the native point,
occupancy, retreat and ROAM lifecycle ports. It does not supply a porting
denominator or count confirmed runtime crashes.

At `616b61d`, host macOS ARM64 normal preprocessing of all **191** top-level
`src/legacy/*.c` files finds **3,502** active function definitions. Real clang
preprocessing excludes the preserved `#if 0` original bodies; all 191
preprocessor logs have zero warnings. Full preprocessed outputs, input
hashes, function/caller inventories and scripts are retained under
`ROOT/monster-retreat-predicates-native-entry-port`.

The heuristic finds **303** definitions casting an integer parameter to a
pointer. **289** have a fixed 32-bit `int`/`int32_t`/`uint32_t` operand, while
**14** use native-width or other integer types. Of the 289, **26** have direct
non-test Go `C.symbol` references and **171** have direct active C callers.
References may store or compare callback addresses rather than invoke them.
Forwarding entries without their own cast, cgo preambles, aliases, pointer
returns, globals and integer fields need separate analysis. Other target
preprocessor configurations also remain outside this host inventory.

Semantic inspection is required before changing a candidate:

| Example | Classification |
| --- | --- |
| `548630` collision target and `548830` door data | Already use `uintptr_t`; these are not fixed 32-bit argument truncations. |
| `4597E0` server-options hide argument | Integer control flag, also returned through a legacy pointer-shaped result; a cast alone does not establish a pointer input. |
| `56F9E0` protected-mana argument | Integer protection handle resolved by `56F590`, with sentinel results; changing it to an object pointer would change its meaning. |
| `549D80` and related death symbols | Some Go references classify callback addresses and select existing native implementations. Raw C ABI and actual dispatcher behavior must be checked separately. |
| `545520`, `545580`, `5455B0` | Actual legacy-int public entries each SIGSEGV above 4 GiB after independent native-layout, typed sound/action/buff and whole-allocation positive controls pass. Frozen original comparison precedes the native port. |

The three retreat predicates now use typed `nox_object_t*` public arguments
and shared native Go cores. Their result remains the original canonical
integer. Raw entries retain original faults and short circuits; ordinary
callers keep their earlier admission checks. Original C bodies remain exact
under `#if 0`, and the frozen original getters/consumers are unchanged.

At ported source `43ae3d9`, the repeated host scan finds **3,499 active C
definitions / 301 integer-cast candidates / 287 fixed-32-bit candidates**.
Only the resume and cast predicates disappear from the cast inventory; the
third, forwarding retreat predicate had no cast of its own. All three typed
public ABIs are independently typechecked and compared with the original.
The fourteen native-width/other candidates remain separated. This scan
introduces no new fixed-32-bit candidate or preprocessor warning.

At that preceding `43ae3d9` source, raw `5456D0` still has an integer object
argument. No direct production call is found outside its definition/header;
the ordinary dispatcher already uses a native Go update. That classification
does not close indirect reachability. The subsequent actual C reproduction
SIGSEGV above 4 GiB after independent layouts, real native owner/Push
composition and typed sound/action/buff/predicate controls pass.

Runtime `8212e8c` ports raw `5456D0` to a typed `nox_object_t*` entry and shares
its update core with the ordinary dispatcher. The raw entry retains original
faults, pointer-load order, cached MUD/live owner behavior, unspilled distance
arithmetic and ordered action writes. Ordinary admission, including guarded
missing health, remains intact. Five original references are frozen before
engine changes, including a corrected C-prefix-versus-Go-EXT layout check;
the initial preparation failure and actual old C crash remain preserved.

The repeated `8212e8c` host scan finds **3,498 active C definitions / 300
integer-cast candidates / 286 fixed-32-bit candidates**. Only `5456D0` leaves
the fixed-width inventory. Fourteen native-width/other candidates remain
separate; no new candidate or preprocessor warning is introduced. Integer
object entries `5455E0`, `545790` and `5457C0` remain follow-up targets at that
source.
Their ordinary native dispatcher implementations do not port those raw C ABIs.

Runtime `7d404efc` subsequently ports ROAM start `545790` and cancel `5457C0`
to native object arguments and pointer returns, sharing their stores with the
ordinary dispatcher while retaining its admission rules. Clean `f6a41c78e`
has identical engine inputs and a corrected opaque-bit fixture preparation.
Both old actual C entries SIGSEGV above 4 GiB after native positive controls.
The independent original executes 75 instruction bytes and 1,321 contexts
per FPU profile, including six faults. The first frozen fixture's unaligned
typed-pointer construction fails strict preparation; a raw-byte replacement
is proved identical across complete 64 KiB allocations at untouched pre-engine
source, retaining original goldens/expected values/guards and all failed files.
The original four other references remain exact. Corrected normal/strict
selections each pass 21,126 events, with matching identities and no new skips.

The repeated host scan finds **3,496 active C definitions / 298 integer-cast
candidates / 284 fixed-32-bit candidates**. Only the two ROAM entries leave
the fixed-width inventory; no new candidate or preprocessor warning is added.
Raw cancel keeps signed unchecked native array arithmetic. Diagnostic
out-of-array projections do not establish named-field alias equivalence after
pointer expansion or certify malformed states; valid slots and ordinary
empty/invalid-head behavior have separate checks.

An additional scan reuses all **191 hash-verified preprocessed C inputs** at
`f6a41c78e`. It finds **45 definitions** explicitly casting a standalone pointer
parameter value to a fixed 32-bit integer; eleven overlap the earlier
all-integer candidate inventory. Scalar field/index/call operands are excluded;
the preliminary broader 100-match inventory remains preserved. Twelve of the
45 have direct non-test Go C.symbol references, which may store callback
addresses rather than invoke them. These views must not be added into a
remaining-defect total. They are semantic triage, with local aliases,
pointer-shaped returns,
callbacks, globals/fields and other narrowing expressions still outside the
expression scan. The full inventory and provenance are retained in
`ROOT/monster-roam-lifecycle-native-entry-port/active-C-typed-pointer-parameter-narrowing-inventory-v2.json`.

For example, `544A00` accepts `nox_object_t*` but explicitly casts it to `int`
before computing the position address. It also returns an integer-shaped
object pointer and stores the selected pointer in uint32 scratch. Its callback
`544A40` still takes two integer object arguments; outer edible-check `5455E0`
is another legacy-int public entry. This family requires its own original
execution, actual C reproductions, native scratch/return and ordinary-caller
analysis before porting. No crash or completion credit is inferred from these
new static findings alone.

Runtime `c351c49c0` subsequently ports all three edible-family entries.
`5455E0` returns a native action-slot pointer, `544A00` uses native search
input/result/shared scratch, and `544A40` accepts native candidate/owner
pointers. Raw live globals, nested search and fault order remain distinct
from ordinary admission and local scratch. All three actual old C entries
fault above 4 GiB after native positive controls. Eight original references
are frozen before the engine change; each FPU profile executes 3,406
contexts, including thirteen faults, across 439 instruction bytes. Normal
and strict selections each pass 14,761 events, including fifty fresh actual-C
production children, without changing those frozen references.

The new hash-bound host scan finds **3,493 active definitions / 297 integer-
cast candidates**. Only `544A40` leaves that cast view: forwarding `5455E0`
was missed because it had no cast of its own, while `544A00` belonged to the
typed-parameter view. The latter now contains **44 definitions**, eleven
overlapping the integer-candidate view. Preprocessor warnings remain zero.
Source/header hashes are independently rebound to clean `c351c49c0`;
earlier working-tree snapshots are not relabelled as clean-source captures.
Both inventories and their provenance remain under
`ROOT/monster-retreat-edibles-native-entry-port`.

The remaining direct Go references need semantic classification: several
parser casts reduce successful non-null results to a boolean, some references
identify callbacks that already dispatch to native Go implementations, and
others call raw C entries. In particular secret-wall packet entries `4DF120`
and `4DF180` accept pointers but narrow them internally, and truncate player-
unit traversal pointers. Their existing ordinary Go implementation does not
close those public C paths. Original execution and actual C reproduction
must precede a port; a static match supplies no crash or completion credit.

Runtime `f6672bd1c` ports both secret-wall packet entries to the existing
native wall record type and full native unit/update/player pointers. Raw
live traversal, cached wall WORD, unchanged packet identity, zero return and
original faults remain separate from the ordinary guarded player snapshot.
Seven original references are frozen before engine changes; both original
FPU profiles execute 172 instruction bytes and 1,188 contexts, including
26 faults, with byte-identical outputs. Both actual old C entries SIGSEGV
above 4 GiB after independent native controls. Normal/strict comparisons and
fourteen fresh default-C pipeline children pass in each profile.

The repeated host scan now finds **3,491 active definitions / 297 integer-
parameter cast candidates**. The separate typed-pointer view contains **42
definitions**, retaining eleven overlaps with the integer view. Only
`4DF120` and `4DF180` leave the typed view; all prior integer-candidate
identities remain. All 191 C inputs and headers match clean `f6672bd1c`, with
zero preprocessor warnings. The original working-tree captures retain their
actual provenance under `ROOT/secret-wall-state-native-entry-port`.

At the preceding `f6672bd1c` source, raw wall-destroyed entry `4DF0A0` remains
an integer-pointer candidate.
Its existing ordinary Go counterpart does not port its public C ABI. The
complete 118-byte original body has distinct direct-Send return and recipient-
32 broadcast paths. Both original FPU profiles execute 1,974 contexts,
including 43 faults, with byte-identical complete outputs. This is original
execution evidence only: its native getter/consumer, actual C reproduction
and port are subsequent gates under `ROOT/wall-destroyed-native-entry-port`.

Runtime `c335ee23d` subsequently ports that public entry to a native wall
argument, retaining the signed C-int recipient/result and direct-Send versus
recipient-32 broadcast behavior. The raw player traversal core is shared
with both preceding secret-wall entries. Six references are frozen before
engine changes; the actual old C entry faults above 4 GiB in two routing
modes after native controls. Normal/strict selections each pass 6,363 events,
including thirteen new and fourteen preceding actual default-C children,
without changing the 4,312 prior inputs or six frozen references.

The new hash-bound host scan finds **3,490 active definitions / 296 integer-
parameter cast candidates**. Only `4DF0A0` leaves that view. The typed-pointer
view retains **42 definitions / eleven integer-view overlaps**, with zero
preprocessor warnings. All 191 C source/header hashes match clean
`c335ee23d`; no working-tree capture is retrospectively relabelled.

Engine `b65798bc5` then ports raw Chakram driver `4EB250` and callback
`4EB340`, including native object/point/result pointers and shared pointer
scratch. Original integer class bits replace the old numeric float cast.
Ten references are frozen first; the actual old driver crashes above 4 GiB
after native controls, while the callback reproduces a valid-target class
failure. Complete original execution covers 376 bytes and 272 contexts per
FPU profile, including nine faults. Ordinary native startup rounding also
fails 75 of 99 original comparisons before correction; those failures remain
preserved. Normal/strict selections subsequently each pass 7,618 events,
including fourteen actual default-C production children per profile.

The hash-bound host scan now finds **3,488 definitions / 294 integer-cast
candidates**. Only these two entries leave that view. The typed-pointer view
has **41 candidates / ten overlaps**, with only `4EB340` removed. All other
candidate identities remain and all 191 preprocessing logs have zero
warnings. The manifest's existing 4,043 records remain unchanged. Working-
tree snapshots keep their original source labels and are rebound by exact
C source/header hashes to the committed engine. A clean ARM64 product and
six host E2Es subsequently qualify. Its native CI remains incomplete:
Windows amd64 passes, while 386 terminates in an unchanged original Trigger
nil-fault case; exact failed/passing PE comparison is diagnostic only.
Unacquired macOS 26 jobs retain their original failures, with identical full
native gates requested on an optional standard macOS 15 ARM64 image.

Build, strict-pointer, product and CI results belong in
[porting-completion.md](porting-completion.md), with their exact source
revision. Static inventory changes alone earn no campaign, audio, runtime
stability or overall 100% completion credit.
