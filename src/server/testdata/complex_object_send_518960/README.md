# Original complex-object sender 00518960

These references execute the unchanged GAME.EXE SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The executable and stock assets are not redistributed here.

Phase A contains 14,110 contexts/calls per FPU profile, with warm type
caches. Phase B contains 15,004 frozen contexts and 15,392 calls per
profile, including cold dictionary lookups, ordered lookup mutations,
repeated calls, cached versus live update data, sprite/NPC bytes,
coordinates, direction and network-code mutation. Both use actual
selector, predicate, WORD-name-ID, frame, RNG, converter, health-report and
direction bodies. Network publication is declared; Phase B also declares
dictionary inputs. These are bounded original executions, not whole-engine
or peer-network equivalence claims.

The startup profile uses x87 control word 0x0e3f and truncates signed-DWORD
coordinates toward zero. The 0x023f nearest profile is diagnostic. Integer
observations match; coordinate packet bytes differ in 13,941 Phase A and
15,392 Phase B calls. The native comparisons use the unchanged startup
packets without coordinate normalization. The original FPU control, stack,
global neighbors, memory writes and nonvolatile registers are guarded.
Native comparisons check relocated fields and whole guarded memory bytes;
they do not claim to count native memory stores or reproduce host FP flags.

The static declared scope is 1,192 bytes across 12 bodies, including the
378-byte root and excluding its six alignment NOPs. The sum of maximum
observed byte counts per function across the two phases is 1,189; this is
not a directly instrumented cross-run coverage union. Three bytes in the
58-byte network-code helper's nil branch are unentered by this nonnil
sender. Inline tables contribute 83 bytes, the RNG table 16,384, direction
table 2,048 and threshold four. See the raw proof files for each profile's
actual execution counts and table reads.

The full actual C and direct Go comparators each exercise 29,114 contexts /
29,502 calls with actual native helper bodies and C-owned allocations.
Only dictionary input effects and network publication results are declared.
The actual health helper creates its important packet; the test drains and
compares that native payload. The frozen protocol retains a full int32
enqueue return, while the stock message-list boundary returns bool 0/1.
The public C ABI remains int(int, uint32_t*, int); the pointer is interpreted
immediately as a native Object, with no PE32 indexing or integer narrowing.

Twenty-three separate ordinary modes use actual PlayerDraw index 50
(AnimElectricZap), the default factories, real registry/cache and RNG.
Each runs once through C, the ordinary Go sender and the pure Go packet
method. The packet method intentionally performs neither health reporting
nor network publication. A separate root-package projection checks its
production caller. The 14,110-context network-code test independently
checks the full zero-extended helper result, including unsigned secondary
code bounds on 386. Existing fixtures and expected values are retained.

Reproduction requires Python with Unicorn and a legally held original:

```sh
python original_complex_object_send_phaseA_v2.py --game /path/GAME.EXE \
  --control 0xe3f --fixture /tmp/phaseA.json.gz --proof /tmp/phaseA-proof.json
python original_complex_object_send_phaseB_v1.py --game /path/GAME.EXE \
  --control 0xe3f --fixture /tmp/phaseB.json.gz --proof /tmp/phaseB-proof.json
```

Use separate outputs for 0x023f. Original inputs, outputs and producer
sources were sealed before implementation. Private records retain initial
caller-stack scratch, registry initialization, NaN comparison and manifest
preparation failures; no failed output replaces a reference.
