# Original object velocity oracle

`original_objects_velocity_5118a0.py` executes the unchanged complete
`GAME.EXE` root at `005118A0..00511B57`: 695 bytes, SHA256
`2721db862566a7311b9e12cb0a7333e76f42c548665135151a912760eefe5208`.
The executable SHA256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The executable is not distributed here.

Both startup PC53/ToZero (`0xe3f`) and the nearest counterfactual (`0x23f`)
execute all 695 instruction bytes independently. There are 426 contexts;
the fixture records startup results only. No original arithmetic, conversion,
instruction bytes or control flow is replaced. Type lookup, two-pass queue
enumeration, collision preparation, springs, action/buff queries, trace, tile,
collision enqueue, synchronization, collider and map updates are declared
services. Their snapshots, return domains, live arguments and mutations are
checked. This is not a claim about their transitive implementation.

The 132 prefix cases are input-only records from the earlier arithmetic
diagnosis. Their old expected results and the host C calculation are not
used as expectations. All expectations are generated afresh by this script.
Native tests use C-owned objects, health/cache frames, native-width pointers
and complete frame/canary checks. Separate root-package tests compose the
actual queue, collision preparation, springs, map trace, synchronization,
collider and map index for moving/frozen objects through Go and the public
C ABI. These two tests include original preparation's acceleration reset.

## Required emulator calibration

Unicorn 2.1.4 leaves a signaling NaN signaling after `FLD m32fp`. Intel's
masked-invalid behavior quiets it during the load. Among the 424 contexts
that preceded the two native composition cases, this changes four of the
24 mixed-NaN cases; the other 420 contexts retain their prior results.
The generator rejects that uncorrected load behavior before collecting any
original instruction coverage.

Apply `unicorn_flds_masked_snan.patch` to Unicorn 2.1.4 and build its x86
shared library. The patch only quiets the loaded ST0 value and records #IA.
It deliberately leaves FT0 conversion for arithmetic memory operands alone;
prematurely quieting those operands changes NaN selection. Neither the
production Go arithmetic nor the original instruction stream is changed to
accommodate the emulator. The fixture includes the patch SHA256
`183d9b99af3f1eec4f6a951a8a7e6c314ea7615a4ebdc66f02697e3764d07e12`.

This correction follows Intel's [FLD instruction reference](https://cdrdv2-public.intel.com/812383/253666-sdm-vol-2a.pdf)
and [masked invalid/NaN rules](https://cdrdv2-public.intel.com/671436/253665-sdm-vol-1.pdf).
An independent Rosetta x86-64 FLD probe confirmed quieting, payload/sign and
#IA before generating the corrected fixture. Rosetta is not native x86
hardware evidence. `TestObjectsVelocityOriginalX87NaNLoad5118A0` checks the
actual FLD instruction on native x86 CI and restores the x87 environment.
ARM64 skips that hardware-specific test.

```sh
git clone --depth=1 --branch 2.1.4 https://github.com/unicorn-engine/unicorn.git unicorn-src
git -C unicorn-src apply /absolute/path/to/unicorn_flds_masked_snan.patch
cmake -S unicorn-src -B unicorn-build -DUNICORN_ARCH=x86 \
  -DUNICORN_BUILD_TESTS=OFF -DUNICORN_INSTALL=OFF \
  -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release
cmake --build unicorn-build --parallel 2
LIBUNICORN_PATH=/absolute/path/to/unicorn-build python3 original_objects_velocity_5118a0.py \
  --game /path/to/GAME.EXE --actual regenerated.json \
  --nearest nearest.json --proof proof.json
cmp regenerated.json objects_velocity_original_5118a0.json
```

Install the matching Unicorn 2.1.4 Python binding first. On macOS the oracle
requires JIT execution permission. The earlier uncorrected fixtures, native
C counterfactual failures and the four-context correction delta are retained
in the external validation archive; they are not overwritten.
