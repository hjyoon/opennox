# Native classic map-generator theme reader

R310 fixes the theme-reader pointer transport used by the classic `#name`
generator entry. The native entry accepts a `const char*` name, retains the
actual `FILE*` for crypt setup and dispatches algorithm, spell-set and exit
records with native theme pointers. Existing integer entry declarations
resolve registered legacy tokens. Callers holding native pointers use
`mapgen_theme_native.h` before any pointer-to-integer conversion.

The algorithm value local is restored to its original 60-byte array.
Skeleton selection uses `getMemPtr` to read native side slots at original
four-byte offsets. It preserves the packed PE32 table and the original
`NONE`, `HALL_RING` and unknown-name selection behavior. Theme storage stays
1,116 bytes; scalar offsets, parser order, defaults, return values and
partial writes are preserved.

## Original evidence

The retained REA/Ghidra instructions, pseudocode and complete-body byte
receipts were independently matched to the immutable local `GAME.EXE`,
SHA-256 `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The original executable was neither executed nor modified by this work.

| Original routine | Address | Sealed bytes |
| --- | --- | ---: |
| Theme reader | `0x0051E260` | 734 |
| Algorithm-data reader | `0x0051EBB0` | 1,020 |
| Spell-set reader | `0x0051EFB0` | 120 |
| Exit reader | `0x0051F800` | 495 |

The oracle adds these four code records and three data records: the
12-byte skeleton table and its two string literals. Every previous record
is retained. `noxoracle code-verify` passes 2,999 code and 659 data ranges.

## Executed checks

`TestMapgenThemeNativeOriginalContractsR310` executes complete production
parser bodies with the actual pointer-token registry and controlled file,
token, formatter, lookup and settings callbacks. Each of O0, O2 and
ASan+UBSan passes 236 cases covering native pointers and integer adapters,
open/crypt/settings outcomes, dispatch, scalar offsets, float conversion,
boolean and skeleton selection, spell count boundaries, four exit
directions and partial EOF writes. Guard bytes and packed skeleton slots
remain intact. Input values stay within their original storage capacities;
the 256-byte path buffer is checked with a 244-byte theme name and the value
buffer with tokens up to 59 bytes.

`TestMapgenThemeNativeEncryptedBinfileR310` creates encrypted themes in a
test-owned temporary directory and uses the actual CGo/binfile implementation.
Its four cases exercise a native name, a registered high-bit name token,
the 244-byte theme-name boundary and premature ambient-light EOF. On ARM64,
the theme, name, file handle and allocated room/hall pointers are all above
`UINT32_MAX`. It checks original fields and frequencies, file closure,
graph cleanup, surrounding bytes and unchanged packed skeleton data.

The two tests contribute nine passing parent/subtest events. Four full
Darwin ARM64 suites pass at clean code revision
`98080022714fdfa52fb03acf22cfec0bc6c0bdc8`; all previous terminal outcomes
and multiplicities remain. The focused mapgen union passes 157 events with
race detection, checkptr level 2 and `GOEXPERIMENT=cgocheck2`. Four product
builds and help commands pass; REA `inspect-go-binary` agrees with Go's own
build metadata on architecture, revision, configuration and clean status.

```sh
NOX_E2E_SEAT=headless ./scripts/go.sh -C src test ./legacy \
  -run '^TestMapgenThemeNative.*R310$' -count=1
```

## R311 conditional and generation boundaries

Conditional dispatch now passes the actual encrypted `FILE*` to native
condition and comparison readers. Class selection follows native player
links and `info.playerClass`; it no longer reads the PE32 class offset from
an ARM64 player. The integer entries remain adapters for registered legacy
tokens. Original comparison order, the fixed experience-level rule,
percentage RNG, partial output and IF/ELSE/ENDIF recursion are retained.
The ordinary `#name.map` entry also extracts the theme from the requested
map name, rather than slicing the unrelated installation directory string.

`TestMapgenConditionNativeEncryptedR311` exercises 270 cases using actual
encrypted binfiles and native player slots. These include all three classes,
inactive players, player counts, experience comparisons, percentage endpoints,
invalid operators, premature EOF, nested conditionals and complete conditional
themes. Native file/player addresses exceed `UINT32_MAX` on ARM64, and legacy
entry cases force high-bit tokens. Output guards, player bytes and file closure
are checked. This runs in an isolated child so parser and RNG state do not
change existing tests.

