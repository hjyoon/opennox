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

## Remaining scope

The retained stock map inventory lacks original external theme/AreaMap inputs.
Broader room, recursive-hall, prefab, exit and conditional-theme combinations
remain outside the three fixed-seed authored generation runs.
Arbitrary-length tokens and out-of-capacity records are outside the
bounded evidence above. Remaining callback/layout classification is unfinished.

All 3,249 pre-existing tests, fixtures and YAML files retain their hashes,
as do 1,562 personal installation files totaling 571,159,691 bytes. The
existing manual-save PNG comparisons remain failed pending their separate
fixture decision. These parser checks leave the overall weighted maturity
estimate at 79%; they do not provide a measured feature-coverage denominator.
