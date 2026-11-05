# Original raw Inversion inventory checker before native implementation

This additive reference follows the three-entry original, independent native
initial-memory and full-consumer freezes at `26086de87`, `da839f839` and
`bbee00976`. None of those producers, inputs, goldens, getters or consumers
changes. The raw `4FA4F0` C body still accepts two integers. No new native
engine is implemented by this reference.

The actual original checker calls `4EC580` at `4FA54C`, then reloads the
modifier callback for the indirect call at `4FA561`. Its existing ordinary
Go collision helper restores the registered-effect path; it does not close
this raw public C ABI. Merely casting the old callback comparison to compile
new predicate signatures would leave this coupled entry unsafe.

All three complete original ranges were already sealed. Their unchanged
GAME.EXE input is SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.

| Actual original body | Bytes | SHA-256 |
| --- | ---: | --- |
| Checker `004FA4F0` | 159 | `8bf018e2ff189e5e4f3725004765a960730e7b8d0748a27f6c5a148c0e6967c2` |
| Parent lookup `004EC580` | 38 | `7e7580cf0eb5c57ba46812b911b2fd4fa3224fc26ee646d777e207b350e91859` |
| Inversion `004E03D0` | 22 | `fe6639dcaebfa8919164437706cf56e6298ca32a42c741a2fc1e770373922d6f` |

`generate_independent_inversion_inventory_4fa4f0.py` chooses 207 inputs before
the native checker. They cover nil/unmapped roots, empty/bad inventory,
equipped flags and exact class mask `0x13001000`, both eligible modifier slots,
invalid slot-zero/one decoys, eleven signed strength boundaries, callback
identity decoys, shared modifiers/data, multi-item order, early success before
bad slots/next links, and live owner chains. Bad targets with both positive
and weak matching effects prove that parent lookup precedes the strength
test. A Player with a bad owner proves the cached owner is not followed;
terminal objects and intermediate Players exercise the original short paths.
The encoded/decoded input SHA-256 values are:

```
69c228a2eb2ef67ce89f8d2127766d0b209cb29d80433a3d43fda9ec6e515e1a
79196fbdb001e865d8a41b541da2adee663da885794d0c208b07a398cf63ac5c
```

`original_inversion_inventory_4fa4f0.py` executes all 219 original bytes in
both PC53 FPU control modes, `0x0e3f` and `0x023f`. Parent lookup and Inversion
execute their actual original bodies; there are no declared external
services. Each profile performs 267 original parent calls and 263 original
Inversion calls, with 2,269 observed external memory accesses and thirteen
intentional memory faults at six original instructions. All complete
64-KiB external heaps stay unchanged. Full 64-KiB stack before/after bytes
and all ordered writes reconstruct exactly; caller arguments stay unchanged.
Successful returns preserve nonvolatile registers and stack balance. Full
declared PE section images and FPU control/status remain unchanged. Volatile
entry registers are explicitly initialized independently for each input.

The original parent helper first reads owner DWORD **508**, then reads class
**byte 8** only when that owner is nonzero. A terminal object returns before
its class is read. A Player with an owner returns itself; other objects follow
the already cached link. This is `ObjOwner` (native offset 552 on 64-bit),
not original field 130/offset 520. Native layout and fault domains still need
their own independent controls.

Complete heap/stack/caller/register/access/write/call/return/fault outputs are
identical between FPU profiles. One canonical compressed output is retained,
with separate execution proofs. Its encoded/decoded SHA-256 values are:

```
0bb44fa16475b2311e8d1ec5aa862fbeb87b7b94141719efe681e44310269853
df0ec5b5b7bb81bc8c36e7e7ff2f7e936f3d741949926d91c2c4f86e6dcea609
```

Both committed producers reproduce their full inputs/outputs byte-exactly.
Evidence: `ART/modifier-inversion-inventory-4fa4f0-native-entry-port/repository-original-reference-before-native-qualified-v1.json`.
The earlier 3,373 code/713 data seals stay unchanged. Source/whole-memory
observations add no native initial-memory, unchanged-native-C failure,
implementation, stock-product or earned-campaign credit. Independent native
initial/fault/default-parent controls and complete native consumers must
freeze before the related four raw C engines change.
