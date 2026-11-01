# Original timer percentage scaler 486640

GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The complete 35-byte body has SHA-256
`16ee2a31d94d61e1e2c501d96c492c8485b231d5686bfcc278e0c8fc1d76e3c6`.
The following thirteen NOPs are separate alignment, excluded from execution.

Both original FPU profiles execute 576 independent contexts and one separate
nil-group fault. Every body byte executes without child, instruction or
arithmetic substitutions, and both output byte streams are identical.
The input contains signed rates, percentage-word and fractional-word edges,
wrap boundaries, standard rates and 256 deterministic interior pairs.

| Artifact | SHA-256 |
| --- | --- |
| Original output gzip | `2b19bbbe0fef266d762ca90a7280d5e0f60e7bc599d6673da4d69c045ddf5ddf` |
| Decoded original output | `251a9e349276c71661bc4ae49bc41f5543a17198913740f9f8903c93c03ca513` |

The entry reads the DWORD at group+36, logically shifts by16, multiplies
within a wrapping 32-bit register and performs signed division by100 toward
zero. It does not clamp or retain an unwrapped wider product. TimerGroup
contains three 32-byte native Timer records and the second Current remains
at offset36. A nil group faults. Input records and caller storage are not
written by the original body.

Typed Go, actual public C and the existing LegacyGo wrapper compare complete
32-bit returns, the full 96-byte group and surrounding guards. The legacy
Go int parameter retains the old C.int narrowing. Original raw-register,
stack and CW/TOP observations are retained diagnostics, without corresponding
native instruction-count or register/stack equivalence claims.

ROOT/timer-percentage-native-entry-port preserves frozen declarations,
reference executions, pre-edit consumers, valid native Go/typed-C high-pointer
getter controls, and the actual old public C SIGSEGV. Existing inputs,
assertions and expected fixtures are unchanged. The existing public
int(void*,int) ABI retains two PE32 argument words. The old C body is archived.
The registered old playback-rate callback is distinct from modern native
AudioFX initialization; no user sale-silence or full old audio-chain fix is
claimed by this leaf port.