The generator entry also retains its detached server-object list through
native side slots and clears typed `script_id` fields. Object placement uses
native object fields and pointers for coordinates, flags and creation. Map
saving passes its native compression destination directly and copies the
four-byte extension with `memcpy`, avoiding an unaligned integer store.
`TestMapgenEntryNativeObjectList` exercises empty and populated lists;
`TestMapgenSaveMoveNativeR311` exercises 48 save and 18 placement cases.
Each complete production body runs with O0, O2 and ASan+UBSan, preserving the
original bounded path inputs, call order, failure branches and untouched bytes.

The R313 full-generation probe then exposed a separate cleanup defect: the
8,192-byte work buffer was truncated into a four-byte global, and cleanup read
eight bytes across that global and its neighbor before freeing the result.
Allocation now registers the native pointer in the existing mapgen registry;
cleanup resolves and forgets that token, frees its allocation and clears the
four-byte slot. Registration failure releases the new allocation. The complete
source audit found only these allocation/free consumers and the unchanged
four-byte variable metadata. `TestMapgenBufferNativeR313` uses the actual
registry with both allocator failures, native addresses above `UINT32_MAX`,
adjacent guards, 12 repeated lifetimes and repeated cleanup. The extracted
baseline fails its pointer-identity contract in all three compiler modes;
the corrected bodies pass O0, O2 and ASan+UBSan.

REA/Ghidra records and independent PE reads cover seven further complete
original bodies: condition dispatch (164 bytes), condition evaluation (714),
comparison parsing (177), generation entry (433), generation step (688), map
save (298) and object placement (240). The last body spans two new disjoint
ranges and the preserved six-byte Door callback comparison. Every previous
oracle record is unchanged; the combined oracle verifies 3,007 code and 659
data ranges. Whole-body byte coverage does not imply complete runtime coverage.

## R313 controlled generation, save and load

A dedicated ARM64 product built from the same 6,611 compiled-source,
test and module files as the integrated tree completes the ordinary-clock
`#probeclassic` path. An authored encrypted theme selects a fixed seed, one
room, no recursive halls, no prefabs and no campaign exits; its valid empty
`AreaMap.lib` index and stock wall/floor names are explicit inputs. All runtime
files live in a separate owned data copy, with audio disabled.

The API observes `so_beach` -> `#probeclassic` -> `$probeclassic`. Generation
writes a 2,440-byte map and a 2,516-byte NXZ file, then the normal loader reads
its 16 named sections and final terminator, including WallMap, FloorMap and
ObjectData.
The dedicated process exits naturally with code 0 after its ordinary quit
command; capture cleanup verifies the owned process group and removes its
temporary root without termination. REA receipt:
`ev_1cc469a2f530e147544dd7a01ef9f9f832708768932a87d204115eb2f088bcff`.

The same immutable product also generates and loads a `NONE` skeleton with
recursion limit 3 and a `HALL_RING` skeleton with recursion limit 1. Their
saved maps have 22 walls/58 floor halves and 258 walls/960 floor halves,
respectively, compared with the initial map's 16 walls/49 floor halves.
These counts come from the actual saved-map parser; transient room counts
were not instrumented. Both processes exit naturally with code 0. The actual
NXZ decoder also reconstructs all 2,440 bytes of the initial map exactly from
its compressed file. The read-only decoder receipt is
`ev_f72202b5b42afb1817bc38a894f62a9a3fd404352dd4b9839878c0c38497a9cb`.

The three initial runs use a source-bound development artifact. A subsequent
clean server product at code commit
`0eaa6e6d5a606aeace14d06d88697f1e70bdbda8`, SHA-256
`eaac8153ab2740b4ca92d65db9dcd54adf1a6c592055ff9e977c19707d953780`,
also completes the initial generation/save/load scenario and exits naturally
with code 0. Its generated map and NXZ bytes exactly match the initial run.
REA receipt:
`ev_f18e248ce22c30ceb14f10a0d408ee43250cc281bc5f764c4bcb0d6634a3e5af`.
The two broader scenarios remain separately identified development runs.
The earlier exact-current abort and private diagnostic trace remain retained
as the negative controls. Captures retain sampled process/environment limits;
they do not supply an atomic OS image binding or complete process trace.

