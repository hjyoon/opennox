# Complete raw inventory consumers before the four native engines

Original reference `1d655911f` and independent native initial controls
`8a1e94d1d` precede this preparation. All four old C bodies/prototypes and
the earlier three-entry reference/getter/consumer files remain unchanged.
Use both `OPENNOX_MODIFIER_PREDICATES_BEFORE=1` and
`OPENNOX_INVERSION_INVENTORY_BEFORE=1` for this preparation. The native
consumer branch requires all four native C prototypes; it receives zero
native-engine leaf credit until those engines actually execute.

The complete checker reference contains 207 records for each of two original
FPU modes. Its complete 64-KiB external heaps and stacks, unchanged 12-byte
caller windows, ordered private-stack writes and return/fault registers are
consumed. Ordered writes must reconstruct every final stack byte. Each
profile has 2,269 external memory observations and thirteen intentional
faults. No external heap byte is written by the original checker/callees.

The new typed getters add first/next inventory, owner, flags and low class
byte to the already frozen effect/object/slot accesses. They retain the
base pointer separately from the field selector, including nil bases.
All 242 independent invalid read domains and complete native-allocation
guards remain checked. The full PE32 projection compares every external
heap byte; all pointer roles and native callback identities are explicit.

An optional observer wraps the existing owner helper's two actual field
reads. Its traversal algorithm stays unchanged: cached owner before live
class byte, and no terminal class read. All 267 actual default-C parent
invocations are rerun, including four original faults. Their complete
internal read sequences also match the original. The observer is nil in
ordinary production. This instrumented source is distinct from the earlier
initial-memory gate, which covered the uninstrumented helper.

Before any engine, 414 complete native transports replay both original
profiles' external accesses as data. Actual typed reads, ordered values,
widths, fault prefixes and whole final native/PE32 heaps must agree. No game
algorithm selects these replay operations. Original instruction addresses
remain in the immutable reference; the native semantic correspondence does
not assert identical physical machine instructions or caller stack layout.

The compiled engine consumer additionally checks the actual public C root,
every parent/effect call and return, full-width arguments and caller-owned
zero-initialized scratch. The actual dynamic C carrier records entry and
return separately, so faulting callbacks retain their real entry without a
fabricated return. The earlier independent six-address carrier control is
rerun. Parent service and indirect callback invocation remain real C calls.
Every original read, call and return must be consumed without extra events.

Go 1.27.3 normal and strict HD preparation versions one and two each pass
2,464 events per mode, including the earlier three-entry freeze and seven
owner-helper tests, with no failures, skips or build warnings. Version two
separately consumes faulting callback entries in its compiled native branch.
Both versions precede the freeze commit. Exact source snapshots, all six pre-execution mains per
version, expected old-C child faults and complete logs are retained under
`ART/modifier-inversion-inventory-4fa4f0-native-entry-port/`. Duplicate build
trees and executable copies are removed only after archive byte verification.

Evidence: `complete-frozen-consumers-normal-and-strict-qualified-v2.json`
and `full-checker-consumer-machine-code-before-engines-v2.json` in that folder.

All original producers/independent inputs/goldens and the previous freezes
remain exact. Raw engine, current product/E2E and new CI credit are pending.
