# Independent native inventory memory and actual parent controls

Original reference `1d655911f` precedes these controls. The raw checker in
GAME4.c and its two-integer GAME4.h declaration remain byte-identical. The
actual C parent entry and its existing Go implementation also remain exact.
None of the previous three-entry producers/getters/consumers is changed.

`generate_inversion_inventory_initial_4fa4f0.py` reads only the frozen
independent inputs. It reconstructs all 207 complete 64-KiB initial heaps
without executing any original or native body. Every byte agrees with the
actual original pre-state. It independently declares all selected object,
array and modifier words/pointers, plus byte reads and invalid read domains.
The encoded/decoded fixture SHA-256 values are:

```
1b79975ddc7aeac9f60850ada1cd975fdbe105a52556c433e8e35b1e499755a2
6e21535d557aeb9ff7a6b9a7d0b6cd77eb8c2fde6b889f696909a6158bd2117e
```

The new native adapter reuses the frozen whole-heap/projector container,
without modifying it or its earlier constructor. A separate initializer
relocates the new object's inventory/owner/flags fields and all three arrays
and four modifiers. Pointer roles and actual C callback identities are
explicit; unrelated allocation bytes stay guards. Thirteen prior layout
comparisons plus five inventory/owner/flag/class comparisons agree between
C and Go. Every valid field read also agrees with an actual typed C getter.
All original heaps project back exactly after initialization and after reads.

There are 242 actual nil/protected typed-read domains per profile. These
independent initial-field probes include domains skipped by the checker's
admission gates, so this count is distinct from the thirteen original
function faults. Valid selected class-byte and class-word reads are separate.
No arbitrary alias between differently widened structures is asserted.

The actual existing C `4EC580` parent entry is invoked on every one of the
267 argument occurrences observed in the original checker execution. Its
263 pointer returns and four memory faults agree with the actual unchanged
original 38-byte parent body; all native allocation bytes and the complete
original projection remain unchanged after the calls. This adds coupled
input coverage for an already typed helper. It does not implement the raw
inventory checker or replace its original helper with a mock.

The actual raw checker's old two-integer C entry is also executed in an
isolated re-execution of the exact retained legacy main. Complete high native
root/target allocations, typed C/Go reads, layouts, actual typed parent and
whole-allocation guards pass before the raw call. Its truncated root then
produces the expected C pointer fault. No new engine masks that failure.

Go 1.27.3 normal and strict HD (cgocheck2/checkptr=2) each pass 427 events,
including 207 initial-memory leaves, 207 actual-parent case groups, the old-C
fault control and seven existing owner-helper layout/order/cycle tests.
Six actual parent mains are retained before execution. Parent mains and Go
drivers exit zero without forced termination; both expected child faults and
their complete stacks are preserved. Source inputs stay exact through both
profiles. Successful WORK/report/executable duplicates are removed only
after complete byte verification of their archives.

Evidence: `ART/modifier-inversion-inventory-4fa4f0-native-entry-port/independent-initial-heaps-match-original-before-native-v1.json`
and `native-initial-and-actual-parent-normal-strict-qualified-v1.json`.
The complete WIP source is `inversion-inventory-initial-source-v1.zip`;
actual mains/logs are in `local-suite-archives`. Both identical original
reference outputs are retained once in a verified content archive; raw
duplicate copies are removed with a cleanup receipt and the committed
canonical fixture remains exact.

Complete native checker consumers, including the owner's actual memory
accesses and the callback reload/call boundary, must next freeze before any
of the four coupled raw engines changes. Native implementation credit is zero.
