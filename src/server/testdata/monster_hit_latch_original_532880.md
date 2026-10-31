# Original successful-hit timestamp reference

`original_monster_hit_latch_532880.py` executes unchanged GAME.EXE bytes,
SHA-256 `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
It loads the complete 48-byte interval at `00532880`, executes all 33 body
bytes, and does not credit the 15 trailing NOP bytes as executed. The
existing oracle already seals the complete body; this port adds no range.

Both startup precision-53 / ToZero (`0xe3f`) and diagnostic nearest (`0x23f`)
profiles execute 3,072 inputs with no service or numeric-result substitution.
They produce identical integer outputs. The producer independently observes
the native PE32 return role, timestamp DWORD, conditional live Frame read,
exact original writes, both update records and full heap/stack/global/code
guards, nonvolatile registers and unchanged FPU state.

The input matrix varies boundary timestamp/frame words, class/flags and two
valid linked update records. Class and flags do not gate this root. A nonzero
timestamp skips the clock; writing frame zero leaves the timestamp empty for
the next call. The frozen runtime comparators use startup-profile outputs,
fixed hashes/dimensions and real C-owned native allocations. Both public C
and Go calls must return the cached update pointer, preserve every other
native byte and perform exactly the original conditional clock reads.

Additional binding checks use the ordinary server clock through the actual
public C symbol, including frame-zero reentry and a nonzero class-zero record
that must avoid server lookup. No replacement clock is installed for those
production-binding checks. Existing DefaultDamage and player-damage cases
remain unchanged; their shared implementation preserves each parent's current
clock policy. Their complete transitive behavior is a separate scope.

The one retained extracted C caller is the old int-ABI DefaultDamage4E0B30
parent. Its PE32 word is explicitly widened for the typed root declaration.
That adapter does not make the parent or its upstream C callers native64-safe;
their migration remains open. Ordinary registered Go DefaultDamage already
uses native object pointers and now shares this root.

Original allocation and ordinary native Go action controls passed before
the actual C entry reproduced a pointer-truncation SIGSEGV above 4 GiB.
The original executable, all before/failure evidence and all 4,099 preceding
test/fixture/E2E inputs remain preserved. Nil/invalid-pointer fault semantics,
arbitrary caller chains, full campaign behavior and overall 100% completion
are not established by this bounded reference.
