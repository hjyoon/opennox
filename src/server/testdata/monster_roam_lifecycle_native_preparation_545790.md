# ROAM native fixture preparation correction

The original five references were frozen before runtime `7d404efc`. Its normal
selection passes 21,126 events. The first highres/cgocheck2/checkptr2 run
completes the twelve actual C/ordinary production modes, then rejects the
native fixture's construction of an intentionally unaligned typed Waypoint
pointer at seed 00000001, before the original comparison invokes the engine.
Its exact binary, full JSONL and all original reference copies remain retained.

The consumer changes only how it prepares those opaque pointer bits: a raw
little-endian uint64 store at the independent C mirror offset replaces the
typed Go pointer conversion. Every seed, case, expected value, original
golden, fault, upper-bit guard, whole-record assertion, selector and strict
option stays unchanged. No checkptr or warning suppression is added.

An isolated checkout of the pre-engine reference-only `c8293c51` compiles the
corrected consumer and a separate ordinary-compiler equivalence test. All
five seed patterns produce byte-identical complete 64 KiB native allocations
with the old and corrected preparation. Both unchanged legacy C entries also
SIGSEGV after their independent native layout and positive controls pass.
That exact before binary and complete outputs are retained under
`ROOT/monster-roam-lifecycle-native-entry-port/pre-engine-reference-v2` and
`pre-engine-corrected-fixture-preparation-v2-qualified.json`.

V2 preparation was corrected after discovering the strict failure. It is
revalidated on pre-engine source, and is not retrospectively described as
chronologically frozen before the original engine commit. The other four
frozen reference files, original execution producer/outputs and all runtime
engine sources remain unchanged. The original provenance's out-of-array
native-projection limitation remains in force.
