# Original killable predicate 528190

GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The complete 49-byte body has SHA-256
`e862f2b35fe38c43eceff07d9e058e334aff486967f7510274bd91ee4e848eea`.
The following 15 NOPs are separate alignment bytes, already sealed in the
existing oracle; they are excluded from execution coverage.

Independent input declarations contain 147 valid cases and one nil-Object
fault. Both original FPU profiles execute every body byte without child,
instruction or arithmetic substitutions. Their output bytes are identical.

| Artifact | SHA-256 |
| --- | --- |
| Frozen input gzip | `536ee1254fa9f85487665c3ea5e46209f2c90b5766b61b9a44dd6d58382917a6` |
| Original output gzip | `43c192733569c7b8e413805f8e8a20bbbdd34c7665eda62f8a893e2e864468af` |
| Decoded original output | `2734e78743632c23916542c1e8c4da35ac5c390fbfa06358d728b78773c81f7c` |

The raw root first reads Object.HealthData. A nil Object faults. Nil health
returns0; otherwise `(Cur != 0 || Max == 0)` returns1, including Cur0/Max0.
Class is not read. The existing oracle address, size and hash are correct;
its older human description incorrectly said both-zero health was rejected.
Only that description is corrected after unchanged-original execution.

Raw C, raw Go and the existing guarded Go wrapper compare full 32-bit returns,
complete native Object/20-byte HealthData records and neighbor guards. WORD
neighbors are poisoned to detect accidental DWORD reads. Native64 records must
exceed4GiB. Original register, stack and CW/TOP observations remain diagnostics;
no native raw-register, stack or instruction/read-count equivalence is claimed.
The guarded wrapper and FIGHT/sight helpers retain their existing nil-Object
false policy, separate from the raw original fault.

ROOT/killable-public-native-entry-port retains both reference executions,
producer/declarations, source snapshots, frozen consumers and failed actual
pre-port C entry. Valid native Go/typed-C getters pass first, then public C
SIGSEGV reproduces with the same Object. No old input or expected value is
regenerated from native execution. The first native overlay passes449new and
five unchanged controls, 454 total, without consumer repair or build warnings.
