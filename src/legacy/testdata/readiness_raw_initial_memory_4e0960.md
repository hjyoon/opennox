# Raw Readiness independent native initial memory

This preparation follows original reference freeze `e92c637c0`. The raw C
entry still has its original integer ABI and body. No native algorithm or
complete original-output consumer is installed at this checkpoint.

The generator reads only the independently chosen 113-input gzip and computes
initial bytes, selected field domains and pointer roles. It does not read the
original output, invoke the emulator or compute return/branch/fault outcomes.
All 113 complete PE initial heaps agree with the separately retained original
reference. The initial gzip remains byte-identical across both native runs.

Actual C offsetof/sizeof/alignment getters and Go unsafe layout checks agree
on 15 facts. Native item InitData/class, four modifier slots and callback/
signed DWORD strength fields use their actual layouts. The C object is912
bytes; the Go object includes two additional native handle words and is928.
Tests compare the shared C prefix and separately check those Go extensions.
Partial readable domains and predecessor/successor guard pages are retained;
no native domain padding is introduced. Unobserved expanded native fields
are not asserted to alias their PE32 byte positions.

Normal and strict HD (`cgocheck2`, `checkptr=2`) Go1.27.3 each pass119events,
zero skips/failures/warnings:113whole initial heaps,1808typed C/Go probes,
52unreadable field domains and the15layout facts. A real unchanged-C child
first verifies healthy complete native data and callback addresses above4GiB,
then actually faults in the old integer entry. Its parent and both preserved
package mains exit unforced0. Four actual mains are archived before execution.

The original comparison identity0x004e0950 maps to actual C.nullsub_22. The
independently chosen NONMATCH0x004e0951 is relocated, on64-bit hosts, to a
pure C opaque token sharing the callback's low32bits but differing in its
high32bits. Neither code path dereferences or invokes that token. Typed C
uintptr and Go pointer loads retain its full value and reverse-project it
to the unchanged PE token. A deliberately wrong both-sides uint32 comparison
must compare equal while full pointers differ. Three existing input rows
prove this per profile, without changing original inputs/outputs or domains:
`callback/2/004e0951`, `callback/3/004e0951`,
`priority/004e0951/004e0950`.

The first actual v1 initial runs passed before this additional distinguishing
native identity case was introduced. Their exact sources, seals, mains and
logs remain. The accepted v2 adds only the native token relocation/control
and assertions in two initial Go files; header/generator/initial gzip and all
prior56/reference8 inputs remain exact. Evidence:
`ART/readiness-4e0960-native-entry-port/readiness-raw-initial-normal-strict-qualified-v2.json`.

Keep these initial model/getter/generator/fixture bytes frozen before complete
consumer acceptance and native implementation. `readiness_raw_4e0960.h` is
the separately mutable ABI declaration; a later native implementation must
update its declaration consistently and preserve the exact old C body.
This initial-only checkpoint supplies no raw native engine, product, earned
campaign, other-platform or overall100credit.
