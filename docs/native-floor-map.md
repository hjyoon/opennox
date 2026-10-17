# Native FloorMap region boundary

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
