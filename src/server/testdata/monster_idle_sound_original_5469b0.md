# Original idle-sound entry

Original executable SHA256:
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
No original asset or executable is included.

Both startup FPU controls, `0xe3f` and `0x23f`, produce byte-identical
5,670 contexts: 5,594 idle-root contexts, including four raw faults, and
76 explicitly ancillary sound-set-getter contexts, including 15 raw faults.
All 353 instruction bytes execute: root `5469B0`190, aggression predicate
`534390`38, action getter `50A020`27, sound-set getter `424300`31 and RNG
`415FA0`67. Only audio `501960` is a declared world service. No predicate,
getter, RNG arithmetic or frame/FPS load is replaced by a service result.

The idle root has a void ABI; incidental EAX values receive no return-value
credit. Its distance comparison admits unordered values. A signed stack
index of -1 reads original MUD+528, the timer field; the native projection
preserves that alias explicitly. All 24 stack slots and -1 are compared;
other malformed indices receive no equivalence credit. The ordinary shared
action getter's established negative-index guard is unchanged.

The root retains its entry MUD, computes 20/60 times FPS with DWORD wrap and
signed RNG bounds, reloads frame after RNG, then writes the cached timer.
Actual `424300` reloads class and current MUD. Sound-set timer aliases must
therefore read the newly written ID. Audio-callback mutations remain visible.
Inputs include raw NaN/infinity bits, class/enemy/action admission, timer and
FPS wrap, seeds 0/1/4094/4095, independent/embedded sound sets and mutations.

Inputs, producer, results and native consumers are frozen before engine edits.
Gzip SHA256:
`ae7d8a867fa0981d17f3c76c920415e45f96989fe07e49f460913eb59b35d31d`.
Decoded SHA256:
`5315122cb88a7e9c2cf228139fc2dbb2bb9d4ba06b28a03ce12f5d5d0a6fa1d2`.
Original code, RNG tables, heap writes, registers, stack, FPU control/TOP and
canaries are checked. Native comparisons cover actual public C and direct Go
entries, whole C-owned records/padding/neighbors, native pointer projections,
high argument-word sentinels, actual prand, ordered services and fault prefixes.

Before porting, the actual public C entry SIGSEGVs only after independent
C/Go layouts, full initial-record guards, sound-getter C/Go and action-getter
C controls pass with valid native pointers above 4 GiB. Its exact executable
and raw failure output are retained under
`ROOT/monster-idle-sound-native-entry-port` along with original outputs and
seals. The reference-only commit is `cf2e182`; preceding engine is `2db2967`.

Separate composition tests connect real clocks, prand, Audio.EventObj,
registered descriptors, delayed replay, allocation and observers through the
raw core and ordinary server entry. The public C composition also obtains
clocks and prand from its production bindings, retaining only the explicitly
declared audio observation service. These checks do not certify physical
speaker output, all campaign paths, networking or overall port completion.
