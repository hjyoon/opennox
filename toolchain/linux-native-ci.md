# Native Linux qualification

`port-linux-native.yml` runs on an Ubuntu AMD64 machine with the pinned Go
toolchain and native GCC. It builds, verifies and runs help for the selected
product, then tests every package with that product's tags. SDL and OpenAL use
headless backends for tests. Game assets and campaign replay are separate.

Product builds and help use the default link mode and allocator. Full native
tests use PIE and `MALLOC_MMAP_THRESHOLD_=0` so their existing C static-region
and C-owned allocation checks actually exercise addresses above 4GiB. Fixed
low-address cases, numeric checks and native fault checks are retained. The
server job also runs the damage regressions with the default product link mode.
The allocator setting is passed through `go test -exec` to test processes and
their children. Compilers and linkers use their default allocator. Each test
profile is recorded in the artifact; it is not a product setting.

The four variants are `opennox`, `opennox-hd`, `opennox-server` and
`opennox-server-hd`. JSON test output, expected and observed package lists,
failures, compiler/build metadata and executables are preserved as artifacts.
A missing package result fails the evidence check. Artifact retention is seven
days; keep qualification artifacts locally when archiving a checkpoint.

A push to `port/go1.26-multiarch` changing the workflow or
`linux-native-ci-checkpoint.txt` runs all four variants. Update the checkpoint
file when intentionally qualifying a new native source checkpoint. Routine
campaign scenario changes reuse prebuilt products.

The manual workflow defaults to the server variant. To request all variants:

```sh
gh workflow run port-linux-native.yml \
  --ref port/go1.26-multiarch -f products=all
```

Native CI proves the reported build and test scope. It does not certify an
interactive Linux desktop, audible output or full campaign completion.
