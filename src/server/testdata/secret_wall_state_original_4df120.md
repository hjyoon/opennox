# Original secret-wall state packets 4DF120 / 4DF180

Reference preparation precedes the native engine change. Original GAME.EXE
SHA-256 is `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Both complete x86 functions execute: 86 bytes each, 172 instruction bytes.
Twenty alignment bytes are excluded. The executor substitutes no instruction
or function result. Player-unit First 4DA7C0, Next 4DA7F0 and packet Send
4E5420 are the declared world-service boundaries; none supplies the packet
encoding, pointer-chain loads, traversal control or final return result.

Both PC53 ToZero `0x0e3f` and nearest `0x023f` execute 1,188 contexts,
including 26 fault prefixes. The complete compressed and decoded outputs
are byte-identical across profiles. Inputs cover both opcodes, wall WORD
boundaries, all 256 unsigned recipient bytes, empty/reordered/repeated-unit
traversal, shared update/player records, live field changes after First,
Send and Next, packet-buffer mutation, ignored Send results and faults before
and after an earlier successful send. A null result from First terminates
normally, even in the diagnostic case named `fault-first-unit/-1`.

The wall WORD is cached before First. Each packet's three bytes and complete
routing arguments are captured before the Send service mutates state.
Next receives the current unit after Send; future recipients are loaded
through the live unit/update/player chain. Both valid paths return zero
from the terminating iterator, independently of Send's result. Guards check
the whole original image, complete 64 KiB heap including service mutations,
protected caller stack, valid-return nonvolatile registers, and FPU CW/TOP.

The native consumer compares the actual public C entry and the raw Go
contract against every original event prefix, return and fault boundary.
Independent existing C record layouts check the Go wall, object prefix,
player-update prefix and player-index offsets. Complete native 64 KiB
allocations and canaries must match only the declared service edits; object
class and unrelated poisoned fields must not add raw-entry admission.
The existing ordinary snapshot/admission implementation has separate tests
and is not replaced by this raw traversal contract.

Before porting, both actual old C entries SIGSEGV above 4 GiB after existing
typed C and ordinary Go player traversal, full native wall/object/update/
player pointer-chain controls and seven whole-record guards pass. The exact
qualified before executable is SHA-256
`fea25e0e210c10b74cf1062051f4d4399f529fa1a6b222a7a908dbe9f2dc732c`.
The first preparation attempt incorrectly used nonzero `alloc.New` arguments
as initializers. Its ordinary control rejected before target entry; explicit
post-allocation field stores correct only fixture preparation. That first
source, executable and failed output remain preserved alongside the passing
before reproduction. No original expected value changes.

Canonical gzip SHA-256:
`c20df57cec10adf344255c8bba6316a6c6292520ddf655fab73d866150cb8f50`.
Decoded SHA-256:
`f5051b573ac657881e21adaae6090f2f9257095cc5653bb2ca63686c025fb25b`.
Full original proofs, exact C bodies, preparation history, before executables
and logs are retained under `ROOT/secret-wall-state-native-entry-port`.
