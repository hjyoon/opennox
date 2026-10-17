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

The controlled services do not establish original `GAME.EXE` equivalence or
full map-loading coverage. The version-3 writer and region-free reader still
contain older 32-bit tile/node arithmetic and need separate conversion and
coverage. This change addresses the native region argument and its directly
related scalar variables.
