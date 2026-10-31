# Native object-update dispatcher 00519410

The fixture executes the unchanged 590-byte GAME.EXE body at 00519410.
GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`;
the body SHA-256 is
`26e7eeddc206c1a029fff0f989e2acc358d0cbe1dfd5518e6337ac74527baeb9`.
All 590 instruction bytes execute on valid inputs. Two original x87 control
profiles, 0x0e3f and 0x023f, produce identical observations for 5,918 contexts
and 5,930 calls each. Four additional nil-pointer contexts record faults
separately. The original executable and every failed native input remain
preserved under ROOT/object-update-dispatch-native-entry-port.

The thirteen child calls are declared service boundaries: friend publication
4D97A0, owner check 4EC4F0, game flags 40A5C0, minimap tracking 4173D0, ray
535250, sprite 518AE0, complex monster 518960, player 518C30, phantom 5187E0,
simple 5188A0, special reports 527E50, out of sight 528A60 and shadows 528A90.
This certificate does not execute or port those child bodies. The separate
ordinary integration uses their production native bindings and actual local
queues; its lower network publication boundary is controlled.

The input producer was frozen before original execution and comparator or
engine implementation. Both original runs and the actual C-before pointer
fault were sealed before implementation. The comparator was then frozen;
its first build exposed a missing `comparable` allocation constraint. Only
that generic constraint changed, with the original comparator and build
failure retained. Inputs, expected returns, event transcripts, assertions
and allocation guards were unchanged.

The compressed fixture SHA-256 is
`ab4bcac74cc81066399619ca7ff7dddf6ac4ae8b3d523ff923c7c8d92828d4a3`;
the decoded JSON SHA-256 is
`103d8a315d8b0fb5e25c66b81798fcedefc1a8f198d4a60a8de14f3020588638`.
Comparisons check the public byte independently from the successful-send
boolean, ordered child arguments and scripted mutations, every observed
object/state word and complete native allocations with adjacent guards.
Pointer writes are translated to their corresponding native fields without
masking pointer values or allowing unrelated native bytes to change.

Cases retain sender precedence, pending/seen/friend masks, original low-byte
reads, cached update and reloaded Player pointers, camera and target bits
copied before callbacks, callback mutations, repeated calls, recipient/target
aliasing and all 256 return bytes. Ten representative valid player indices
include 0 and 31; this is not an exhaustive test of all 32 indices. Invalid
player indices are outside the raw entry's original valid-input contract.

The public C ABI remains `char(nox_object_t*, nox_object_t*)`. Its original
return is not a boolean: this fixture contains 48 successful publications
with return byte zero and 4,483 calls without publication whose return byte
is nonzero. The shared native core therefore returns byte and sent status
separately. The ordinary Go adapter retains its recipient/nil/index guards
and guarded sprite registry policy while sharing the root's original target
destroyed check and mutation order. Four original nil faults remain faults
in both actual C-to-Go and direct Go comparisons; no strict exemption is added.

The previous PE32 C implementation is retained under `#if 0`. It reproduced
SIGSEGV after valid native getter controls: both recipient and target were
above 4 GiB, and the recipient update was at the native offset 872, while
the C body read PE32 offset 748 and narrowed both pointers to `int`.
This bounded entry certificate does not establish overall porting completion,
resolve the separate Windows memory issue or finish the nine campaign audits.
