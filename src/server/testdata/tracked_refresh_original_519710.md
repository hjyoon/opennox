# Original tracked refresh 00519710

The independent fixture executes all 80 unchanged GAME.EXE bytes at 00519710.
GAME.EXE SHA-256: 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a.
The root SHA-256 is f7168e841a14f221378b33686bc1b33cb4cfe364c212a45c26429991b0eb3bdc.

1,975 independent valid contexts / 1,999 calls execute in each FPU profile
0x0e3f and 0x023f, with byte-identical outputs. The one child, 00417270,
is a declared tracked-count service boundary. No child-body execution or
whole-game equivalence is claimed. The direct DWORD frame global is 0084EA04;
there is no frame function call in this original root.

The fixture SHA-256 is `bb8195ad55dbea82058220f37af0996a19ef1fa3e22ae2f722da4ced5fa810f8`;
decoded JSON SHA-256 is `afce8985033666560a40efb098041d7f5097b10385f319209b28a6678083a560`.
Inputs were frozen before original execution. The original output and these
comparators were frozen before engine edits. Two separate nil faults are
recorded; nil is not silently converted to false by the raw entry.

The signed count uses a zero early return, signed greater-than-60 comparison,
then signed 60/count division with truncation toward zero. Elapsed frame
subtraction and the final comparison are uint32. All 256 input index bytes
are checked as argument forwarding under the declared child boundary, not as
valid indices for the production player registry. Cases cover quotient edges,
negative counts, wraparound, callback changes to last/frame/player/index,
and repeated calls. The root loads last after the count callback and before
the direct frame read. It does not reread Player after the count callback.

Actual C-to-Go and direct Go checks compare the full return word, child
arguments and mutations, frame access count, full native update/player
allocations and neighboring guards, plus original post-call state. A separate
production binding check covers actual minimap nodes; controlled negative
counts do not claim a production minimap can produce negative sizes.
The guarded ordinary convenience path retains nil handling separately.
This entry does not close Windows memory corruption, audible fidelity,
the user sale-silence report, campaign final audits or overall 100% porting.
