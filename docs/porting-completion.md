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

The native FloorMap conversion, corrected original coordinates, signed
bounds and whole-byte flag behavior are documented in
[native-floor-map.md](native-floor-map.md).
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

The conditional classic map-generator floor path now uses native brush
pointers, cell/head ownership, node links and scalar flood coordinates.
Rectangle/column producers use contiguous float pairs and exact original
binary32 operands. Controlled native cases cover these operations; the
documented `543680` initialization policy defines previously unspecified
transient state. The ordinary server's `#name` entry needs external
theme/AreaMap inputs absent from the retained stock map inventory. Full
generation, upstream theme/file/conditional-parser pointer conversions and the remaining
callback/layout classification remain open.

Three fill producers now use contiguous float coordinate pairs, exact
binary32 operands and the original double-to-float staging. Their 15 fixed
vectors exercise the actual token registry, native fill pointers, callback
and RNG order, and canaries in optimized and ASan+UBSan execution. They add
33 macOS terminal events. The integrated native/declaration-guard union
passes 164 events normally and under race, checkptr level 2 and cgocheck2.
These controlled results do not establish complete classic generation.

The spell-name lookup helper now restores its original 60-byte output array.
Retained REA/Ghidra instructions and pseudocode identify two adjacent 60-byte
arrays; all 139 bytes of the original body match an independent PE read and
are added to the code oracle without changing previous records. The helper
retains its uppercase conversion, `SPELL_` prefix and lookup handoff. Its
source-bound regression uses the production formatter with 310 bounded ASCII
inputs per optimization/sanitizer mode. The tested input lengths are 0 through
53 bytes so the prefix and terminator fit; arbitrary-length input and complete
classic generation remain unverified.

For this correction, the Darwin ARM64 normal and highres suites each pass
81,906 terminal test events with 47 skips; server and server+highres each
pass 77,078 with seven skips. These totals include parent/subtest events.
All previous test outcomes and multiplicities remain, with four new passing
events from the spell-name regression. Only the existing two subtest families
whose names contain runtime pointers above `UINT32_MAX` need address-label
normalization for comparison; nil and fixed low addresses remain literal.
The focused mapgen union passes 148 events both normally and with race,
checkptr level 2 and cgocheck2. All 3,245 pre-existing test/fixture/YAML files
retain their SHA-256 values, and the oracle verifies 2,995 code and 656 data
ranges. These results retain the overall 79% maturity estimate.

The retained clean R306 normal product also completed the headless Wiz08e
tower scenario through REA process capture. The stock tower selected and
naturally launched at player, monster and NPC victims; server and client HP
agreed on damage of 34, 35 and 35 respectively, and the process exited normally.
These were direct map loads with controlled victims, virtual game time and
audio disabled. Three software frames were observed without golden comparisons;
the result does not establish chapter traversal, endings or compositor parity.

The Quest reward decoder now retains native UTF-16 title pointers in all six
self/other spell, guide and ability locals. Only their declarations change;
lookup, player identity, live names, formatting and output order remain.
The existing actual CGo dispatcher test reproduced a SIGSEGV with `-O0`
before this correction. Optimized Apple ARM64 code had retained full pointer
registers despite the incorrect `int` declarations, masking the defect.

A new source-bound regression injects those six actual declarations into
the unchanged reward/settings fixture. Its 2,160 reward and 512 settings
inputs, expected values, varargs checks and buffer checks remain intact in
each of the O0, O2 and ASan+UBSan modes. The unmodified baseline fails the
O0 reward contract at case 541. An initial sanitizer compile failed on the
current Darwin SDK's deprecated `sprintf` declaration; that log is retained.
The new runner suppresses only that Darwin diagnostic while preserving the
fixture and fatal sanitizer checks.

After the declaration correction, the unchanged actual CGo test passes all
156 isolated child cases with `-O0`. The focused union passes 13 outer Go
events both normally and with race, checkptr level 2 and cgocheck2. Its new
source-bound test supplies ten of those events, including parents and groups.
Child stdout is not counted again in the outer terminal totals.

Stock sprite scanning found 139,080 images, 126,687 compressed streams and
21,522,854 runs. Only opcodes 1, 2, 4 and 5 occurred; speculative opcode 6/7
changes therefore do not explain translucency in those unchanged assets.
Retained Host Load samples show normal player status and disabled layer
alpha. These observations do not resolve the reported visual symptom.

