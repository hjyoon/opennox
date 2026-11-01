# Original unit height reporter 4D9020

The original GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The complete 191-byte body has SHA-256
`e744b412ef9521c8ee8e171663f0694f7a9d169fa79276aea2d057bebfe3b4bb`.
The following single NOP is separate alignment, excluded from execution.

Unchanged root and actual 39-byte QWORD converter execute 370 raw contexts
and one separate nil-Object fault in each original FPU profile. Unit-code
lookup and netlist submission are declared raw child boundaries. Another
544 independent production contexts execute the actual unit-code helper;
only netlist submission remains a declared boundary. Its 55 reachable bytes
execute; the three-byte nil-only child return is unreachable through a valid
root, which faults first on a nil Object. Both profile outputs are identical
for each dataset. No instruction or arithmetic substitution is used.

| Artifact | SHA-256 |
| --- | --- |
| Raw input gzip | `cd988d73882de912f60ddec00b978adae6d2938636785a0cc3e15755b945d9be` |
| Raw output gzip | `62d9899356eba94b9b6f57dbc7097f492ae258bf96de57c91f1010c469429516` |
| Decoded raw output | `0234c77edee395b6245412978d6396c02c19179ac59377177109c6b9fce60855` |
| Production output gzip | `efffd2dc94547a837240af0de68e9e970dc8df8b3aa4a333d48a3049e7d3fe3b` |
| Decoded production output | `17f2d7c1bd41df592f18df9c1eb59136f189d9d1a040ed60db5584dd065892eb` |

Field5 bit0x20 chooses a six-byte packet with a direct WORD network code
and three separately truncated low bytes. Otherwise FCMP treats negative
and unordered height as opcode0x5f and nonnegative height as0x5e. This branch
precedes the code child; height reload follows it and negation follows the
cached branch. Both branches submit kind1 and preserve the complete return.
The existing signed-QWORD converter's integer-indefinite low byte is zero.

Native comparisons retain complete Object and neighbor guards, callback
mutations, child arguments/order, packets and 32-bit returns. Production
comparisons use real unit-code lookup and real netlist acceptance/full-queue
rejection. Original register/stack/control-word/TOP observations are retained
diagnostics, without native stack or instruction/read-count equivalence.

ROOT/unit-height-native-entry-port retains independent declarations, both
original profile executions, frozen consumers and unchanged-source before
results. Valid above4GiB Go/typed-C getters pass, then the old actual public
C entry reproduces SIGSEGV. The existing ordinary Go sender independently
reproduces an opcode0x5e versus original0x5f quiet-NaN discrepancy; that
specific before check compares only the opcode chosen before its code child.
No existing input, assertion or expected fixture is regenerated from native
execution. The original C body remains archived after porting. Existing
public int(int,object*) and its two PE32 argument words remain unchanged.
