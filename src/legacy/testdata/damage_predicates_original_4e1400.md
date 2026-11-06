# Raw damage predicates: original reference and pending native preparation

This reference covers the complete original `sub_4E1400` and `sub_4E1470`
leaves. It does not implement their native public C entries, change the
ordinary Go helpers or port their raw DefaultDamage parent.

The unchanged original GAME.EXE SHA256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The executed roots are `004E1400..004E146F` (112 bytes, SHA256
`0ac2ef45c5c13d020b99663c9fcac15321cdf3b825cabb3b44c746aae5a4fb73`)
and `004E1470..004E1491` (34 bytes, SHA256
`fa632520ae8b970b679dd6c5b8bd947f81f4b41d65295f6d3a26edbcb975b92e`).
All 146 bytes and 49 decoded instructions execute in each retained original
FPU run. Both leaves have zero callees or substituted external services.
Padding after the second leaf is excluded from this execution scope.

The independent chosen input set has 487 cases: 317 attack and 170 friendly.
The input producer computes no predicate return, branch or root-fault outcome.
Its compressed SHA256 is
`1f96b7f8b86e8fd9f37cd9ee8c71993b8f1c9a2093ca0e5ba9afb0586f6fdac6`;
decoded SHA256 is
`a6f5a269759c38d3e850436c5acea2b61b629a17e18fd2eb6d4c27026971be55`.
The producer and independent input were sealed before original execution.

Retained PC53 tozero `0x0e3f` and nearest `0x023f` executions produce identical
complete outputs. Each records 865 ordered heap reads, nine declared after-read
stimuli, 1050 stack reads and 45 intentional memory faults (33 attack and 12
friendly). Successful canonical integer returns are attack 128 false/156 true
and friendly 140 false/18 true. Original body instructions write no memory.
Complete 64KiB initial/final heaps and stacks, read widths/values, stimulus
writes, caller words, actual argument lists, returns, all integer registers,
fault EIP/address/width, full original image hashes and FPU certificates remain
in the canonical output and two independent proof records. Final heaps
reconstruct from declared stimuli; stacks, caller words and original image
remain unchanged. These statements describe retained original executions;
no original emulator was rerun during reference installation planning.

Canonical output compressed SHA256:
`1148775df123f822dcc2d5b8b2d97b311edba63eeaa3fbfecb7bb07d548247b7`.
Decoded SHA256:
`81f7a5a71486d1ee3904ca9087f72b8aa4bc40d48bfeb670e592e881c97f74f3`.
The canonical gzip is retained once. Installation copies the canonical
synthetic data byte for byte; it does not create a second large planning copy.
The original executable, maps/audio/video/assets and Save files stay outside
source and are not duplicated by this reference.

The existing main `game-exe-functions.json` already seals these exact two
ranges. Preserve its entire 3373-code/713-data file and all earlier seals byte
for byte. The separate `game-exe-damage-predicates-4e1400.json` repeats just the
two existing canonical records, with the same code-v2 full-tree/executable
binding and strict nonoverlap. It introduces zero new original PE ranges.
Future actual CLI verification of main and dedicated manifests must be recorded
separately; a static PE hash check is not a noxoracle CLI execution.

The raw declarations remain `int sub_4E1400(int source, uint32_t* weapon)` and
`int sub_4E1470(int weapon)`. Attack reads the weapon class first when weapon
is non-null. Its WAND path conditionally reads weapon subclass, then the PE32
UseData pointer and one flags byte. Its WEAPON path conditionally reads
subclass. The fallback uses the already cached class. With no weapon it reads
source class and conditionally a source subclass byte. A null source and null
weapon fault in the raw entry; the ordinary alias's nil-guarded methods differ.
Friendly first checks weapon nil, then class and conditional subclass DWORD.
The nine stimuli retain cached class, subclass-test flags, UseData pointer and
byte-test behavior. Neither root should acquire eager reads or extra nil guards.

Exact original external widths are source class DWORD+8, source subclass
BYTE+12, weapon class DWORD+8, weapon subclass DWORD+12, UseData pointer
DWORD+736 and Wand flags BYTE+96. A DWORD access in place of either BYTE read
can introduce a protected-boundary fault even if its mask gives the same bit.