## Other platform observations

The existing Windows gate for clean R304 revision `f2efbe9b1` built, verified
and ran help for both 386 and amd64 server products, then failed its Go test
step. The prior `9116f3164` and `730981c07` runs failed that step too. These
Windows results are separate from the four passing macOS configurations.
See the [R304 Windows run](https://github.com/hjyoon/opennox/actions/runs/39023157802).

The 386 diagnostics include existing test constants overflowing `int` during
compilation. The amd64 failures include resource exhaustion, file/fixture
platform assumptions and an existing actual Quest reward CGo dispatcher
fault. CRLF simulation reproduces several source extraction/hash failures,
but the remote checkout bytes were not captured, so the EOL explanation
remains an inference. The reward pointer declarations described above are
corrected from a failure reproduced on macOS as well. This does not establish
a Windows runtime pass for the candidate or resolve the independent resource,
386 test-constant, EOL or fixture-lifecycle findings. Existing tests and
fixtures remain unchanged.

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

## Independent loopback peers

The new `scripts/e2e/client-live-peer.yaml` exercises discovery, class/color
selection, a bound local drawable, movement and disconnect against a separate
dedicated server. It requires an isolated local server and
`NOX_E2E_REAL_SLEEP=true`. This opt-in mode retains virtual ticks and the
deterministic RNG, while platform `Sleep` also waits in real time. The default
E2E clock is unchanged. Production networking timeouts and retries are unchanged.

With the same first candidate executable, the default clock exhausted the
socket polling loop with error -23 and no registered player. Real Sleep
allowed the socket handshake, but its virtual-time jump made subsequent
scheduled steps overdue. The `checkpoint` action now queues the next phase
from the clock at phase completion. A separate `wait-client-connected`
action first requires the connected flag and matching positive player/drawable
netcodes, with a 500-engine-iteration timeout. The original scenario's
500/120/30 join, movement and settling durations remain intact.

The final development candidate SHA-256
`4d51a31c1964d0d2d3685c986668df27611f624488eca228f507dd69196684bb`
and clean `9116f3164` server SHA-256
`79971a621582c30f69582fa64f1c22e5fa430b507c87ae6193d995b43da1da73`
passed the complete loopback run. Across 177 API samples the player count
changed 0 to 1 to 0. All 54 observed server/player/drawable bindings agreed
on netcode 185 and stayed within eight world units. The server player moved
231.002 units and its replicated client position followed. Both processes
exited naturally with code 0; the server used authenticated quit. The REA
capture and selected filesystem observations were untruncated, and owned
process cleanup completed. Receipt:
`ev_cb9c61cc8a559f80363cc0a5b7cd2e7d0088da45487411b935cc794370e1872b`.

The client was a source-bound development build with `vcs.modified=true`.
This run establishes actual loopback socket and movement behavior with a
virtual-clock client and ordinary-clock server. It does not establish a
fully ordinary-clock client, map transitions, reconnect, WAN traversal or
runtime compatibility with the original Windows executable. Captured frames
are observations, not screenshot-golden comparisons.

Clean revision `730981c0758c8bfff709943f2f8691ce36589956` subsequently
passed both normal and highres client/server pairs. The normal run observed
158 samples, 44 bindings and 193.198 units of movement; highres observed 157
samples, 43 bindings and 115.264 units. Both retained the 0-to-1-to-0 player
lifecycle, matching netcodes, natural process exits and complete captures.
All four product hashes and personal files stayed unchanged. Receipts:
`ev_90c3c3cfbcce65a0aed81beed677d7fc21b45d2036523a91b9b8fbc1cdbf5b39`
and `ev_d4ea07b3028a63af9cfdedc887af970e09df966973347779c167b706e4ff55ae`.

## Map handoff, fresh connection and observer participation

A later private scenario kept the same clean `730981c0` server process and
endpoint while changing `so_beach` to `estate`, disconnecting client 1 and
connecting a fresh client 2. Its first run failed overall with controller
exit 1: map handoff passed, but client 2 remained in the arena's normal
observer state and did not move. That failure and its input are preserved.

Source and packet observations establish the prerequisite: a non-Coop,
non-Chat participant enters as an observer and uses Action to join play.
The connection gate establishes a bound drawable; a movement scenario must
also follow the game mode's participation flow. A separate diagnostic adds
the existing normal Action input for client 2. It retains the original
500/120/30 join, movement and settling steps and the bounded observer checks.

That diagnostic passed with the same server process throughout. The map
handoff completed in 13.417 seconds within its 20-second bound, with real
client map loading, rebinding and 114.817 units of estate movement. Fresh
client 2 then sent Action, cleared only the observer status bit, acquired
normal unit flags and moved 200.881 units. Its maximum observed replication
distance was 4.445 units. Both clients and the server exited naturally with
code 0. REA retained all 239 raw/rendered frames with no selected filesystem
omissions; original personal files and product hashes stayed unchanged.
Receipt: `ev_a9e72c7639f543883e1490dd9c2e67095bd0317252147d96ca069bf93568dd47`.

The HUD was observed; MOTD dismissal was not established. Both peers already
held the estate assets, so this run covers map handoff and fresh-process
reconnection, not asset download. It does not establish combat, every game
mode, a fully ordinary-clock client or original Windows-peer interoperability.

## Other runtime evidence

The unchanged native-audio Warrior scenario passed on clean `bd4c363df` in
294.137 seconds using OpenAL Soft 1.25.2's null backend. It exercised stream
playback, pause/resume, buffer consumption and context/device cleanup with
real worker time and virtual game time. This is lifecycle evidence, not
audible output. Receipt:
`ev_c2616791f51ac78639e80ce12ebf4c6b740db68d55dc8387112f083620b7e46a`.

Clean `9116f3164` dedicated server ran for 300.002 real seconds with all
`NOX_E2E*` variables removed. The run collected 61 samples, made three map
transitions and ended with authenticated quit and natural exit 0. Goroutines
remained 8 to 11 and ended at 8; final heap allocation was 55.2 MB. This
zero-player, audio-disabled idle/load run does not establish combat or
replication stability. Receipt:
`ev_5e0adef52668070bb70fa4b4553520d6d9ff0d919533d33c3e6b0b53a6192097`.

## R306 Quest title pointer validation checkpoint

| Configuration | PASS events | SKIP events | Passed packages | FAIL |
| --- | ---: | ---: | ---: | ---: |
| Normal | 81,902 | 47 | 40 | 0 |
| Highres | 81,902 | 47 | 40 | 0 |
| Server | 77,074 | 7 | 36 | 0 |
| Server, highres | 77,074 | 7 | 36 | 0 |

Each mode retains every R305 terminal identity, action and multiplicity and
adds ten pass events. All 3,244 existing tests, fixtures and E2E files retain
their hashes; no existing input, assertion or expected value changes. The
selected 35 source/fixture dependencies remain unchanged throughout these
runs. All 1,562 personal installation files retain their baseline hashes.
The focused union passes 13 events normally and with race, checkptr level 2
and cgocheck2; the unchanged actual dispatcher also passes separately at O0.

Only six local declarations in the decoder change from `int` to `wchar2_t*`.
All other bytes of that C file remain unchanged. The initial actual CGo O0
SIGSEGV and source-bound O0 varargs failure remain recorded beside the final
passes. These results establish the native pointer contract for the tested
reward paths and do not supply a complete gameplay or campaign denominator.

REA observations seal three reward-arm subintervals totaling 571 bytes
inside the formal `0048EA70..004942D1` dispatcher. They are not three complete
function bodies. The code oracle verifies 2,994 code and 656 data ranges,
retaining all earlier records, their relative order and metadata. The
original executable is unchanged and was not run.

## R305 fill producer validation checkpoint

| Configuration | PASS events | SKIP events | Passed packages | FAIL |
| --- | ---: | ---: | ---: | ---: |
| Normal | 81,892 | 47 | 40 | 0 |
| Highres | 81,892 | 47 | 40 | 0 |
| Server | 77,064 | 7 | 36 | 0 |
| Server, highres | 77,064 | 7 | 36 | 0 |

Each mode retains every R304 test identity, result and multiplicity and adds
33 pass events. All 3,242 existing tests, fixtures and E2E files remain
byte-identical; no existing expected value changes. R134 and the whole-header
declaration guards remain unchanged. The integrated 164-event native/guard
union also passes race, checkptr level 2 and cgocheck2.

The original fifteen fill vectors pass both optimized and ASan+UBSan runs,
including the actual registry and allocations above 4GiB. Only the three
specified producer bodies change in `GAME4_2.c`; all its other address
sections and header bytes remain. The oracle verifies 2,991 code and 656 data
ranges with every prior record retained. Initial unsorted-record validation
failed and is preserved; inserting new records by address corrected the
manifest order without changing compiled code or existing seals.

All 1,562 personal installation files, totaling 571,159,691 bytes, retain
their baseline hashes. Original REA observations, controlled C callbacks,
macOS Go suites and the separate Windows findings have distinct scopes.

## R304 native floor validation checkpoint

| Configuration | PASS events | SKIP events | Passed packages | FAIL |
| --- | ---: | ---: | ---: | ---: |
| Normal | 81,859 | 47 | 40 | 0 |
| Highres | 81,859 | 47 | 40 | 0 |
| Server | 77,031 | 7 | 36 | 0 |
| Server, highres | 77,031 | 7 | 36 | 0 |

Each configuration retains every R303 terminal test identity, result and
multiplicity, adding 52 pass events. Only the native allocation addresses in
two existing fault-boundary test families are normalized; literal zero and
fixed low addresses remain distinct. These event totals include parents and
subtests and do not measure feature coverage.

The floor/region/brush union contains 124 events. It and the seven unchanged
declaration guards pass normally and with race, checkptr level 2 and
`GOEXPERIMENT=cgocheck2`: 131 events per run. The 37 brush/list/flood cases and
eight rectangle/column cases also pass optimized and ASan+UBSan C execution.
Two brush cases define deterministic initialization of unspecified original
stack state; two others check the legacy-token adapter with a controlled
resolver. These are distinct from original-defined behavior and execution
through the real token registry or a complete generator.

The first normal/server runs failed seven existing guards because a native
prototype change altered the whole-header SHA. The product now exposes its
typed entry through a dedicated header and preserves the old token ABI with
a resolver wrapper. `GAME4_2.h` and all 3,238 existing tests, fixtures and E2E
files remain byte-identical to R303. No existing expected value changed in
R304, and R134 remains unchanged. All 1,562 personal installation files,
totaling 571,159,691 bytes, retain their baseline hashes. Initial failure logs
remain preserved alongside successful final runs.

The oracle verifies 2,988 code ranges and 655 data ranges, retaining every
previous record. REA/Ghidra analysis adds 14 original complete procedure
bodies and three constant ranges; the original executable was neither run
nor modified. Controlled coverage and the remaining generator limitations
are detailed in [native-floor-map.md](native-floor-map.md).

## R303 native and harness validation checkpoint

The combined floor-region and live-peer harness changes passed all four
configuration suites with Go 1.26.9 on macOS ARM64:

| Configuration | Test pass events | Skips | Passed packages | Failures |
| --- | ---: | ---: | ---: | ---: |
| Normal | 81,807 | 47 | 40 | 0 |
| Highres | 81,807 | 47 | 40 | 0 |
| Server | 76,979 | 7 | 36 | 0 |
| Server, highres | 76,979 | 7 | 36 | 0 |

These are terminal Go test events, including parent tests and subtests, not
feature coverage. Each configuration retains all prior terminal cases and
adds 75 pass events: 57 native checks and 18 harness checks. Comparison
preserves multiplicities while normalizing only native allocation addresses
in the names of two existing fault-boundary tests.

The native focused union passed 74 events with race, checkptr level 2 and
`GOEXPERIMENT=cgocheck2`. The final E2E subset passed 27 events under the same
strict settings. Source-bound optimized and ASan+UBSan C runs also passed.
The oracle verifies 2,974 code ranges and 652 data ranges, retaining every
previous record; these totals do not measure original-function coverage.

All 2,744 preserved Go test/YAML files remain byte-identical. The only changes
to existing C fixtures in this checkpoint are the approved original-derived
R300/R302 expected row locations and the corresponding exclusion check.
Inputs, cases, assertions and canaries remain. R134 is unchanged. All 1,562
personal installation files, totaling 571,159,691 bytes, retain their baseline
content hashes.