## R314 dedicated progress and saved exits

The clean R313 dedicated product generated and loaded three authored exit maps,
then panicked in the fourth generation when the throttled progress callback
reached the client renderer. A dedicated server has no renderer. The corrected
`5235F0` progress entry returns immediately under `EngineNoRendering`.
This is an explicit dedicated-server port extension; the original only has its
editor-mode guard. Generation, ordinary time and RNG remain active.

REA/Ghidra and an independent PE read agree on all 125 original bytes at
`005235F0`, SHA-256
`65f78187da6533b2ad93e693d30d9b11715cdee4b6470ae716c04f63ef9aa3a6`.
Both original `JLE` branches use signed 32-bit time comparisons. The rendered
path now retains those comparisons and the exact 250ms strict interval.
Its three-byte opcode/low-16 sequence packet uses byte stores, removing a real
UBSan unaligned-word failure while retaining the full DWORD increment before
the callback and live timestamp reload afterward. The additive whole-body
fixture exercises rendering, editor and dedicated paths in O0, O2 and
ASan+UBSan, including rollover, sequence wrap and callback/input mutations.
The oracle now verifies 3,008 code and 660 data ranges; every preceding record
is unchanged. Original GAME.EXE was inspected statically and never executed.

Clean code `dd142aa82dcbcc0061683c99b3ee4c44b5ae484e` produces a dedicated ARM64
server with SHA-256
`8e9b5c228a3831669fa8fe4d573c378dc5d35f5d0a61716b5b32be7968112dff`.
In one ordinary-clock process it generates, saves and loads north, south, east
and west authored themes, then regenerates north after an intervening stock
map. All five generations complete and the process exits naturally with code 0.
The repeated north map and NXZ bytes match. Existing `set maps allow.all true`
is explicit preparation for repeated synthetic map selection after entering
coop mode; it does not supply generator state. Both earlier negative captures
remain separate. Final generation receipt:
`ev_492d9622ac957da02f304e9e7410c350963d0fa849774749c06e98a706486ad1`.

The themes use real stock `Exit1`, exact `so_beach.map` link data, a nested
false Wizard branch, zero-player equality and 0%/100% alternatives. Each saved
map is 2,512 bytes and contains one PlayerStart and one Exit1. Read-only section
and XFER decoding confirms all 16 named sections, the exact destination and
the corresponding rotated boundary position. Actual `cnxz.DecompressFile`
reconstructs every map byte from each of the four NXZ files without modifying
the inputs. Decoder receipt:
`ev_5a2743797c184acae30ed3134b2d9e6635f68557a9398a040a184745cbe90d99`.
These are authored generation/serialization checks, not a runtime exit collision
or destination-map handoff, prefab workflow or original external-theme oracle.

## R315 CRT stream and required prefab

A required authored AreaMap record loaded and analyzed successfully but failed
every placement attempt: its room coordinates were about 63 million and
86 million, outside the theme. The old C ABI shim returned the low 32 bits of
Go's unrestricted `RandInt`, while both classic range helpers expect the
original CRT's `0..32767` result. A private diagnostic product retained the
successful record index, dimensions and marker analysis before `within=0`.
The failing clean-product captures remain separate from that diagnostic.

REA/Ghidra and an independent PE section read agree on all 253 bytes of
`00402000`, `0040200D`, `00526AB0`, `00526AC0`, `00526B00` and `00526BC0`.
The CRT updates its unsigned dword with `214013 * state + 2531011` modulo
2^32, then returns `(state >> 16) & 32767`. The native C ABI now preserves
that stream and seed width. Its seed shim still calls the existing Go platform
hook exactly once with the unchanged unsigned seed; the Go platform generator
itself is unchanged. The original thread-local game-thread stream is represented
as one mutex-serialized stream across CGo callbacks. Separate concurrent CRT
thread states and arbitrary-input range-helper equivalence are outside this
correction. Original GAME.EXE was inspected statically and never executed.