The five separate native-initial draft files are still pending compiled
verification and old-C positive controls. They reconstruct all 487 initial
heaps from independent inputs only and expose 2922 individual field probes,
including 1222 unreadable field domains. Those 1222 domains are not the 45
original executed root faults. All initial fixture bytes and existing draft
sources remain unchanged by this reference plan.

Partial and unaligned domains are intentional. On the current 64-bit native
layout, class is offset12, subclass16 and UseData848, derived from current
named C offsetof/static assertions and named Go unsafe.Offsetof facts. An
original last-source-subclass-byte case is relocated to leave exactly17 native
readable bytes; its base is misaligned by7 modulo8. The subclass byte remains
readable while following bytes stay protected. Eight chosen inputs use this
shape. Twelve weapon inputs are unaligned by1 and four by3; their native
UseData storage retains that unalignment. Boundary UseData storage expands
from4 PE bytes to the actual pointer width with exact reverse correspondence.
No whole-object padding or forced alignment may erase these wanted domains.

Static v2 initial review found an additional strict-control issue: computing
`unsafe.Add(nil, offset)` can trigger nonrecoverable checkptr arithmetic before
an intended nil-field memory fault. The separately retained v2 draft performs
an actual width-appropriate nil scalar read before nonnil offset arithmetic.
It never returns a synthetic zero or inserts an admission nil guard. All v1
files and original/initial gzip bytes remain exact. Native compiler nilchecks
can report address0; this control preserves original PE role/field/read-domain
projection and does not claim identical native physical fault addresses/PCs.

The initial Go controls intentionally read pointer-free BYTE/DWORD/native-sized
scalar storage at actual named offsets. They never cast the partial base to a
whole pointer-bearing `server.Object`; strict checkptr can terminate on that
misaligned conversion before the intended field access. The C getters use
memcpy only for already independently readable fields and are never called
on a wanted protected field. This establishes an initial correspondence design,
not a production raw getter, runtime fault transport or Go-heap/GC result.

A future full consumer/runtime must keep potentially faulting memory reads in
Go, with before/after observation and exact widths. C memcpy of a protected
field can kill the process outside Go recovery. Properly aligned real Go-owned
pointer fields must load as actual unsafe.Pointer values so GC tracing survives.
For unaligned C-owned pointer storage, one proposed transport is a full-width
pointer-free scalar Go load, followed by a pure C uintptr_t-to-void* conversion
which never dereferences. Keep the base live until conversion; legal C graphs
retaining Go pointees additionally require those pointees to remain explicitly
pinned/live. A KeepAlive on a C buffer alone does not root a Go pointee hidden
inside integer bytes. Never cast Go-heap uintptr bits back to unsafe.Pointer in
Go or use the scalar path to evade cgo checks. The aligned Go graph, unaligned
C graph, protected reads, cross-transport fault handling and real pinned Go
GC cases must all be validated before runtime fidelity is claimed.

The active raw C call-site audit finds two attack calls and one friendly call,
all inside `int nox_xxx_damageDefaultProc_4E0B30(int,int,int,int,int)`.
Its source `a2` and weapon `a3` are already integer arguments. Widening only the
two child signatures or zero-extending their low DWORD cannot restore native
high pointer bits lost before the parent boundary. That raw parent remains a
separate unresolved port. Static references do not establish actual gameplay
admission, so neither dead-code nor product-fix credit is assigned here.

The ordinary Go `defaultDamageAttackQualifies4E1400` and
`defaultDamageFriendlyException4E1470` remain unchanged. Their nil/class
methods and eager subclass policy are separate admission contracts; they must
not be exported as raw replacements without preserving the raw read/fault
contract. Their legitimate native gameplay paths receive no regression or
completion assertion from this reference.

Installation must wait for every current local Go/full-five/product driver to
terminate. Root reviews and commits eight additive reference files first,
keeping the current76 source inputs and entire main manifest exact. Independent
native initial controls must then execute and freeze; complete whole consumers
must freeze before either engine is installed. Historical proof source hashes
retain their original values: whole GAME3_2.c/header hashes have since changed
for the Readiness port, while these two raw predicates and their raw parent
remain byte-identical to the historical initial-draft commit.

This reference does not close a raw predicate, parent, product, campaign,
cross-platform or overall100% gate.
