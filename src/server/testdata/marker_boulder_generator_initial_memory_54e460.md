# Initial memory for the three raw death entries

`generate_initial_memory_54e460.py` independently describes the initial
64-KiB PE32 heap for all 200 frozen Marker/Boulder/MonsterGenerator inputs.
It reads no native implementation and executes no function body.

The fixture includes the complete initial heap digest, source and returned
parent read/fault domains, six object-role records, both potential marker
slot allocations, returned-parent slot domains and initial world scalars.
Pointer words retain their original PE32 meanings, including nil and the
unmapped domain; these are not native addresses.

Encoded SHA-256:
`6cb297188ca30a429f2414b80ba455710e3b078bca25a8dd10325a1108dd9570`.
Decoded SHA-256:
`a2a4eec4e3b90b0dfd4b75606836bc8945c082151bbdf624084408bbfdc0b503`.
The standalone producer reproduces these bytes exactly.

A separate control reuses the sealed original producer's complete initial
setup in Unicorn. All 200 actual emulated heaps and independent read/fault
domains match this fixture before any raw body executes. Only the two FPU
initialization instructions execute in that control. The original executable
and setup source hashes are recorded in its proof.

Evidence is retained under
`marker-boulder-generator-native-entry-port/three-entry-initial-memory-preparation-v1`
in the port artifacts. See `actual-original-initial-memory-before-native-v1.json`
and `three-entry-independent-PE-initial-memory-preparation-qualified-v1.json`.

Native relocation, complete C/Go allocation guards and protected fault
domains still need independent verification. Complete output consumers must
then be compiled and frozen before implementing the three native raw bodies.
This initial-memory fixture supplies no native implementation, original body
execution, reentry, or native whole-function comparison credit. Existing
original inputs, two-profile outputs, producers and C bodies remain exact.