Six added code ranges bring the oracle to 3,014 code and 660 data ranges;
all previous records are unchanged. The original binary32 scale at `00583080`
was already covered by the data oracle. Additive tests verify fixed arithmetic
vectors, repeat seeding, concurrent stream consumption and the actual C integer
and floating mapgen boundaries. Seven focused events, including the unchanged
fixed-seed forwarding test, pass normally and with race detection, checkptr
level 2 and actual `GOEXPERIMENT=cgocheck2`.

Clean code `0666df6e687bffb82986dd6356c3607020130f0d` produces the dedicated
ARM64 server with SHA-256
`daa2a07e5d34d1063ed3153bbe2c0ca814cf6e9ab623f0ca054e777a381e6ef2`.
The same south-marker library and encrypted theme that failed with the old
product now complete two ordinary-clock generations, saves and generated-map
loads, with a real intervening stock map. Both MAP and NXZ hashes repeat exactly
and the server exits naturally with code 0. The authored 176-byte AreaMap
contains one real stock Exit1 and one ExitSouthMarker; the theme requires it
with `MUST_OCCUR` and contains no separate `EXIT` block. Runtime receipt:
`ev_4dbd70f03e067bf47e94659d6ac4a71b300f9405696f99a70e9a81ff30d286fd`.

Read-only decoding finds 16 named sections, 18 wall records, 81 floor halves,
one PlayerStart at `(3739,3577)` and one Exit1 at `(2911,2657)` with exact
destination `so_beach.map`. Both corrected collection snapshots after load
agree with the saved positions; the connection marker is absent. The actual
native NXZ decoder reconstructs all 2,856 MAP bytes exactly from the 2,952-byte
NXZ without changing either input. Decoder receipt:
`ev_65f06173d7f0bc7df0df9497f83b62c6fc001919df6ebddf9952b51ccd375850`.
This bounded record contains no authored prefab floor, wall, waypoint, group
or script section. Other marker/room combinations and original external
AreaMap assets remain unverified.

## R315 actual exit movement

Four separate fresh solo-host processes also pass actual right-mouse movement
from the generated PlayerStart through a loaded Exit1 to ready `so_beach`.
These use the separately pinned clean R314 code
`dd142aa82dcbcc0061683c99b3ee4c44b5ae484e`: north is normal; south, east and
west are highres. All use ordinary time and exit naturally with code 0 in
about 148 seconds each. StaffWooden, StreetSneakers, StreetPants and StreetShirt
each pass server/client inventory counts before and after the handoff. No
direct callback, contact helper, relocation or measured map command supplies
the transition. Initial source-map selection is explicit preparation.

The north observer saw an early API map-name update before destination
placement, so its separate native MAP READY and inventory assertions establish
readiness. Highres collection snapshots wait for MAP READY and retain item
type/script IDs, equipped flags and player name/class. The handoff conclusion
combines actual input/load observations with the static ExitCollide binding;
no instruction-level collision trace was captured. Receipts, in N/S/E/W order:
`ev_b41e9110646455215411915a82da6d0af75224767d80dede40021ff4cff6e907`,
`ev_d18dec3d5a67d8f83667d509446b5f2cf2c73edc426914db082ba86b747eaf53`,
`ev_8aedb97030338f05f741e0e1ae32877238cc60f1bcbde7df1bf70616ae19dac1`,
`ev_2087ba6e9b51d4321c40866f695a4007399a66d1dc50d52f728dc26aec0aa7ed`.
This is four particular authored solo-host routes, not every product/direction
pair, independent-peer handoff or completed campaign route.

## Remaining scope

The retained stock map inventory lacks original external theme/AreaMap inputs.
Broader room, recursive-hall, prefab-section and player-dependent conditional
combinations remain outside the bounded authored variants above.
Arbitrary-length tokens and out-of-capacity records are outside the
bounded evidence above. Remaining callback/layout classification is unfinished.

All 3,270 pre-existing protected tests, fixtures and YAML files retain their
hashes in the independently recovered workspace repository. All 1,562 personal
installation files totaling 571,159,691 bytes remain unchanged. The existing
manual-save PNG comparisons remain failed pending their separate
fixture decision. These parser checks leave the overall weighted maturity
estimate at 79%; they do not provide a measured feature-coverage denominator.
