# Native floor maps and brush state

`nox_server_mapRWFloorMap_422230` and its version-3 delegate
`nox_xxx_tile_422C10` receive a native `void*` region. Receiving that argument
as `int` truncated a valid allocation above 4GiB before the bounds reader
accessed it. The Go wrapper now calls the declared C function directly, so
CGo checks the argument type when generating the call.

Region pointers remain separate from signed row deltas, row counts, byte
offsets and column indices. File fields retain their existing widths.

## Regression coverage

`TestFloorMapReadNativeRegionR300` uses the real Go wrapper, C implementation
and `cryptfile` service with C-owned regions above 4GiB. Version 3 and 4 run
in separate child processes. Both must complete, preserve allocation canaries
and consume exactly their section bytes without consuming a trailing sentinel.
The original code crashes in both children.

`TestFloorMapNativeRegionDataR300` compiles four complete current production
functions against controlled I/O, allocation and geometry services. It checks:

- A version-4 read with a negative row delta and a native subtile node.
- A version-3 region read containing two columns of scalar tile data.
- A version-4 region write with two native linked subtile nodes, exact output
  bytes and the original region passed to the geometry service.

Run these tests from the repository root:

```sh
NOX_E2E_SEAT=headless ./scripts/go.sh -C src test ./legacy \
  -run '^TestFloorMap.*R300$' -count=1
NOX_E2E_SEAT=headless ./scripts/test-cgocheck2.sh \
  -race -gcflags=all=-d=checkptr=2 -run '^TestFloorMap.*R300$' -count=1 ./legacy
```

The controlled services do not establish full map-loading coverage.

## Version-3 tile rows and links

The version-3 writer and region-free reader now index native tile rows and
follow typed subtile links. They retain the original 128-row scan bounds,
four 32-bit header fields, one-byte flags/counts, 16-byte scalar transfers,
row-major order and geometry callback order. Native pointer fields are never
serialized. Four independent crashes were reproduced before this conversion.

`TestFloorMapV3NativeDataR302` compiles the complete production codec and
bounds function against controlled services. Seven cases cover both writer
forms, both reader forms, two-layer linked nodes, the `[127][127]` edge, empty
grids, exact bytes/callback order and allocation canaries. The seven cases
pass optimized and ASan+UBSan execution; the original R300 cases also pass.

```sh
NOX_E2E_SEAT=headless ./scripts/go.sh -C src test ./legacy \
  -run '^TestFloorMap.*R(300|302)$' -count=1
```

REA/Ghidra 12.1.3 static analysis used `GAME.EXE` SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The codec body `00422C10..004235B6` and allocator body
`00422160..004221D1` are sealed in the existing code oracle. The original
executable was neither run nor modified; decompilation is derived evidence.

## Original region coordinates and bounds

The version-3 selected-region reader uses the original `23*(row+1)/46` for
odd rows and `23*row/46` for even rows. The earlier expressions omitted the
factor 23. A two-by-two region starting at row 3 now places both layers at
`tile[1][2]`. R300/R302 expected rows follow the original instructions;
their input streams, assertions, allocation guards and I/O checks remain.

`sub_428170` views the descriptor coordinates as signed 32-bit values. Its
earlier unsigned view reversed mixed-sign endpoints before the clamps. It
retains the original asymmetric rules: clamp negative minima to zero and
maxima at or above 5888 to 5887, without normalizing inverted results.
`TestRegionBoundsSignedOriginalR303` and
`TestRegionBoundsSignedCGoR303` cover 20 literal original-derived vectors,
including mixed signs, equal endpoints and integer extremes. The latter
uses the existing CGo entry with native C-owned input/output and canaries.

## Whole-byte flags in version-4 bounds

The version-4 writer considers every nonzero flag byte when finding bounds,
matching the original four byte comparisons. The earlier `flags & 3`
condition excluded high-bit-only cells admitted by version-3 input. The
record payload still selects its two layers with bits 1 and 2.
`TestFloorMapV4WholeFlagBoundsR303` passes actual version-3 input through the
version-4 writer and checks exact headers, payloads, callback widths and
sentinels. It covers high flags, both layers, the grid edge and empty grids.
A layer-clear fixture applies the byte operation directly; it does not
establish execution of the separate legacy editor routine.

The flag field is already `uint8_t`; its three neighboring bytes are separate
fields. `TestFloorMapV3FlagByteR303` checks preservation of those neighboring
fields with nonzero byte patterns.
A read-only scan of the hash-bound 157 stock maps found only version 4 and
flags 0 through 3. The high-bit difference concerns admitted custom/legacy
version-3 input, not a reproduced stock-map failure.

The original `FloorMap` dispatch record, its name and the decoded
`00422230..00422A34` code interval are also sealed in the code oracle. REA
decoded the entry and read matching bytes; LLVM supplied sequential decode
and boundary evidence. Ghidra did not register this interval as a procedure,
so its extent is recorded as a derived code interval rather than a Ghidra
function body.

```sh
NOX_E2E_SEAT=headless ./scripts/go.sh -C src test ./legacy \
  -run '^Test(FloorMap|RegionBounds).*R30[023]$' -count=1
```

## Native brush, layer ownership and flood coordinates

The conditional classic map-generator floor path now passes a pointer to a
contiguous 32-byte `nox_floor_brush_t` scalar record. Its tile, variation,
pattern, anchors, kind and edge fields retain their original offsets. The
record contains no pointers; the argument carrying its address uses native
pointer width. Separate C locals no longer stand in for the original stack
record in `sub_51D8F0`.

