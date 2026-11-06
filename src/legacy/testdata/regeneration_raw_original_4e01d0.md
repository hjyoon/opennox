# Raw Regeneration original reference before native implementation

The old public `nox_xxx_effectRegeneration_4E01D0(int, int)` C body remains
active. Its separately registered three-argument native variant does not
close that raw C ABI. This reference precedes raw implementation changes;
it supplies no native initial-memory, old-C high-pointer fault, native
consumer, implementation or product credit.

The original GAME.EXE is unchanged, SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
These three complete bodies were already sealed; no existing code/data
manifest entry changes.

| Executed original body | Bytes | SHA-256 |
| --- | ---: | --- |
| Raw Regeneration `004E01D0` | 204 | `7b5720413e3ae9a93e233ed1ec5d6e07fd973d1ae6f79bc1a9fd8055d4d2daea` |
| Current HP `004EE780` | 26 | `c842830466581834b537f438907e121a61547ce43c6d07ef033207e13970b674` |
| Maximum HP `004EE7A0` | 27 | `f45238ed5aa92034eef46ed574b2967bc0595a25e2a1d3b149fe6bc20f6475d6` |

`generate_independent_regeneration_raw_4e01d0.py` chooses198inputs without
calculating expected native results. It covers null/unmapped and last-field
pointers, self-held items, early gates protecting invalid later fields,
unsigned cooldown/frame wrap, exact dead flags, WORD health comparisons,
unsigned duration/armor division and wrapped DWORD products. Declared field
and clock perturbations distinguish cached holder/class/rate/HP/clock values
from live reloads and expose both original integer-DIV fault instructions.
Input encoded/decoded SHA-256 values are:

```
e7a9dc92b938fd023aef24b062e14b0e7127259350ec67cd2bdf8b14c030b713
ceea31a4b90206d5f2979049f1d28f0aaa1ba7250f6ecf8e825360bbf462fa66
```

`original_regeneration_raw_4e01d0.py` executes all257original bytes and85
decoded instructions in each PC53 FPU mode, `0x0e3f` and `0x023f`. Both HP
getters execute their real original bodies. Frame and FPS are read from
the original global addresses `0084EA04` and `0085B3FC`. A fresh emulator
per input isolates architectural state after integer division traps.

There are exactly two explicit external service boundaries: armor-flag
lookup returns the independently supplied flag word; HP adjustment records
the actual cached holder and delta1. Neither external body nor final HP
effect is executed or credited. Complete native consumers must retain this
declared comparison scope and independently verify actual default services
and product HP effects. The root ABI is void: hardware EAX is retained as
an original execution certificate, without treating it as a defined return.

Each profile records341actual maximum-HP and175actual current-HP calls,
61external armor queries and97external adjustment calls. There are2,822
ordered reads,22declared input-stimulus writes,9memory faults and25integer
divide faults. Complete64KiB heap and stack before/after bytes and ordered
writes reconstruct exactly; caller words stay unchanged. Successful returns
preserve nonvolatile registers and stack balance. Full PE image bytes stay
unchanged apart from declared clock stimuli; original file and FPU control/
status remain unchanged. CPU registers and fault prefixes are retained.

All complete output bytes match across both FPU runs. One canonical gzip
is retained, with separate execution proofs. Encoded/decoded output hashes:

```
19aa7f1b50b6958762d06a912e7c82e2adfe055b6b79fc237845c6743fb2554e
e3bf75e4030f35a57e20155abc2d6f1d18780f0d2a0ba000978d6040d0db6268
```

Evidence: `ART/regeneration-4e01d0-native-entry-port/repository-original-reference-before-native-qualified-v1.json`.
Initial sandbox SIGILL and the earlier shared-emulator interrupt-observer
failure are preserved. All3,373code/713data seals and38prior modifier
reference/initial/consumer/helper inputs remain exact. Independent native
initial/read-fault controls and complete raw-C consumers must freeze before
the raw engine changes. Overall100 remains false.
