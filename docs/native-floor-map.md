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

### Unresolved selected-region coordinates

The existing version-3 selected-region reader computes `(row+23)/46` for odd
rows and `row/46` for even rows. Original instructions and decompilation use
`23*(row+1)/46` and `23*row/46`. A two-by-two region starting at row 3 therefore
places both layers at `tile[1][0]` in the current product and `tile[1][2]` in
the original-derived expectation. The existing R300 fixture expects the
current mapping. This conversion preserves that expectation and leaves the
coordinate discrepancy open. Its selected-region read case is a preservation
check, not proof of original coordinate equivalence.