Native producers call `nox_floorBrushApplyNative_51DA70` through the dedicated
brush header. The existing `sub_51DA70` declaration and `GAME4_2.h` remain
byte-for-byte unchanged. Its compatibility wrapper resolves the unsigned
32-bit legacy token before calling the native entry; it never treats that
token as a native address. Separate null-token and high-bit-token cases cover
this adapter.

The writer selects scalar fields and the owned subtile head separately. A
native cell is 56 bytes: left scalars begin at 4 with their head at 24; right
scalars begin at 32 with their head at 48. Casting `cell+4` to a standalone
layer would put the left head at the wrong address. Subtile nodes are 24 bytes
with a native `next` pointer at 16. Tail removal, duplicate return values,
merge/coalesce order, byte flags, adjacent fields and the other head remain
covered by controlled cases.

The flood worklist retains three 32-bit scalars per entry, its 500-entry
limit, duplicate checks, LIFO removal and original neighbor order. Its y
coordinate is a scalar; it is no longer read into a native pointer slot.
The grid scale uses the exact original binary32 operand `0x3CB21643`, promoted
into the existing double intermediate. No valid-grid output difference was
observed for the earlier decimal scale, so this correspondence improvement
is not counted as a reproduced additional failure.

`sub_5245A0` and `sub_524610` pass explicit two-float coordinate arrays. Their
half-step and step use original words `0x41821B8F` and `0x42021B8F`, matching
the original `FADD float` operands and per-coordinate stores. Truncated
decimal double literals produced different results for negative/tie vectors
even after the array conversion. Eight cases verify exact binary32 callback
coordinates/order, nonpositive dimensions, input preservation and the
retained C return declaration. Ghidra infers void returns for the original
two routines; return checks do not claim a defined original return value.

`sub_543680` originally initializes only brush offsets 0, 4, 24 and 28. The
remaining original stack state is unspecified. The portable record now
initializes its remaining transient fields to zero. Two explicit determinism
cases cover that policy separately from 33 original-derived native cases;
they do not establish equivalence to unspecified original stack contents.

The 37 native cases and eight coordinate cases pass optimized C and
ASan+UBSan execution. The integrated floor/region/brush union passes 124
terminal Go test events. Together with the seven unchanged declaration
guards, 131 events pass normally and with race, checkptr level 2 and
`GOEXPERIMENT=cgocheck2`. The new source-bound C cases use controlled callbacks
and native structures; they do not execute a complete classic map generator
or directly exercise its Go/CGo boundary.

REA/Ghidra decoded 14 complete original procedure bodies for this path,
including the two coordinate producers. Independent PE reads match every
body and referenced constant; all sessions closed and the original file hash
remained unchanged. The oracle retains its previous records and now verifies
2,988 code ranges and 655 data ranges.

The `#name` server entry needs external theme/AreaMap inputs absent from the
retained stock inventory. Upstream theme, file-handle and conditional-parser
pointer conversions remain open. This work does not establish a successful
complete stock or custom map-generation run.

```sh
NOX_E2E_SEAT=headless ./scripts/go.sh -C src test ./legacy \
  -run '^TestFloorBrush.*R304$' -count=1
```

## Fill producer coordinate pairs

`sub_524950`, `sub_5249C0` and `sub_524B50` now pass contiguous two-float
coordinates to rectangle, column and flood callbacks. Their existing signed
32-bit fill token still resolves through `nox_mapgenLegacyPtrResolve`; the
native pointer returned by that registry remains intact.

The producers use original binary32 words `0x41821B8F`, `0x42021B8F` and
`0x42821B8F` for half-step, step and double-step. These operands are promoted
into the existing double intermediates before the original float stores.
In `5249C0`, the first pass uses the unrounded double offset while the second
reads its stored float copy. With width 9, height 2 and x origin
`-97.58073425292969`, the first pass starts at `2^-18` and the second at zero.
Replacing both paths with one float intermediate would lose that distinction.

Fifteen fixed original-derived vectors cover callback order and coordinate
bits, signed division, zero/negative dimensions, conditional skips, exact RNG
calls and ranges, native fill/input canaries, and null tokens. The controlled
program links the actual production token registry. A low-bit collision
forces a high-bit token for a live fill allocation above 4GiB; signed-int
transport must resolve back to that exact native address.

Using identical inputs and expectations, the baseline passes six vectors and
fails nine, including nine ASan stack-buffer-overflow reports. Converting
only the coordinate arrays passes nine and leaves six numerical mismatches.
The final candidate passes all 15 optimized and ASan+UBSan vectors.
`TestFloorBrushFillPointsOriginalR305` repeats both modes from complete current
production bodies, typedef, resolver and registry. It adds 33 terminal Go test
events on macOS, including mode and parent events. Together with prior floor
checks and the unchanged declaration guards, 164 events pass normally and
with race, checkptr level 2 and `GOEXPERIMENT=cgocheck2`.

Optimized cases run on every supported 64-bit target. Sanitizer mode is
required on Darwin/Linux and skips on other targets, avoiding an additional
runtime dependency in the existing Windows MinGW gate. The macOS runs skip
neither mode. Source-bound callbacks do not establish a complete generator
run or a direct Go/CGo call into these producers. Integer return checks retain
the current C declarations; Ghidra infers void for two original routines.

REA/Ghidra complete bodies for these three functions total 1,168 bytes and
match independent PE reads. The oracle adds those three bodies and the
double-step constant, reaching 2,991 code ranges and 656 data ranges. Every
previous record and its relative order remains. The original executable
remains unchanged and unexecuted.

```sh
NOX_E2E_SEAT=headless ./scripts/go.sh -C src test ./legacy \
  -run '^TestFloorBrushFillPointsOriginalR305$' -count=1
```
