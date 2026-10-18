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

## Remaining scope

The encrypted integration uses themes without conditional expressions.
`sub_51E670`, `sub_51E800` and `sub_51EAF0` still transport file handles
through integers; class conditions also use legacy player pointers and
offsets. Those paths require separate native conversion and original
comparison. The retained stock map inventory lacks the external
theme/AreaMap inputs required for a complete classic generation run.
Arbitrary-length tokens and out-of-capacity records are outside the
bounded evidence above. Full generation and remaining callback/layout
classification are unfinished.

All 3,249 pre-existing tests, fixtures and YAML files retain their hashes,
as do 1,562 personal installation files totaling 571,159,691 bytes. The
existing manual-save PNG comparisons remain failed pending their separate
fixture decision. These parser checks leave the overall weighted maturity
estimate at 79%; they do not provide a measured feature-coverage denominator.
