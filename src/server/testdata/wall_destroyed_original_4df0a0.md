# Original wall-destroyed packet entry 4DF0A0

These references precede the native engine change. Original GAME.EXE SHA-256
is `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
All 118 instruction bytes of the complete function execute, including both
return paths. Ten alignment bytes before 4DF120 are excluded. No original
instruction or target-function result is substituted. First 4DA7C0, Next
4DA7F0 and Send 4E5420 are the declared world-service boundaries.

Both PC53 ToZero `0x0e3f` and nearest `0x023f` execute 1,974 contexts,
including 43 fault prefixes; their complete outputs are byte-identical.
The wall WORD and three-byte opcode-58 packet are cached before services.
Recipient 32 broadcasts through live unit/update/player traversal and
returns zero from its terminating iterator. Every other signed 32-bit
recipient goes directly to Send and preserves its signed 32-bit return.
The direct path must not inspect player units or call either iterator.

The input corpus includes wall WORD boundaries, all unsigned player-index
bytes, signed recipient boundaries, positive/zero/negative Send results,
empty/reordered/repeated/aliased records, live field changes and packet
mutation. Direct paths deliberately contain invalid unit/update/player
records that must remain unread. Original guards cover the complete image,
64 KiB heap including declared service edits, protected caller stack,
valid-return nonvolatile registers and FPU CW/TOP.

The independent native consumer compares actual public C and raw Go calls
against every event prefix, return and fault boundary, with complete native
64 KiB record/canary checks. It reuses the unchanged earlier independent
wall/object/player-update/player layouts and original service contract.
The existing ordinary Go sender has a void return and separate admission;
it does not replace this raw public C contract.

Before porting, the actual C entry SIGSEGV in both direct-recipient-0 and
broadcast-recipient-32 modes with native records above 4 GiB. Existing typed
C and ordinary player traversal, full wall/object/update/player controls and
seven whole native records pass before either target invocation. The full
before executable and both C crash outputs remain in the verification ROOT.
Before executable SHA-256:
`492c074587d5d44f130736fa326456ca7096ecf1bc36c97abf2623bcfaf63107`.

Canonical gzip SHA-256:
`bc6d8d208240d85438212f9b841824de6e1842818a56dc002cff8b18b0f3e05a`.
Decoded SHA-256:
`2e109604ed6e162a5ce07617b404245a5d46a8568ea9fc46bfb3e4915c6fd5e3`.
Full proofs, original C body and pre-engine records remain under
`ROOT/wall-destroyed-native-entry-port`.
