# Original NPC animation pointer 533D00

Both sealed FPU profiles execute 19,668 unique contexts and 20,004 parent
calls each: 4,816 NPC and 14,852 non-NPC contexts, with identical results.
All 361 parent body bytes and 550 existing helper bytes execute (911 total).
Seven following root NOP bytes are excluded. Actual selector inline-table
reads cover 83 bytes; dynamic weapon/frame table inputs read 100 distinct
bytes. This does not claim full table-index or feature coverage.

The unchanged weapon helper 4FA280, frame helper 4F9F90, selector 533790,
Mimic/Plant/Zombie predicates and WORD type helper execute. Only dictionary
lookup 4E3830 is declared. It returns valid type records/nil and applies
specified input mutations. No helper, predicate, scratch or final pointer
result is supplied. The reference observes helper calls, repeated returns,
scratch writes and bytes, cached entry update, the post-selector table
reload, cache publication and complete heap/stack/global/code/FPU guards.

The primary native C and Go comparator uses actual C weapon/frame/selector
calls and actual native predicates with registered type records. Its frame
provider reads the pinned input table. Native scratch bytes and surrounding
memory are compared; an original scratch-store trace is retained without
claiming a native store-count measurement. Fourteen separate ordinary modes
use the unmodified factories, actual PlayerDraw parsing and runtime weapon
table; they do not replace frame providers or predicates.

The independent original producer and its static dependency live together
in `monster_NPC_animation_533d00/`. This keeps the earlier selector's static
dependency byte-identical. The producer requires Unicorn 2.1.4 and accepts
`--game`, `--fixture`, `--proof` and `--control` (`0xe3f` or `0x23f`). Use
fresh output paths; it never overwrites a sealed fixture or the original.
GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.

Before-port C pointer crashes and the older private Go NPC projection are
preserved under the verification ROOT. That projection selects 4,816 NPC
contexts / 4,928 calls: 4,396 children pass and 420 differ in scratch bytes.
It does not certify a pre-existing public native Go parent wrapper. A log
literal 19,668 denotes all reference rows rather than that selected subset.
Preparation syntax, argument-stack observation and comparator API-signature
failures remain separate preserved records. Overall porting 100% and the
final nine chapter 1–3 audits remain open; chapter 4+ gameplay is excluded.

Four lookup-mutation contexts clear the cached table and return raw address
0x80. Initial strict execution correctly rejects Go Add(nil,128). The port
isolates only that original numeric nil-base address conversion in a
noinline go:nocheckptr leaf; ordinary native pointer arithmetic remains
instrumented. All original cases and expected addresses remain unchanged,
and the first strict failure and pre-fix source stay preserved.
