# macOS ARM64 completion evidence

The current engineering estimate is **79%**. This is a weighted maturity
estimate, not measured function, feature, line or branch coverage. Passing
tests or adding sealed original ranges does not supply that denominator.

| Area | Weight | Estimated maturity | Remaining completion evidence |
| --- | ---: | ---: | --- |
| Build and native ABI | 15 | 95 | Applicable four-configuration suites and classification of reachable stock callbacks/layouts |
| Gameplay | 25 | 80 | Executed three-class feature matrix, including items, AI, death, pets and objective modes |
| Rendering and UI | 15 | 80 | Host Load translucency diagnosis, fresh-load pixel discrepancy, normal/highres SDL input and compositor checks |
| Audio | 10 | 85 | Production backend lifecycle under ordinary game time and audible output |
| Networking | 15 | 85 | Independent socket peers, replication, transitions, reconnect and original protocol compatibility |
| Campaign | 15 | 60 | All three classes through chapter endings, saves, reloads and progression |
| Stability | 5 | 50 | Bounded ordinary-clock load/combat/transition soak with resource and shutdown checks |

The calculation is `(15*95+25*80+15*80+10*85+15*85+15*60+5*50)/100 = 79`.
All seven areas remain partial. Physical second-machine, WAN traversal and
overnight runs can strengthen evidence; they are not additional implicit
completion requirements.

## Current findings

The native FloorMap conversion and its remaining original-coordinate
discrepancy are documented in [native-floor-map.md](native-floor-map.md).
The separate incoming Quest player defect wrote the PE32 `PlayerUpdateData`
offset 552 into native respawn markers. The C producer now calls a native
helper that writes typed `Field138` (offset 648 on ARM64). Regression checks
preserve every other byte and the original Quest/host/unit guards. The
existing CTF fixture gains only an aborting external-symbol stub; its inputs,
cases and expectations remain intact.

Original comparison uses REA/Ghidra 12.1.3 and `GAME.EXE` SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The codec and allocator bodies and incoming-player body are sealed in the
code oracle, retaining every existing range. These are static observations;
the original executable has not been run by this investigation.

Stock sprite scanning found 139,080 images, 126,687 compressed streams and
21,522,854 runs. Only opcodes 1, 2, 4 and 5 occurred; speculative opcode 6/7
changes therefore do not explain translucency in those unchanged assets.
Retained Host Load samples show normal player status and disabled layer
alpha. These observations do not resolve the reported visual symptom.

## Campaign observations and test integrity

On clean revision `bd4c363dff7e34b366d54df88535ae8d0208aef2`, the normal
ARM64 client passed the existing Conjurer headless scenario: bear damage and
release, summoned spider replication, necromancer damage/removal, Charm
aggression/follow behavior and escape open/close. The scenario positions the
player and switches maps, so it does not establish complete chapter traversal.
REA receipt: `ev_1062cf2649743dd389fe4b1b2900c8e10dfbc265f85c44d690f04fac2206ac3e`.

Manual save/refuse-overwrite/overwrite/reload checks passed. A fresh process
restored exact player statistics, inventory and position. Its unchanged
existing screenshot golden differed by 889 of 786,432 pixels (0.113%),
including a blue actor and foliage. The actor's identity is not established
by the pixel comparison. Two reruns produced the same actual image.
The full fresh-load scenario remains failed; the golden is preserved.

Collecting failures with `NOX_E2E_FAILFAST=false` exposed a shutdown defect:
the first close event could exit before a later scenario callback reported
the accumulated error. Such an exit code 0 is not accepted as a pass. The
shared E2E shutdown hook now propagates stored failures while retaining
successful quit behavior and existing fatal errors. The unchanged failing
scenario was rerun in collection mode: saved-state restoration and movement
still execute, followed by normal cleanup and exit 2 instead of exit 0. Its
actual screenshot bytes remain identical. REA receipt:
`ev_66cd97283b1025b703b0c928ae36c6c291f93fd7959cbe2025599f2845e5b89f`.

A missing screenshot golden is automatically created by the existing
recorder even when override is false. Therefore each visual parity claim
must also establish that the expected file existed before execution. A
newly created image supplies no comparison evidence.

These campaign captures use isolated data copies, headless rendering,
virtual game time and disabled audio. They establish neither compositor
parity nor audible output nor an ordinary-clock stability run.
