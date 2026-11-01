# Original special-state sender 527E50

The unchanged GAME.EXE has SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The complete root at `00527E50` has 470 instruction bytes, SHA-256
`c7157fcaa70667410fa5633c45a59bad89aaff83e8cb91bdc6168f149531767f`;
the following ten alignment NOPs are excluded.

Independent declarations were frozen before original/native execution:
1,252 contexts, 1,272 calls and six separate raw nil-chain faults per profile.
All 470 root bytes execute on valid inputs. Original `0xe3f` and `0x23f`
outputs are byte-identical. This integer root performs no FPU arithmetic.

| Artifact | SHA-256 |
| --- | --- |
| Frozen input gzip | `4efe17e42281d76cde882232e932508619b8405a83f9a55cfcd5950e84c9d61d` |
| Original output gzip, both profiles | `5b06030fdcf4683598c04a69b61e65966f35a190df00d5b9c7b91c7f0c7a3a86` |
| Decoded original output | `5ef4e65a74a11f708054c3d79a45d175d613ca69244695ecdb774dc2ed83c37e` |

The original loads recipient.UpdateData, its Player, and the unsigned index
byte before testing the unit pointer or index32 boundary. It caches that index
through all callbacks. Eight ordered live-word tests cover animation, health,
hidden, extended status, height, enchantments, team modifiers and NPC state.
Enemy nonzero suppresses health reporting while retaining the health-bit clear.
Each bit clears after its callback; callback changes to later bits are visible.
Unhandled mask bits admit the root and remain unchanged. Return words are 0/1.

All nine report/enemy child bodies are declared service boundaries. This
reference executes the complete root; it does not execute those original child
bodies or establish their complete-body equivalence. Report return words are
declared unused inputs. Raw root-write traces, original stack/register/FPCW/TOP
checks are diagnostics, without native CGo stack or write-count claims.

Native C and Go consumers compare ordered service arguments, live flags at each
boundary, full returns, all 32 peer words and complete guarded Object, two
PlayerUpdateData and two Player records. Native64 records must exceed4GiB.
The recipient has class0 like the original allocation, preventing an accidental
new raw-entry class guard. Separate ordinary-production checks retain the
existing class/nil policies, exercise the factory through C/public Go/ordinary
adapters, and check report packets at the declared lower transport boundary;
height uses the real client netlist. They add no child-original-body claim.

`ROOT/object-special-flags-native-entry-port` retains both original executions,
their producer, frozen declarations, C-before getters and actual SIGSEGV,
all comparator versions and raw build/test results. The first C-before command
failed to locate a relative overlay before compilation; the absolute-path retry
used the same frozen source/input. The first combined overlay's legacy consumer
failed its generic allocation constraint; changing `any` to `comparable` leaves
every original input, output, seal and assertion intact. Earlier drafts remain.
