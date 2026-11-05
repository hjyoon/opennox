# Original Inversion/Grip predicates and Grip traversal

This is a reference freeze before a native implementation. The repository's
three raw C entry bodies and integer argument declarations remain unchanged.
An ordinary Go modifier/collision dispatcher does not close those public C
ABIs. Native initial memory, full output consumers and actual high-pointer C
failure controls must be frozen and verified before changing these bodies.

The immutable GAME.EXE input has SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The existing Inversion seal and two added seals cover these complete bodies:

| Entry | Bytes | SHA-256 |
| --- | ---: | --- |
| Inversion `004E03D0` | 22 | `fe6639dcaebfa8919164437706cf56e6298ca32a42c741a2fc1e770373922d6f` |
| Grip traversal `004E03F0` | 139 | `919c3f4fb8031a9ebf9aebea590e0d6e0f6e5c4db0070ed4b09e5621c30cefa5` |
| Grip `004E0480` | 22 | `0ea4bb7047e57a59eaa5ee0ebf55591a37f990c49ab8cde7aae8bafdc3924a9f` |

`generate_independent_modifier_defend_predicates_4e03d0.py` chooses 298 inputs
without using a native implementation. They include eleven scalar boundary
values, the sign bit, negative values, separate and aliased output words,
unaligned output, the last readable strength word, ignored argument words,
nil/unmapped input and output, class gates, both modifier slots, callback
identity decoys, shared modifier pointers, and an unreadable later slot that
must be skipped after an early successful Grip result. Slots zero and one
contain invalid decoys. The inputs' encoded/decoded SHA-256 values are:

```
adc9b7435ce18b9006cf77e7e84057d18b1ed42272bda5432dd81d99b99ad5c7
25552006fbfe8479cee144bf9cc82b0080b4ed5f5578d823efb3e89183c9d1c5
```

`original_modifier_defend_predicates_4e03d0.py` executes all 183 original
instruction bytes (73 decoded instructions) in both PC53 control modes,
`0x0e3f` and `0x023f`. The nested indirect calls execute the actual original
Grip body; no external helper is substituted. Each profile records all 298
complete initial/final 64-KiB heaps, ordered writes, caller argument windows,
return values and faults. There are 18 intentional memory faults in each
profile. Successful returns preserve nonvolatile registers and the expected
stack balance; both FPU controls/statuses and the complete PE image remain
unchanged.

The complete outputs are byte-identical between profiles, so one canonical
compressed output is retained with two separate execution proofs. Its
encoded/decoded SHA-256 values are:

```
f3cc6af63c5eb0a6b0ced5cf5b9af406cd72b14047562d67222a514020ce9f79
118b828db9ae95e022339f711ec1a218e5574ab0f4ad3e0c79f93aee09c90fe7
```

The original predicates use signed `setge`/`setl` after comparing the DWORD
strength with one. For `0xffffffff`, original Inversion writes/returns zero
and original Grip writes/returns one. The current decompiled C bodies instead
read an unsigned `uint32_t`, which gives the opposite results for negative
strengths. They also accept the effect pointer through `int` and read the
Win32 offset 96. Both differences remain pending implementation work.

The original traversal admits item class mask `0x03001000`, caches the
modifier-array base and visits slots two and three. It calls only the exact
Grip callback and returns one when that callback writes zero. The original
machine body reuses caller argument three's stack word as its loop counter
and argument four's word as callback output. Both observed words are retained;
the argument window is not falsely declared immutable. This describes the
original stack, without claiming identical physical stack layout in another
ABI or arbitrary aliases between unrelated native structures.

An initial, unfrozen producer incorrectly asserted that caller argument three
stayed unchanged. Its source, inputs and complete logs are retained in
`ART/modifier-defend-predicates-native-entry-port/first-original-producer-caller-window-failure-v1.zip`;
the failing exit is in `original-tozero-v1-exit.json` in the same directory.
The successful producer records both original scratch words. No native engine
or repository C body was changed to resolve that producer failure.

Reproduction and independent C layout evidence live under
`ART/modifier-defend-predicates-native-entry-port`. The C-only layout probe
observes native effect size 208, callback/strength offsets 128/140, object
class/init-data offsets 12/760 and pointer-sized modifier slots. This probe
does not establish Go layout correspondence, native initial-memory fixtures,
actual unchanged production-entry failures or native output equivalence.
