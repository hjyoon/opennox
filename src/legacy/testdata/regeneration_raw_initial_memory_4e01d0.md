# Raw Regeneration independent native initial memory

This preparation follows original reference `ccceee5ba`. The active raw
`nox_xxx_effectRegeneration_4E01D0(int, int)` body and prototype stay unchanged;
the raw native engine is absent. No implementation or product credit is added.

The generator reads only the 198 independent inputs, without reading original
outputs or a native engine. Every complete 64-KiB initial PE heap matches the
separate original reference. Typed object, modifier and health fields expand
pointer widths while retaining DWORD/WORD widths. Native guard bytes and
protected pages before and after the C-owned heap cover 36 selected actual
nil/protected read-fault domains.

Eleven actual C/Go layout facts cover holder, class, health pointer, damage
clock, flags, rate, current/max HP, shared object prefix, effect size and
pointer width. Two Go-only EXT handles follow the C prefix. The first new
test incorrectly compared complete Go object size against C size; its failure
and both actual mains remain archived. The corrected test checks the common
prefix and both extensions. Production layouts and prior frozen inputs stay
unchanged.

The actual old-C child passes layout, typed C/Go fields and whole-memory
positive controls above 4 GiB, then the unchanged two-integer raw entry causes
a real SIGSEGV. The separately registered three-argument Regeneration callback
and its two existing integration tests remain unchanged and pass.

Goenv Go 1.27.3 normal and strict HD (`cgocheck2`, `checkptr=2`) each pass
205 events, including all 198 initial heaps and 36 read-fault domains, with
zero skips, failures or warnings. Both drivers and four PREEXEC mains exit
unforced; archives are verified before WORK/output/report duplicates are
removed. Evidence directory:
`artifacts/opennox-port-20261106/regeneration-4e01d0-native-entry-port`.
Receipt: `regeneration-raw-initial-normal-and-strict-qualified-v1.json`.

| Canonical local-suite archive | SHA-256 |
| --- | --- |
| normal v5 | `de561d047b1ffb08be7c1db8fd5e3e9f725c291d7dd4e5e550ba4ffaa212f30b` |
| strict HD v5 | `085408942dc1562fab4423c73280bead555e03f25040273cc2d5725146e0cc63` |

Encoded/decoded initial fixture SHA-256:

```
0eda2dacd7e98102d8aa922088c5c0944d9d38cf00160304e788ee57c4066427
0a06fbc35a302a823f96c3c188efa2f200a8cd33750953ce4370de5dda7cc652
```

Complete compiled consumers must freeze before raw implementation. The
dedicated declaration header is intentionally mutable during that later ABI
migration; the initial generator, fixture, controls and tests remain frozen.
All 38 earlier modifier inputs remain exact. Overall100 is false.
