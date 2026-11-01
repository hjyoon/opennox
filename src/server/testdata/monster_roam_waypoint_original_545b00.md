# Original waypoint entry comparison

`monster_roam_waypoint_original_545b00.json.gz` contains unchanged GAME.EXE
execution results for five public entries, frozen before the native port:

| Entry | Original code bytes |
| --- | ---: |
| Record 545B00 | 89 |
| Previous 545B60 | 80 |
| Valid 547EE0 | 44 |
| DeadEnd 545BB0 | 161 |
| Next 545C60 | 319 |
| Actual validity child 579EE0 | 23 |
| Actual RNG child 415FA0 | 67 |

The original executable SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
All 783 instruction bytes execute without instruction or arithmetic
substitution. The 9,220 contexts and seven faults produce identical outputs
under x87 controls 0xE3F and 0x23F. Name lookup, diagnostic print and pop-action
are declared world services. The actual original RNG executes, including its
4097-word read-only table boundary; 546 calls retain arguments, returned
values and final indices.

The fixture gzip SHA-256 is
`1873692bde6653cfba29094021513f6bcbc59091390d5f48d563af4d6784bded`;
its decoded SHA-256 is
`08657d083b0e6d772d7e8075234b2c59487812ea0208ef69becd48ae54a3b86e`.
Earlier 7,192-context/two-fault and 9,220-context/seven-fault outputs remain
retained; adding RNG observations changes none of their existing fields.

The native consumer compares actual public C and typed Go calls, complete
native records, padding and surrounding guards, ordered services, pointer
identity, RNG call bounds and state, and fault mutation prefixes. Ordinary
RNG indices 0 and 4095 also execute the production prand implementation.
Raw original index 4096 is an explicit one-past-table diagnostic supplied
through the observed RNG service; it is not native Rand.Reset coverage.

The domain includes history rotation, duplicate removal, unsigned increment
wrap, signed Previous cursor -1, all connection counts 0..32, duplicate RNG
choices, all 24 action stack slots, upper flag/code words and nil faults.
Malformed accesses outside these declared record/array domains receive no
equivalence credit. DeadEnd reuses the low byte of its first by-value stack
argument; both original ROAM callers immediately discard those argument
slots. Other caller bytes remain unchanged. This scratch observation does
not impose a blanket caller-stack immutability assertion.

Full drivers, independent inputs, code/data seals, two-FPU outputs, caller
inspection, pre-edit consumer hashes and five actual C-before SIGSEGV logs
with valid above-4GiB native controls are retained under
`ROOT/monster-roam-waypoint-native-entry-port`. Previous C bodies remain
under `#if 0`. This is bounded entry evidence, not a whole-port completion
percentage or complete campaign audit.
