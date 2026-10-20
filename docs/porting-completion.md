# macOS ARM64 completion evidence

## Current campaign scope

The latest user instruction limits campaign verification to **Warrior,
Conjurer and Wizard, chapters 1 through 3 only**. This covers nine class/chapter
cells and every stock submap in those chapters. Chapter 4 and later gameplay,
and full-story endings, are excluded from the completion requirements.
An original transition into chapter 4 may establish the chapter 3 boundary;
no chapter 4 objective progression is required or pursued.

Each chapter needs original earned progression and required objectives,
class mechanics available in these chapters, and original Save/Load evidence.
Existing build/ABI, UI, audio, networking and stability requirements remain;
later-chapter content must not become an implicit verification gate.
Historical sections below retain the earlier broader scope, which this
explicit user restriction supersedes.

| Class | Chapter 1 stock maps | Chapter 2 stock maps | Chapter 3 stock maps |
| --- | --- | --- | --- |
| Warrior | War01a | War02a, War02b | War03a, War03b, War03c, War03d |
| Conjurer | Con01a | Con02a | Con03a, Con03b |
| Wizard | Wiz01a | Wiz02a, Wiz02b, Wiz02c | Wiz03a, Wiz03b, Wiz03c |

The nine-cell inventory distinguishes earlier product evidence from current
runtime evidence. These states are not completion percentages. A map list
or a process exit does not prove a complete chapter.

| Class | Chapter 1 evidence | Chapter 2 evidence | Chapter 3 evidence |
| --- | --- | --- | --- |
| Warrior | Earlier original completion and chapter Save/Load; current replay pending | Gauntlet partial; zipper/Horrendous completion pending | Completion unproven in inspected records |
| Conjurer | Earlier natural Con02a entry and manual Save/Load | Earlier original objectives and natural Con03a Save/Load | Original rescue, Foreman, Captain and ending boundary proved; normal/HD ordinary earned-Save Load; complete current replay pending |
| Wizard | Earlier original completion and Wiz02a AUTOSAVE Load; current replay pending | Entry and MeetHorvath objective observed; completion pending | Completion unproven in inspected records |

The historical full-game engineering estimate is **79%**. It predates the
chapter restriction and is not a measured completion percentage for the nine
class/chapter cells. This is a weighted maturity
estimate, not measured function, feature, line or branch coverage. Passing
tests or adding sealed original ranges does not supply that denominator.

| Area | Weight | Estimated maturity | Remaining completion evidence |
| --- | ---: | ---: | --- |
| Build and native ABI | 15 | 95 | Classification of remaining reachable stock callbacks/layouts after the passing integrated four-configuration suites |
| Gameplay | 25 | 80 | Executed three-class chapters 1-3 feature matrix, including available items, AI, death, pets and objective modes |
| Rendering and UI | 15 | 80 | Remaining Host Load reports, unchanged fresh-load golden discrepancy, broader SDL input and compositor checks |
| Audio | 10 | 85 | Perceptual audible-output fidelity and broader long-session playback after the bounded ordinary-clock core-backend run |
| Networking | 15 | 85 | Broader peer/gameplay matrix and original Windows-peer runtime compatibility |
| Campaign | 15 | 60 | Complete all three classes in chapters 1-3, their original objective routes and chapter saves/reloads; later chapters and full endings excluded |
| Stability | 5 | 50 | Broader peer load and sustained gameplay beyond the bounded host and independent-peer profiles |

The calculation is `(15*95+25*80+15*80+10*85+15*85+15*60+5*50)/100 = 79`.
All seven areas remain partial. Physical second-machine, WAN traversal and
overnight runs can strengthen evidence; they are not additional implicit
completion requirements.

## Current findings

R341 extends the earned Conjurer chapter 3 chain through its original ending.
Original mine-exit movement returns to Con03a. The return point lies inside
the entrance cabin; an actual left click and movement through its original
door permit the linked road to the current Captain at 1864,3319. Original
Talk, 10 native delivery ticks and Done twice produce Captain phase 1->2->3
and Con03FindAirshipCap4. A genuine menu Save retains health 29, poison 0,
mana 85, gold 1,013 and all five living workers in the archived native mine.
The inactive map's xfer flags are decoded read-only with the native FlagDead
enum, rather than treating absent workers in the active outdoor map as dead.

Independent original menu Loads of that same phase 3 Save and the third
actual Captain Talk naturally switch to Con04a. Virtual and ordinary normal
processes first observe the transition during its black fade. Fresh ordinary
normal and highres cases allow 120 normal rendering ticks and visibly show
the original chapter 4 briefing. They quit without dismissing the briefing,
pursuing chapter 4 objectives or writing a new manual Save. Only the chapter
3 completion boundary is claimed. These are earned-Save continuations, not
a fresh entire chapter replay on the current revision.

There are 8 executed captures, 6 exit 0,
2 negative exit 2, 2
valid living manual Saves and 16 measured arrival
gates, with zero PlayerDeath or forced termination. Prepared case 7 is never
run; its immediate-fade route is superseded by visible boundary case 8.
The two failed movement routes write no Save and never authorize continuation.
No engine defect is inferred. Each script snapshot uses its own exact
original function identity/global layout; the chapter 4 VM remains uninterpreted.

The latest scope remains all three classes, chapters 1-3 only. Warrior
gauntlet/Horrendous/chapter 3, Wizard chapters 2-3 and reconciliation of earlier
chapter evidence with the current products remain open. Broader scoped
build/ABI, UI, audio, networking and stability requirements remain partial.
No new nine-cell completion percentage is assigned; historical 79% remains
the earlier full-scope engineering estimate. Overall 100% is unproven and
the goal remains active.

Engine/products and all gates remain sealed R334 b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a:
58 focus, 58 strict actual cgocheck2/race/checkptr2, normal/HD
82,108 PASS/49 SKIP each and server/server-HD 77,280 PASS/7 SKIP each, zero
FAIL. No fresh suite or build is claimed. Preservation rehashes 6,645 source,
3,284 protected and 1,562 personal files, each stock clone and all 24 prior
seals including the separate R335 repair.

### Retained R340 findings

R340 proves the original five-worker mine rescue in Conjurer chapter 3.
Original Hunt and moving Charm inputs clear Wolf 7214 before the lower-lift
Save; sole damage attribution is not claimed. At the third lift, 60 normal
ticks reach the upper passenger dwell window. Actual running off the shaft
before descent leaves the platform available to both following workers.
Their original Return triggers and AllFound execute: counter 21=5, state 15=7,
all worker states 16..20=3, RescueMineWorkers4 and native XP +1,000. Player and
all five workers survive the whole case and original menu Save.

An independent original Load and 7,000 normal native ticks recover health 12->29
and mana 85 without health, poison, clock or timer assignments. The original
second and first lift cycles are measured at 30/60/90 ticks; the 60 tick upper
dwell windows permit actual step-off and continuous return-road arrivals.
The original Foreman Talk needs native delivery time before Done; adding 10
normal ticks makes the real dialogue available. Original Done grants Gold500
(513->1013) and Con03FindAirshipCap2. No callback or reward is injected.

Independent ordinary normal and highres processes each load that same earned
Foreman Save through the original menu, perform two nearby walks, run 600
normal native frames and write new owned Saves. Both retain health 29,
poison 0, gold 1,013, counter 5, Foreman state 81=5, the Captain objective and all five
living workers. These are earned-Save continuations, not fresh chapters or
complete campaigns. Preferred Player.plr SHA256 is
672352922b2262240a43c679afa5a105ef8442fa8bd2d82ed64a7b0dcbaa480d.

There are 13 executed captures: 10 exit 0 with valid living manual Saves,
3 negative exit 2, 56 measured arrival gates and zero PlayerDeath. Wrong
cross-floor goals and premature dialogue inputs are preserved as failures
without new Saves. No forced termination, startup failure or prepared-only
case occurs. No port defect is inferred from those input preconditions.

The latest explicit user scope is Warrior, Conjurer and Wizard chapters 1-3
only, as recorded above. Nine cells/18 stock submaps are inventoried; earlier
product evidence is separated from current runtime results. Historical 79%
is not a newly measured percentage for this restricted scope. Captain,
Conjurer 3 ending boundary, Warrior gauntlet/Horrendous/chapter 3 and Wizard
chapters 2-3/current replay remain open. Chapter 4 objectives and full-story
endings are excluded. Overall 100% remains unproven and the goal stays active.

Engine and four products remain sealed R334 b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a.
Reused gates retain 58 focused/58 strict actual cgocheck2/race/checkptr2 passes
and exact full suites: normal/HD 82,108 PASS/49 SKIP each, server/server-HD
77,280 PASS/7 SKIP each, zero FAIL. Metadata and the full source archive are unchanged;
no new engine suite or build is claimed. Preservation rehashes 6,645 source,
3,284 protected, 1,562 personal files, every cloned stock and all 23 prior seals,
including the separate R335 repair. Existing broader UI/audio/network/ABI/
stability requirements remain, without later campaign content as implicit gates.

### Retained R339 findings

R339 consumes the earned GoldKey through original door movement, recognizes
Garrit and Dudley and completes Dudley's original Done dialogue. Native
observations retain state 18=1/19=1 and returned-worker counter 21=3. The
original shared recognition creates seven Imps and one Scorpion; all eight
new script identities are observed. No key, counter or callback is assigned.

Original jump/run input exits the guarded room and escorts both workers
through the north and inner roads. Garrit briefly stops at a corner; three
ordinary walks bring him alongside Dudley again without an AI or Follow
change. A separate original Load and 6,500 normal native ticks, after sequential
learned Charm inputs, recover health 10->29 and preserve two Imps and a Small.

The final successful return road uses three original moving Charm inputs.
Actual native durations are 63->108, 167->212 and 227->272, 45 frames each;
all three Imps are alive and owned before Save. Eight measured road arrivals
reach the lower lift room. Health 20->6->7, poison 0, and all five story workers
survive the whole case including the original Save and 180 post-Save ticks.
Dudley and Garrit finish 36 and 23 units from the player. Two saved Imps later
die during native post-Save combat; an original Load can restore the earlier
saved live state. No external revival or ownership assignment is made.

Waiting for recovery at that point instead loses Dudley. That case exits 0
through original story failure but writes no Save and is never a source.
Another case actually walks onto platform 6164 at 3302,4108 and observes the
player inside its box at Z4. After 90 normal ticks the player is still downstairs
and outside the box. It saves, then PlayerDeath occurs during post-Save
ticks; the whole branch is rejected. Dudley nevertheless reaches the upper
area before Save, sets state 19=2 and counter 21=4 and earns the original 250 XP
(native 7371.3003->7621.3003). Garrit and the player remain downstairs. This
proves one original worker transport/Return event inside a rejected branch;
it does not supply a safe new-return checkpoint. Counter 5 and AllFound
remain unproven. Combat and crowd displacement do not establish an engine
defect.

There are 17 executed captures: 11 exit 0, 6 negative exit 2, 10 valid living manual
Saves, 6 PlayerDeath branches and 45 measured arrival gates. The exit 0 count
includes the no-Save story failure. One genuine Save is rejected because
of later PlayerDeath. No forced termination, startup failure or prepared-only
case occurs. Preferred eligible original Save is case 15, native health 7,
poison 0, gold 513, state15=6/18=1/19=1/21=3, GoldKey 0 and all five
workers alive; Player.plr SHA256 is
28efb0e766c51915f6a240ddb2a5adcbdbf9fae204616a8bc74196c44f23e9b9.
This round does not claim fresh ordinary normal/highres continuation runs.

Engine source and products remain sealed R334 revision
b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a. Reused gates retain 58 focused and
58 strict actual cgocheck2/race/checkptr2 passes and four exact-terminal-
preserving full suites: normal/HD 82,108 PASS/49 SKIP each, server/server-HD
77,280 PASS/7 SKIP each, zero FAIL. Four ARM64 products and their metadata/full
source archive are unchanged. No new engine build or suite is claimed.
Preservation rehashes 6,645 source, 3,284 protected, 1,562 personal files, each
stock clone and all 22 prior seals, including the separate R335 repair.

REA supplies generic reverse-engineering and evidence tools. The OpenNox
completion scope comes from the inherited task record and this document;
actual campaign play is the agent's chosen validation method, not a REA
requirement. The user's provenance question does not add a new completion
criterion.

Next are protected boarding and physical worker returns, then the remaining
mine objective and ordinary normal/highres checks of the earned escort Save.
Complete three-class campaigns/endings, Warrior zipper/Horrendous, broader
UI/audio, original Windows-peer, reachable ABI/native-generator and sustained
stability remain open. Weighted maturity remains 79%, not measured coverage;
all seven areas remain partial and the 100% goal remains active.

### Retained R338 findings

R338 completes Alex's original follow-up dialogue, earned GoldKey and third-lift
transport. The actual Alex conversation and original Done change state 15=5->6,
grant original GoldKey 4849 from zero to one and enable Worker1Unlock 3096,
shaft 2740 and lower platform 6164. Returned-worker counter 21 stays 3;
Garrit and Dudley are still unrecognized. Alex is the original Worker1.

Actual boarding followed by 90 normal native ticks places the player at
3302,4108/Z24 at frame 182, on platform state 1/height 20 with reciprocal
shaft links and the player inside its native box. Original walking steps off
at 3410,4200 and reaches ground Z0, followed by a new healthy manual Save.
The earlier 24-tick wait does not complete that cycle in time and its lower
arrival times out. No clock, frame, position, Z, velocity or elevator field
is assigned. The nearby north door later passes but returns toward the lift.

Continuous original running crosses the lower east Scorpion road through
five measured arrivals at health 28/poison 0, followed by a new Save.
Scorpion 6132 remains alive; the owned Small 6118 dies in native combat.
Standing staff fights instead end in PlayerDeath. A separate optional room
fight leaves Troll 7193 DEAD after five actual staff left presses and native
combat; three later requests have no target and queue no attack. Native
recovery and original Save retain health 29. Chest-use and pickup branches
fail without a new Save; stable loot or a learned healing spell is not claimed.

The inner east room needs earlier Charm input and continued movement.
Original A has native range 300; the first distant unavailable request
queues no key and spends no mana. Later actual inputs start the earned Imp
Charms. Overlapping requests leave only the later pending duration, so no
simultaneous completion is inferred. Movement and eight normal ticks allow
the next attempted charge after the previous duration. The successful route
crosses the physical gap between disconnected NPC waypoint components and
continues through the inner west road. Original A at frame 460 starts Imp
6152's native Charm 459->504, 45 frames. Later object observations confirm
the same original Imp alive and owned. The north corridor Save has health
6/poison 0, no PlayerDeath, and all five story NPCs alive.

An independent original Load and 7,000 normal native ticks recover health
6->28/29 and mana to 85 without health, poison, timer or clock assignments.
The original Save retains two living owned Imps, 6152 and 6741. Imp 6741 was
saved alive before dying during the previous case's post-Save ticks; the
next original Load restores that earlier saved live pet. No external revival
or ownership change is performed.

Independent ordinary-clock normal and highres processes each load that same
earned recovery Save through the original menu, perform two actual nearby
walks, run 600 normal native ticks and write new owned manual Saves. Both
retain health 29/poison 0, returned-worker counter 3, state 15=6, all five
story NPCs and both living owned Imps. These are earned-Save continuations,
not fresh three-class campaigns. Preferred Player.plr SHA is
b3cd5c508d7653873fe1d56382a070b1501dc1cc4fe864089c3b9548aa82eecd,
with mana 85, gold 513 and the original GoldKey still held.

There are 25 observed game captures: 11 exit 0 with valid living manual
Saves, 14 negative exit 2, 83 measured arrival gates and 10 PlayerDeath
branches. Three branches write real Saves before later PlayerDeath; the
entire branches are rejected as continuation sources. There are no forced
terminations, startup failures or prepared-only cases. Every failed,
no-Save, death and poisoned branch remains excluded from continuation.

One prepared route inherited actual eligible source 5 while its prose said
source 6. The executed input and true seed are retained; a separate
provenance correction records the mismatch. Future preparation requires an
explicit expected source matching the actual path. Another source-only
handoff correction identifies extent 6882 as WoodenDoor, not GoldKey.
Neither correction changes engine code, executed inputs or prior seals.

Engine source/products remain b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a.
Sealed R334 gates are reused: 58 focused/58 strict passes with actual
cgocheck2, race and checkptr=2; four exact-terminal-preserving full suites,
normal/HD 82,108 PASS/49 SKIP each and server/server-HD 77,280 PASS/7 SKIP
each, zero FAIL. Four Go 1.26.9 ARM64 products, matching Go/REA metadata and
the 7,014-member full source archive are unchanged. No new engine suite or
build is claimed. Preservation rehashes 6,645 source, 3,284 protected and
1,562 personal files, each cloned stock and all 21 prior R319..R337 seals,
including the separate R335 reference repair. No original GAME.EXE
execution, personal writes or golden changes occur.

Next are Garrit, Dudley and their protected original returns. Four original
Scorpion guards remain outside their room; recognition, dialogue Done,
return counter 5, AllFound and the next mine objective are unproven.
Three-class stories/endings, Warrior zipper/Horrendous, broader UI/audio,
original Windows-peer, reachable ABI/native-generator and stability remain
open. Weighted maturity stays 79%, not measured coverage; all seven areas
remain partial and the 100% goal stays active.

### Retained R337 findings

R337 completes Claude's original recognition, escort and physical return.
The actual NPC crossing of Worker6Return trigger 3036 at 1883,2594 changes logical
20=1->2, returned-worker counter 21=2->3 and state 15=4->5. Native saved XP
increases by the original 250 to 6542.515. Health stays 29/29 and
poison 0 at every measured return-road point. All five mine workers remain
alive after original menu Save. Claude is observed near 1906,2467 afterward.

Independent ordinary-clock normal and highres processes each load the same
earned virtual return Save through the original menu, perform two actual
return-area walks and run 600 normal native ticks. Observed globals retain
20=2/21=3/15=5, with health 29, poison 0 and all five workers alive. Each writes
a new owned original menu Save. These are independent earned-Save
continuations, not fresh three-class campaigns. The preferred Save has
health 29, mana 85, gold 513, XP 6542.515, no surviving owned creatures and
SHA 5ec1cdaf252029e95067db8db60d2d2db35dafd2b5885ad094beed710f7a8d62.

Original walking first detours around warehouse furniture and crosses its
ordinary doors without consuming the held Ruby. Disconnected NPC waypoint
components do not establish a player collision defect: actual input crosses
the physical gap. Early guard fights fail; a continuous original sprint with
an actual Imp7119 Charm reaches the west loop with health 29/poison 0 and a new
Save. Earlier blocked chords and arrival timeouts remain negative evidence.

Actual learned A later controls Imp7133 and then Imp7138; the latter is
observed at 2 HP during the successful second charge and alive/owned after
Save. Native combat clears the other room Imps. Two passive-wait branches
instead end in PlayerDeath. Full object observations show Small7132 taking
a long original route around the cave. Early A attempts have no available
target, queue no key and spend no mana. Waiting for its physical arrival
then allows one actual A at frame 722, with native Charm 721->766 (45 frames).
Original Escort and 5000 normal recovery ticks preserve the party and let
health 19 recover naturally to 29/poison 0 before a new Save.

Original door movement consumes Ruby 1->0. Further original movement crosses
Worker6Rec (function 69), awards its 100 XP and opens Claude's original dialogue. Clicking
the current Done control runs Worker6DialogEnd (function 16), sets logical
20=1 and starts Follow. Counter 22 already equals 3 and
fails the original <=2 extra-key condition, so no fourth Ruby is granted.
Claude is the original Worker6. No quest callback, counter,
key, ownership, target or Follow field is assigned by the harness.

Protecting that Follow clears the returning-road Troll7195 with two actual
staff left presses plus original companion combat. Four extra requests have
no target and queue no attack. A later three actual staff presses plus
native combat kill Wolf7215; three extra requests queue no attack. Ten Small
watch requests have no available target, queue no A and spend no mana.
Another 5000 normal ticks recover native health 13->29. Original combat leaves
no surviving owned companions, while all five story NPCs stay alive.

The Claude recognition script requests seven Bat creations; five spawned
instances are observed. Later actual Bat staff inputs/native combat leave
1000297310, 1000297311 and 1000297313 DEAD. The remaining 1000297309 and
1000297312 are observed alive near 2295,3788 and
2249,3786. No complete Bat clearance is claimed. Tracking a fleeing Bat leads
the hero and Claude into an east alcove. Actual northwest mouse input exits
it, then eleven measured original east-road and aligned warehouse-door
walks bring Claude through the original return trigger described above.
The direct unaligned reverse door chord times out; aligned physical input
passes it. No collision, camera, AI, follow, health, poison, damage, clock,
timer, guide, spell or script state is changed directly.

There are 30 observed game captures: 20 exit 0,
10 negative exit 2, 20 valid living manual Saves,
123 measured arrival gates and 7 PlayerDeath branches. Every death,
negative and no-Save branch is retained and excluded from continuation seeds.
There are no forced terminations, capture startup failures or prepared-only
cases. Optional quit-raced reads are retained; missing reads prove nothing.

Engine source/products remain b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a.
Sealed R334 gates are reused: 58 focused/58 strict passes with actual
cgocheck2, race and checkptr=2; four exact-terminal-preserving full suites,
normal/HD 82,108 PASS/49 SKIP each and server/server-HD 77,280 PASS/7 SKIP each,
zero FAIL. The four Go 1.26.9 ARM64 products, matching Go/REA metadata and
7,014-member full source archive are unchanged. No new engine suite runs or
builds are claimed. Preservation rehashes 6,645 source, 3,284 protected and
1,562 personal files, cloned stock and all 20 R319..R336 seals including the
separate R335 reference repair. No original GAME.EXE execution, personal
writes or golden changes occur.

Next are Garrit, Dudley and the original next mine objective from state 15=5.
Counter 3 does not complete all five mine rescues or three-class stories and
endings. Warrior zipper/Horrendous, broader UI/audio, original Windows-peer,
reachable ABI/native-generator and stability remain open. Weighted maturity
stays 79%, not measured coverage; all seven areas remain partial and the
100% goal stays active.

### Retained R336 findings

R336 completes Logan's original recognition, escort and physical return.
Returned-worker counter 21 changes 1->2 and logical 16 changes 1->2 when Logan
crosses the original Worker2Return trigger at 1660,2481. He then reaches the
original waiting area near 1779,2502. The protected return keeps native player
health 29/29 and poison 0 at every measured road point. Original menu Save
writes a new owned checkpoint. All five mine workers remain alive.

Independent ordinary-clock normal and highres processes load that earned
virtual-clock Save through the original menu. Both retain logical 16=2/21=2,
health 29/29 and poison 0, perform two actual return-area walks, and write new
original menu Saves. These are Save continuations, not fresh three-class
campaigns from new characters. The preferred Save has health 29, mana 85,
gold 513, XP 5506.6514 and
SHA 5d3e76981454cc6d6ba8f8904e4f9d9e7334e34c1a0d67777358c28eb7dcab4a.

Earlier original A inputs recruit adult 7089 and small 7088. The latter has an
observed native 45-frame Charm duration and both original creatures are alive
and owned before/after Save. Original staff clicks and companion combat then
kill Troll 7075 and Troll 7197 outside Logan's door. Eight actual left presses
are observed; two later actions find no target and queue no attack. Neither
all damage nor both kills are attributed to the staff alone. Original XP
raises maximum health 20->29 and maximum mana 78->85.

Original level-up PauseFX holds gameplay for its strict 5000 platform-tick
duration. An early exit 0 attempt writes no Save while that effect is active.
Normal input-loop waiting lets it finish without changing clocks, pause,
frames or timers. Another real manual Save retains poison 1 and is rejected as
a continuation source. A fresh original replay lets poison decay naturally:
native health 24/poison 0 at 2148, then passive recovery 29/poison 0 at 5648, followed
by a new healthy original Save. No health or poison field is assigned.

Actual door movement consumes Ruby 1->0. A second movement crosses Worker2Rec,
runs the original 100 XP recognition and opens the original dialogue. Clicking
its Done control changes logical 16=0->1 and 22=2->3, grants the next Ruby and
starts original Follow. The original trap script creates 13 small spiders
and one Imp. The saved following Logan is alive and close to the player.

Failed escorts expose a story failure rather than a proven corridor bug:
Logan dies near 1028,3300 in two captured branches. One captures the original
failure Load menu while the living hero has 14 HP and native frame 969 stops.
The later arrival timeout is not an engine-crash or blocked-wall proof.
Other failed input routes and two player deaths remain rejected, with no
failed branch used as a continuation seed. An attempted Imp A queues no key
and spends no mana; no new Imp control is credited.

Protecting Logan resolves the observed failure. Two real A inputs recruit
the original ambush small spiders with script IDs 1000258999 and 1000259005,
each with a native 45-frame Charm.
One actual staff left press plus companion combat clears the encountered
lower threats; the other 11 requested strikes find no target. The player
recovers 23->28->29 with poison 0, and saves with Logan close and three living
owned creatures. Original Escort, short catchup waits at cleared corners and
normal walking then allow the actual return described above. No NPC AI,
follow target, ownership, position, script counter or quest callback is set
by the harness.

There are 22 observed game captures: 13 exit 0 and 9 negative exit 2, with
10 valid living manual Saves and 141 measured arrival gates. One separately
retained capture startup timeout has no game output/result and its cloned
Save remains unchanged. Its automatic cleanup is recorded separately
from naturally quiescent game exits. One prepared timing proposal is never
executed. Optional quit-raced reads are retained and never used as proof.

Engine source/products remain b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a.
Sealed R334 gates are reused: 58 focused/58 strict passes with actual cgocheck2,
race and checkptr=2; four exact-terminal-preserving full suites, normal/HD
82,108 PASS/49 SKIP each and server/server-HD 77,280 PASS/7 SKIP each, zero FAIL.
The four Go1.26.9 ARM64 products, matched Go/REA metadata and 7,014-member full
source archive are unchanged. No fresh Go suite runs or rebuilds are claimed.
Preservation rehashes 6,645 source, 3,284 protected and 1,562 personal files, cloned
stock and every R319..R335 seal including the separate R335 reference repair.
Original GAME.EXE execution, personal writes and golden changes remain absent.

Next are Claude and the remaining workers. Return counter 2 does not complete
the five-worker mine or three-class stories/endings. A read-only original
book handler also confirms its unlearned Summon gate; no forced spell/guide
or quickbar binding is supplied. Warrior zipper/Horrendous, broader UI/audio,
original Windows-peer, reachable ABI/native-generator and stability remain
open. Weighted maturity remains 79%, not measured coverage; all seven areas
remain partial and the overall 100% goal remains active.

### Retained R334 findings

R334 makes the campaign input harness aim through the actual viewport and
original mouse/quickbar controls. Earlier fixed pixels missed moving creatures
in ordinary-clock runs. The new cast-at-nearest-monster-input action copies
an existing hostile creature's identity, derives its current screen position
from the client sprite/viewport/Z, and queues mouse motion plus one original
A/S/D/F/G press/release. Bounded native-frame tracking follows camera and
sprite movement; conflicting native cursor identities block the key. It never
assigns CursorObj, spell targets, ownership, AI, health, position, velocity,
guides, spells or script variables, and it never calls a direct cast function.
When the original Charm makes a target non-hostile, the input diagnostic may
report observed_absent; independent duration/ownership observations distinguish
that transition from an actor physically disappearing.

Nine new regressions cover original key events and release, native cursor
conflicts, projection changes, invalid range/death/disappearance, native frame
pause/wrap, bounded reads, retained references and reader/input-only runtime
calls. All 58 focused and 58 strict tests pass, with actual cgocheck2, race and
checkptr=2. Four fresh full suites retain every prior terminal identity,
action and multiplicity, adding exactly nine passing tests per lane.
Normal/highres each have 82,108 PASS/49 SKIP; server/server-highres each have
77,280 PASS/7 SKIP, zero FAIL. All 6,645 tested source hashes match clean commit
b6f0d60ae9eb5de85a5f6c584fefb161c3f2831a. Four fresh Go1.26.9 ARM64 Mach-O
products exit 0 for help, with matching independent Go/REA metadata and clean
VCS state. The 97,044,480-byte full git archive contains 7,014 independently
rehashed members. All gates preceded the first R334 game; unchanged tests and
products were reused for later input routes in this round.

The first real A now creates an original Charm duration targeting adult
AlbinoSpider2132/script1000173065, then a living owned MONITOR/MIGRATE adult.
Virtual, ordinary normal and ordinary highres all independently reproduce it.
The ordinary first-A diagnostics take 48.525/48.503 seconds, finish 0 without
PlayerDeath, and retain health 15/poison 0/living Naldo. These intentionally
write no Save and are not new progression checkpoints. The known room roads
then escape with Naldo and control a real Imp; native mouse aiming also binds
the north small spider. Route 5 writes a living native health 11/poison 0 Save,
but both pets are dead in its later observation. It is a dangerous living
checkpoint, not proof that two pets survive that route.

Actual original staff defense plus adult1747 Charm clears the western small
spiders. Route 12 saves health 9/poison 0/mana 38, XP 3819.692, with living Naldo,
adult1747 and Imp1000220776 before and after Save. Fresh menu Load 13 restores
both creatures as living direct host children with MIGRATE. MONITOR/USEABLE
change after Load; the existing pet reload contract explicitly does not
require MONITOR alone. This is not an established ownership or AI defect.
Stationary Charm waits, jumps alone, and one outer-range approach still die;
all six PlayerDeath branches remain rejected and never seed continuation.

At the failed outer approach, actual player414,2811 is308.8 units from the
bend Imp644,3017, outside original Charm's 300 range; no A was queued. Moving
physically to440,2840 brings it in range. Original A then targets actual
Imp1000220778, staff input defends against nearby small spiders, and the
original roads reach Worker3Return. Map-specific script observations show
state15 3->4, state17 1->2 and returned-worker counter21 0->1; Naldo moves to
the original waiting area and remains alive. The original Worker3Return
function contains the 250 XP award. Total XP changes also include combat and
must not be attributed entirely to the return. The mine rescue journal stays
active; full mine rescue and later objectives are not complete.

Virtual route 14, ordinary normal 15 and ordinary highres 16 each independently make
that original return and finish 0 with a current owned manual Save and living
Naldo after Save. The ordinary segments take 87.510/87.961 seconds and begin
from the same virtual western Save 12, not fresh ordinary characters. Native
saved health is 6/6/4 respectively, poison 0, mana 78 and gold 513. Their saved
XP is 4220.6943/4311.9873/4221.463. Independent native parsing checks positive
health, owned current path, active con03b map, new SHA and the archived map.
Reload of a completed return Save, passive recovery, other workers and later
mine objectives are the next work; the observed return is not a whole-story
or full ordinary-character completion claim.

All 15 executed captures settle: 9 exit 0, 6 negative exit 2, 186 measured physical
arrival gates and 5 newly valid living manual Saves. One prepared case is
explicitly NOT RUN because its planned 13 map observations exceed the wrapper's
12-observation budget; a new immutable route removes redundant observations.
All owned process groups are verified/quiescent, without forced termination,
capture timeout or truncation. Six optional final script reads race orderly
quit; missing reads are retained and no return proof depends on them. Observed
unsupported-native counts are 0. Preservation rehashes 6,645 source files, 3,283
prior protected files, 1,562 personal files, all cloned stock, and every R319
through R333 sealed artifact. Original GAME.EXE execution, personal writes,
golden changes and forced game-state mutation remain absent.

Full three-class stories/endings, remaining worker and side objectives,
Warrior zipper/Horrendous, broader UI/audio, original Windows-peer, reachable
ABI/native-generator and sustained stability evidence remain open. Weighted
maturity stays 79%, not measured coverage; all seven areas remain partial.

### Retained R333 findings

R333 adds read-only game-loop observations for actual player targeting.
The new observe-player-target action copies the original mouse position,
client cursor/drawable, server CursorObj, current weapon, health/poison/mana,
server/client guide arrays and the real player's active Charm duration target.
It never runs target selection, changes input, casts a spell, learns a guide
or writes world/player/spell records. Output retains no object pointers;
eight copied Charm rows and a bounded1024-record scan prevent an invalid
duration list from hanging the observer. Nine new regressions cover copied
values, no mutation, absent records, caster/spell filtering, bounded output,
cycles, reference retention and the runtime's reader-only calls/assignments.

Both49 focused and49 strict tests pass, including actual cgocheck2, race and
checkptr=2. All four fresh full suites preserve every prior terminal identity,
action and multiplicity, with exactly nine new passing events per lane.
Normal/highres have82099 PASS/49 SKIP each; server/server-highres have77271
PASS/7 SKIP each, zero FAIL. Clean code commit
b4732c9c09f51ce8fdb18fd027cf4efdad25f4e2 produces four new Go1.26.9 ARM64
Mach-O products. Each help exits0 and independent Go/REA metadata agrees on
revision, tags and unmodified VCS state. All6643 tested source hashes match
the clean commit. Its97024000-byte full git source archive has7012 members
independently rehashed. Existing protected tests/goldens remain unchanged.

Fresh original-menu Loads of the living R33221 Naldo Save show matching
server/client guide IDs7,8,13,24,25,26 and a real equipped StaffWooden pointer.
The original table names these AlbinoSpider, SmallAlbinoSpider, Imp, Spider,
SmallSpider and SpittingSpider. Charm level remains1. Poison stays0 in the
captured native observations. These findings do not establish a guide,
equipment, max-health or saved-coordinate restoration defect. The weapon
pointer clears after an observed original player death; that is not a Load
failure. Native observations distinguish that transition from a living Save.

The first side-room cast receives the original BeastScroll-required response.
Before that cast, Troll7199 is94.9 units away and adult spider2132 is129.2
away. The native cursor later actually selects2132. Inserting a second real
A at that observed point creates a genuine Charm duration record targeting
original2132/script1000173065, then a living MONITOR/MIGRATE host-child adult
spider. No object or cursor target is assigned by the harness. This proves
adult control in the virtual runs. The later room defense/exit attempts still
record PlayerDeath and are not continuation sources.

The original Hunt menu followed by uninterrupted known room roads escapes
through the door and reaches985,1425 with player health5 and living Naldo.
That intentional no-Save diagnostic finishes0; the adult pet has died.
Its final world observation includes an incoming ImpShot. Moving to1020,1470
in the next attempt preserves health5 initially, but the following Imp cast
receives the original no-creature/pointer message. Subsequent damage and
PlayerDeath reject that branch. No safe Naldo hub return, worker-return
increment or rescue completion is established.

Ordinary normal and highres each finish their separate approximately51-second
no-Save target diagnostics0, without PlayerDeath. Their native guide and
weapon observations pass. The fixed-coordinate second cast does not reproduce
adult control: both casts receive the original BeastScroll-required response,
the actual cursor does not select the adult, and no Charm record/pet exists.
The same saved start does not guarantee identical later NPC positions. These
negative results are retained rather than credited as ordinary Charm success.
Future input should follow actual sprite/cursor observations through original
mouse events, without target setters, instead of relying on this fixed timing.

All10 captures settle:4 exit0 diagnostics and6 bounded negative exit2, with38
measured physical arrival gates. One branch writes a native health7, poison0
owned Save but subsequently records PlayerDeath; it is explicitly rejected.
There are no newly usable living manual Saves. The preferred story source
remains sealed R33221, health15/poison0/Naldo alive. All groups are verified
and quiescent, without forced termination, capture timeout or truncation;
observed unsupported-native-branch counts are0. Preservation rehashes6643
current source files,3282 prior protected files,1562 personal files, cloned
stock and every R319 through R332 sealed artifact. The initial comparison
helper syntax error and an overly broad post-death equipment certificate
assertion are retained separately from game/test failures and corrected in
new helper versions; neither caused suite or game reruns.

Full three-class stories/endings, living worker returns and later mine
objectives, Warrior zipper/Horrendous, side objectives, broader UI/audio,
original Windows-peer, reachable ABI/native-generator and stability evidence
remain incomplete. Weighted maturity remains79%, not measured coverage;
all seven areas remain partial. Original GAME.EXE execution, personal writes,
golden changes and forced game-state mutation remain absent.

### Retained R332 findings

R332 verifies original elevator2 player transport, Alex's original route
and dialogue, further native creature combat, and Naldo recognition through
the original Ruby door. Code and products remain unchanged at
3127d8d938d1bf2e453475b88a74943d96e93b0d. Pure elevator observations identify
Elevator305 paired with pit5382: actual boarding reaches the lower platform,
the player naturally leaves at Z0, and original BobDown enables Alex's next
movement. Following five physical road goals and using the real Talk/Done
controls advances Alex's original state to3. Virtual normal and ordinary
normal/highres all write native health19, poison0, mana78, gold513 and
XP2976.649 Saves. The ordinary segments take about94.12/94.15 seconds and
start from the same earlier virtual first-worker Save; they are not fresh
ordinary characters.

On the north road, actual ranged Charm controls HuntImp3. The two small enemy
spiders, adult spider and enemy Imp are subsequently observed DEAD. The
controlled small spider and newly controlled Imp also die; only the older
Imp survives that saved branch. Fresh Load and the real Escort menu bring
the surviving Imp back to the player. The apparent saved pet-coordinate
mismatch was a comparison against the later post-Save snapshot: archived map
coordinates match the actual pre-Save position, and the next Load restores
that position. No save-coordinate repair is justified. Actual running and
Space then bypass a Troll and defeat another Troll and small spider through
natural combat; the last Imp dies. These outcomes do not establish a broad
combat or pathfinding fix.

Normal passive recovery restores health11 to20 before the next approach.
Actual Charm controls original SmallAlbinoSpider2129. One attempted return
branch writes a native health20 Save but records PlayerDeath during the
post-Save observation; the entire exit2 branch is rejected. A healthy branch
instead physically reaches the original RubyDoor, consumes RubyKey2582,
crosses Worker3Rec, earns100XP and opens Naldo's native dialogue. The real
Done control grants replacement RubyKey4835 and original CreatureFollow.
Case21 writes a living health15, poison0, mana71, gold513, XP3574.8599 Save;
Naldo is alive both before and after Save. The earlier controlled spider has
already died. Original logical global17=1 and key-grant global22=2, while
worker-return global21 remains0. Recognition and following are not a
completed worker return or the six-worker objective.

An actual east-side room route reaches the door, and the original road node
871,1221 permits the outward crossing. Naldo follows out and along the north
cave. The recognition script naturally activates/spawns additional enemies;
later road attempts still record PlayerDeath and provide no continuation
Save. A ranged cast on Small2128 succeeds during one rejected return branch.
The nominal two-pet/adult attempt actually controls only Small2128, then
receives the original BeastScroll-required response; the pet dies before
Save. Lower cursor aiming also fails to establish adult ownership. No adult
Charm, safe hub return, enemy clearance or two-pet success is credited.
Native guide names already include AlbinoSpider and SmallAlbinoSpider, but
the actual rejected spell target is not proven. Original quickbar slot0=9
passes. Inventory clicks do not prove an equipment flag transition or an
equipment restoration defect. Existing max-health20 is consistent with the
untouched level/balance evidence; no health repair is made.

One room diagnostic finishes with a living player and a new native health15
Save but Naldo is DEAD in the final world observation. It is explicitly
excluded as a progression source. This adds a story-NPC survival check to
the usual inner exit0/no-force/no-PlayerDeath/new owned native Save checks.
Case21 remains the preferred story checkpoint; case4 is the earlier safe
Alex checkpoint. Low-health diagnostic Saves are retained without claiming
protection or quest completion.

All32 executed captures settle:23 exit0 and9 bounded negative exit2. There
are14 verified living-player manual Saves, of which13 also retain living
Naldo, and138 measured physical arrival gates. One prepared alternative is
NOT RUN and one invalid-action preparation is rejected before game launch.
No owned process is forced, timed out or truncated; every group is verified
and quiescent. Native health/poison/current owned Path/Map/new SHA and the
archived map are independently checked. Optional missing final reads prove
no completion. All observed unsupported-native-branch counts remain0.

Sealed R32940 focused/40 strict checks, all four full suites, matched clean
ARM64 products and source archive are reused because source is unchanged.
Normal/highres retain82090 PASS/49 SKIP each; server/server-highres retain
77262 PASS/7 SKIP each, zero FAIL, with prior terminal identities/actions/
multiplicities preserved. Preservation rehashes6641 source inputs,3282
protected files,1562 personal files, cloned stock and all R319 through R331
sealed artifacts. Original camera/lookup/parser/input negatives remain
retained. Full three-class stories/endings, worker returns and other mine
workers, Warrior zipper/Horrendous, side objectives, broader UI/audio,
original Windows-peer, reachable ABI/native-generator and stability evidence
remain incomplete. Weighted maturity stays79%, not measured coverage; all
seven areas remain partial.

### Retained R331 findings

R331 reaches the first original mine worker through the stock elevator,
learned Charm, controlled-creature combat and original door/recognition
triggers. No source change is required. A fresh process first Loads the spent
GoldKey Save with one SapphireKey, gold 513 and Charm9. Actual road movement and
the original lower button activate elevator1. The native Use cursor requires
less than 75 world units; attempts from the centered boarding position are
outside that range. Approaching and using the button before boarding works.
Pure game-loop observations show original lower-to-upper transport and the
return trip. Waiting 24 ticks and actually leaving the upper platform writes
living health 20 Saves in virtual normal and ordinary normal/highres. Both
ordinary cases start from the prior virtual mine-door Save.

A rushed upper-road diagnostic reaches six goals and then records PlayerDeath;
it has no new Save and is never a continuation seed. Ranged Charm instead
controls original SmallAlbinoSpider5685. Retreating before the channel finishes
cancels an earlier cast for exceeding the original range. That attempt still
writes a living Save, but no controlled creature is credited to it. Holding
range through the channel succeeds and produces a new living health 20 Save.
A new process Loads the same live owned spider with MIGRATE; MONITOR clears
after Load. Actual Charm on original Imp5485 then succeeds while holding range,
and both live pets are retained in another health 20 Save.

The actual all-creature Hunt command makes the pets roam away; its no-Save
diagnostic ends alive at health 6 without clearing the enemies. The original
Escort menu entry instead keeps both pets nearby. Both small enemy spiders
are observed DEAD while player health remains 20. Continuing the original road
lets the pets defeat the adult AlbinoSpider and writes a living health 18,
poison 0 Save. No order, AI, ownership, team or health setter is involved.

From that virtual road Save, actual movement defeats the original Bat,
consumes SapphireKey at InitialDoor and crosses Worker1Rec. The stock script
prints first worker rescued/five remaining, awards 250 XP and opens Alex's
dialogue. Clicking the real Done control grants RubyKey2582. Alex then enables
the next lift and reaches the original waiting point. This complete
road/door/recognition/Done/Save segment also passes ordinary normal/highres;
all three start from the same prior virtual road Save. Native manual Saves
record health 17/16/17 respectively, poison 0, gold 513, mana 70 and XP 2976.649.
The rescue journal remains active; the worker return count is 0. Five more
workers, return conditions, later objectives and endings remain unfinished.

All 20 captures finish: 18 exit0 and 2 bounded negative exit2, including the
retained rushed-road death. Ten new living manual Saves independently pass
native health/current owned Path/Map/changed SHA/archived-map checks. There
are 74 observed arrival gates. Every owned group verifies and quiesces, with no
forced termination, capture timeout or truncation. Eight intentional no-Save
diagnostics remain separate from saved checkpoints. Missing optional final
reads establish no completion. Script global symbols use the untouched map's
variable offsets so array globals are read from the correct native cells.

Code and all four clean ARM64 products remain
3127d8d938d1bf2e453475b88a74943d96e93b0d. Sealed R329 40 focused/40 strict checks,
four full suites and matched products are reused while source is unchanged.
Normal/highres retain 82,090 PASS/49 SKIP each; server/server-highres retain
77,262 PASS/7 SKIP each, zero FAIL, with every prior terminal identity/action/
multiplicity preserved. Preservation rehashes 6,641 source inputs, 3,282 protected
files, 1,562 personal files, cloned stock and every R319 through R330 artifact
seal. Original camera-name mismatches, lookup/parser warnings, broader input,
audio, original Windows-peer, native-generator/ABI and stability evidence
remain open. Full three-class campaigns, Warrior zipper/Horrendous and side
objectives remain incomplete. Weighted maturity stays 79%; all seven areas
remain partial.

### Retained R330 findings

R330 continues the original Conjurer story from the earned Mayor reward
Save. Actual movement through the enabled town-gate triggers starts the
original Imp scene and brings Aldwyn into the conversation. Actual Done adds
LocateMineForeman2 and unlocks the town gates. A new living manual Save
independently records health 20, poison 0 and gold 513. No NPC, script, camera,
lock or player-position setter is used.

Actual gate movement then triggers EndMission and the original exit into
Con03a. The real chapter briefing is dismissed with mouse input and a living
manual Save is written. This transition passes virtual normal and ordinary
normal/highres; both ordinary cases start from a prior virtual living Save.
A fresh process Loads the ordinary-produced Con03a Save with Charm9 and
gold 513. Approaching Janero before native-cursor Talk makes the conversation
work, followed by four actual road arrivals and a living Save. A direct Talk
attempt before approaching remains a bounded negative.

Seven further original road goals reach the town hub, and another nine reach
the mine entrance. The Millard cursor-selection attempt remains unsuccessful;
no dialogue or new Save is credited to that case. Actual southwest movement
into the stock exit instead enters Con03B and produces a living manual Save.
A new process Loads it and follows the original road to Foreman. Talk and
Done complete LocateMineForeman4, add Con03RescueMineWorkers2 and give the
original GoldKey, SapphireKey and Imp field guide. The first attempt leaves
the automatically opened knowledge book active: Escape closes the book,
so its subsequent menu clicks never write. Its old byte-identical PLR is
explicitly rejected as a new reward checkpoint. An additional real Escape
closes the book before the Save menu and produces a new living Save.

The complete Foreman dialogue, two key checks, book close and manual Save
also pass ordinary normal/highres from the prior virtual Con03B entry Save.
All three independent native PLRs record health 20, poison 0, gold 513,
XP 2443.4712, the completed Foreman objective, the new rescue objective and
Imp among the six named field guides. Section 8 is read independently using
the original version 1/presence/count/name format; no guide setter or fixture
is involved. Older chapter journals are no longer present in these new
chapter Saves. A fresh Load retains both keys, Charm and the new objectives.

Actual right-mouse movement through the original gold-locked mine doors
consumes GoldKey from 1 to 0 on both server and client, retains one SapphireKey
and reaches the far side. A new manual Save records health 20 and gold 513.
The six workers, later mine objectives and complete story remain unfinished.

All 17 owned captures finish with exit code 0 in 14 cases and exit code 2 in
3 bounded negative cases.
Twelve living manual Saves independently pass native health, current owned
Path, Map, changed SHA and archived-map checks. There are
69 observed arrival gates. Every owned group verifies and quiesces, with no
forced termination, capture timeout, truncation or PlayerDeath. The intentional
no-Save diagnostic and unwritten reward attempt remain separate from the
living checkpoints. Optional missing final reads establish no completion.

The original Mayor camera warning is further classified from untouched data:
waypoint 398 is named Con02a.map:RewardLookWP while the original script requests
MayorRewardLookWP. No waypoint alias or camera correction is invented. New
chapter lookup warnings and the Millard input negative remain retained;
progression evidence does not establish full camera or NPC interaction fidelity.

Code and all four clean ARM64 products remain
3127d8d938d1bf2e453475b88a74943d96e93b0d. The sealed R329 40 focused and 40 strict
checks, four full suites and independently matched products are reused without
rerunning unchanged code. Normal/highres retain 82,090 PASS / 49 SKIP each;
server/server-highres retain 77,262 PASS / 7 SKIP each, with zero FAIL and all
prior terminal identities/actions/multiplicities preserved. Preservation
rehashes 6,641 source inputs, 3,282 protected files, 1,562 personal files, cloned
stock and every R319 through R329 artifact seal. No original GAME.EXE,
personal write or golden approval occurs. Full three-class campaigns/endings,
the remaining Warrior gauntlet/Horrendous and broad gameplay/UI/audio/original
Windows-peer/native-generator/stability evidence remain incomplete. The weighted
estimate stays 79%; all seven areas remain partial.

### Retained R329 findings

R329 fixes a naturally reached NPC damage branch. An Urchin's original
ThrowingStone reaches PlayerDamage as a distinct SIMPLE|MISSILE projectile
with an initialized monster source and IMPACT11 damage. The bounded native
admission now uses the existing full armor, fractional carry, item wear and
default damage path, matching the original case 11. Other adjacent missile
shapes retain their prior behavior. Real C-owned records above 4 GiB exercise
the original dispatcher through native HP changes; additional server checks
verify carry, attribution and callback order.

The old and new captures use byte-identical scenarios and the same living
Save. The old 396 unsupported NPC stone events become zero. Two attacking
Urchins are observed DEAD while town NPCs remain alive; this is not credited
as player combat. The fixed path writes a living health 20 Save in virtual
normal and ordinary normal/highres, with zero unsupported native events.

Actual learned Charm quickbar input then controls both original quest
spiders. The first conversion causes the original script to create the second
spider. Selecting and casting on that second spider before crossing the office
doors preserves line of sight. Both live spiders become host children with
MONITOR and MIGRATE, and a new living manual Save records health20. A fresh
process restores the same two script IDs as live host children with MIGRATE;
MONITOR is absent after Load, so MONITOR persistence is not claimed.

Owning both spiders alone leaves the quest incomplete. Clicking the original
all-creature cage menu and its first Banish entry removes them and lets the
original SpiderEngine set its completion flag. The actual Mayor free response
and Done unlock his room and start his walk. A second actual conversation
awards the original 500 gold and completes BanishSpiders4. Independent native
manual Saves from virtual normal and ordinary normal/highres each record
health 20, poison 0, gold 513, Charm1 and XP 2443.4712. The ordinary cases execute
both Charms, Banish and both Mayor conversations from the prior virtual living
Save; they are not fresh ordinary characters. Boots and later objectives remain
open.

The original Mayor free-response camera lookup still logs two unresolved
MayorRewardLookWP warnings in each full path. Reward and living-save evidence
does not establish correct camera framing; that lookup remains a separate
open issue.

All 14 owned captures finish with 9 exits0 and 5 bounded negative exits2.
There are 8 independently verified living manual Saves and
50 actual arrival gates. Every owned group verifies and quiesces, with no
forced termination, capture timeout or truncation. The earlier Save followed
by PlayerDeath remains rejected as a seed. The wrong-room approaches, a close
range death, an incomplete Mayor reward attempt and intentional diagnostics
remain retained. Missing optional final reads support no completion claim.

Code and the four fresh clean ARM64 products are
3127d8d938d1bf2e453475b88a74943d96e93b0d. Fresh focused and strict checks each
pass 40 terminal events, with race/checkptr2 and actual GOEXPERIMENT=cgocheck2
enabled for strict. Four fresh full suites pass: normal/highres 82,090 PASS / 49
SKIP each; server/server-highres 77,262 PASS / 7 SKIP each; zero FAIL. Every prior
terminal identity, action and multiplicity remains, adding exactly 12 new
terminal events. Product metadata is independently matched and the source
archive is rehashed. Preservation rehashes 6641 frozen source inputs, 3282
protected files, 1562 personal files, cloned stock and every R319 through R328
artifact seal. No original GAME.EXE execution, personal write or golden
approval occurs. Full three-class campaigns/endings, Warrior zipper/Horrendous
and the broader gameplay/UI/audio/original-peer/native-generator/stability
matrix remain incomplete. The weighted estimate stays 79%; all seven areas
remain partial.

### Retained R327 findings

R327 earns original gold and completes Aldwyn's first lesson through actual
mouse/keyboard input. Clicking original forest StumpChest1 extent6604 opens
it and drops its native43-gold item; clicking that visible item collects43.
A changed, owned manual Save independently records health20 and gold43.
Walking near the dropped item alone remains a negative attempt. Original
script inspection corrects the earlier fee candidate: Aldwyn compares and
deducts global59=30. The inaccessible town Gold500 is referenced as the
Mayor's reward by global89; unsuccessful counter approaches do not establish
a pickup defect or legitimate possession of that reward.

The earned-gold Save Loads in virtual normal and ordinary normal/highres.
Five actual exterior/house arrivals, native-cursor Talk, Yes and four actual
Next/Done clicks finish the lesson and write new living manual Saves in all
three cases. Each native PLR records health20, gold13, SPELL_CHARM level1,
XP943.4713, LocateAldwyn4 and the new BanishSpiders2 entry; the boots entry
remains incomplete. Both ordinary cases start from the virtual earned Save,
not fresh ordinary characters. Original lesson scripts enable elevators205
and202. One earlier exit0 attempt reaches the runtime lesson but leaves its
last dialog open: its Save attempt never writes, and its old gold43/spells-null
PLR is explicitly excluded as a new lesson checkpoint.

A new process Loads the lesson Save with gold13 and Charm9 in quickbar slot0.
Passive game-loop observations show original pit204 boarding carry the player
to cave elevator205 at3967,460, then return if the player stays aboard.
Waiting48 ticks and actually running to3888,548 instead disembarks and writes
a living health20 Save. Its later live observation is health12 without death.
Eighteen original cave goals then reach the enabled MagicShop elevator202.
A slow final approach lets pursuing monsters also board and displace the
player before the rise. Running the final approach and then making a short
center walk retains boarding and observes original transport into town.
Four further actual shop/town exit goals produce a new manual Save at native
health15, poison0, gold13 and Charm1. No lift, AI, position or health setter is
used. Actual Charm casting, pet ownership and BanishSpiders completion remain
unverified.

Actual run/Space observations also cross the original tavern counter after
establishing facing before the jump. State12/AIRBORNE and the observed
trajectory establish the jump even though native ZVal remains0. No jump-height
repair or Gold500 possession is claimed. A cave attempt reaches five goals
and writes native health20, but original PlayerDeath follows Save; that entire
exit2 branch is retained and rejected as a continuation seed.

Thirty-one owned captures finish17 exits0 and14 bounded negative exits2, with
no forced termination, capture timeout or truncation. Every owned group
verifies and quiesces. Twelve living original manual Saves pass independent
native health/Map/current Path/changed SHA/archived-map checks. Four exit0
diagnostics intentionally make no new Save; the other exit0 without a new
Save is the incomplete-dialog attempt described above. All31 prepared
scenarios pass actual Go-AST action registration, and111 observed arrivals
retain both server/client distances at most twelve units. Fifteen optional
final object/script reads race shutdown; missing reads support no completion
claim. All failed inputs and the rejected post-Save death remain preserved.

Source and the four clean ARM64 products remain
632fb3a4f8e330a820b965719638f2a87c7b8faf. The sealed R326 four-configuration
suites,14 normal input/observation checks and14 race/checkptr2/actual-cgocheck2
checks are reused without rerunning unchanged code. Normal/highres retain
82,078 PASS49 SKIP each; server/server-highres retain77,250 PASS7 SKIP each,
with zero FAIL. All prior terminal identities/actions/multiplicities remain.
Preservation rehashes6639 frozen source inputs,3282 protected files,1562
personal files, cloned stock and every R319 through R326 artifact seal.
No original GAME.EXE execution, personal write or golden approval occurs.
Full three-class campaigns/endings, later Warrior gauntlet/Horrendous and the
broader gameplay/UI/audio/original-peer/native-generator/stability matrix
remain incomplete. The weighted estimate stays79%; all seven areas are partial.

### Retained R326 findings

R326 confirms original GreenElevator200 transport to GreenElevatorPit201
and an actual forest disembark, followed by a living manual Save. The same
path passes virtual normal and ordinary normal/highres. Native saved health
is10/20,15/20 and14/20 respectively, with poison0 and XP743.4713. The two
ordinary cases start from a prior virtual natural menu Save, not an ordinary
fresh campaign. A fifteen-sample passive trace observes the player rise with
the platform, move to3599,1610, remain aboard and later return to2771.5,4473.5.
The disembark cases leave during the normal raised phase. Earlier sparse
endpoint observations did not establish whether an intermediate ride occurred.
No elevator physics repair or forced activation is used.

A new process Loads the forest Save at native health10. Original passive
recovery over5400 ticks restores health20. Five actual linked forest waypoints
then reach Aldwyn's house and a new manual Save. Original native-cursor Talk,
Yes and Done produce the rendered insufficient-money response, close the
dialog and save alive. An independent version3 Inventory-section reader
establishes saved gold0. The previously unverified candidate85 is the native
status Direction1 field, not money. Player spells remain null and the two
journal entries remain LocateAldwyn2 and ReturnBridgeGuardsBoots2; no fee
payment, player spell learning or objective completion is claimed.

Actual center walking also stands ON_OBJECT at CavePit204. Fifteen passive
observations retain reciprocal native links while linked CaveElevator205
stays disabled at state2/height64 and the player stays health20. This is an
original script-state gate, not evidence of a broken native link. The trace
quits normally without a new Save. The boots side-room gate actually reports
that RubyKey is required. Read-only world inventory identifies original
RubyKey6621 inside ChestUrchin4 extent6620 at3835,5374. Attempts to approach
that chest, cross the river or enter the locked room remain negative. Two
failed combat branches each observe one original Urchin DEAD, but neither
produces a living saved combat result. No key, boots, opened chest or new gold
is claimed. One attempted menu Save is followed by original PlayerDeath and
is explicitly excluded as a continuation seed.

Sixteen owned captures finish with eight exits0 and eight bounded exits2.
All owned groups verify and quiesce without forced termination or truncation.
Six new living manual Saves independently pass native Map/Path/health/SHA
checks; one is a partial house-approach checkpoint with no chest completion.
Two exit0 diagnostics intentionally make no new manual Save. One prepared
combat route is superseded and remains NOT RUN. All seventeen prepared
scenarios pass the real Go-AST registered-action verifier;68 observed arrival
gates retain both server/client distances at most twelve units. Legacy directional
walk/run actions target a fixed1024x768 screen center; a world-direction label
is not general camera-edge evidence. Nearest-monster inputs can approach a
target up to512 units away and leave the intended room. Those limits and
all unsuccessful original inputs remain preserved.

Source632fb3a4f8e330a820b965719638f2a87c7b8faf adds only a passive game-loop
elevator observation and three mutation/lifetime tests, with no gameplay
mechanics change. Every prior R323 terminal test identity/action/multiplicity
is retained. Fresh full suites pass82,078 events/49 skips in normal and highres,
and77,250 events/seven skips in server and server/highres, with zero failures.
Fourteen combined input/observation tests pass normally and with race,
checkptr2 and actual GOEXPERIMENT=cgocheck2. All four new clean Go1.26.9 ARM64
products have matching Go/REA revision, tags, architecture and unmodified VCS
metadata, with help exit0. Preservation checks cover6639 frozen source inputs,
3281 existing protected files,1562 personal files, cloned stock and all sealed
R319 through R325 artifacts. No original GAME.EXE execution, personal write or
golden approval occurs. Full campaigns/endings and the broader gameplay/UI/
audio/original-peer/native-generator/stability matrix remain incomplete. The
weighted maturity estimate stays79%; all seven areas remain partial.

### Retained R325 findings

R325 verifies the Warrior's fourth-arena exit, the following hardened-ground
lava bends and the unlocked ZipperDoor entrance. The perpendicular gate route
2780,3700 to2860,3780 passes with both server/client arrival checks and a living
manual Save. A later new process restores the position beyond the gate. The
first eastward lava bend reaches its requested points but dies after Save;
walking along the inner bends instead retains HP20/20 and poison0. The complete
exit/lava/door route then passes ordinary normal and highres with HP20/20,
XP2875.7322, poison0 and the original zipper activation global238 at1. Both
ordinary cases start from an earlier virtual menu Save, not a fresh ordinary
chapter. The original Troll remains DEAD and its death counter1 persists.

Eighteen owned captures produce nine exits0 and nine bounded negative exits2,
without forced termination or truncation; all owned groups verify/quiesce.
Eight new living manual Saves independently pass native Map/Path/health/SHA
checks. Five negative cases write a menu Save before original PlayerDeath;
their changed owned PLRs are retained diagnostically and explicitly rejected
as continuation seeds. One exit0 is an intentional read-only native-link
diagnostic with no new Save. One superseded prepared route remains NOT RUN,
and one preparation fails before creating a route, cloning assets or launching
a game. All nineteen prepared scenarios pass the real Go-AST input verifier.
The preservation audit checks 81 observed server/client arrival gates at most
twelve units from their actual goals.

A new process Loads the ordinary R324 Troll victory Save at HP8/20. Original
passive recovery, without food or health overrides, produces a manual HP20/20
Save. The later zipper remains incomplete: a center run enters the corridor,
and the original moving StoneBlock subsequently pushes the player back without
health loss. Full-corridor goal waits remain negative, while a southwest-lane
attempt causes original PlayerDeath. Neither Space input nor reaching a
partial position establishes complete trap traversal or Horrendous completion.

The Conjurer immediately resumes actual movement after the existing living
Load checks, rather than waiting idle under fire. Original run/Space input
reaches the next two linked cave waypoints and manually saves HP17/20,
XP743.4713 and poison0; the later live observation is HP4/20. All twenty-four
whole-world Urchin rows retain their identities and remain alive; no cave kill
or player spell learning is claimed. Later attempts traverse nine more
original linked waypoints and approach GreenElevator200. Timed approaches and
center-target walking do not observe transport in their sparse checkpoints. One
observation shows player ON_OBJECT at the platform, then off it later; these
attempts die after Save and stay negative.

Native saved-map readers retain the reciprocal200/201 connection. The already
compiled, independently reviewed read-only `assert-elevator-xfer-loaded`
checks all four current elevators: four reciprocal native links, native-width
addresses and zero PE32 pointer slots. It passes without changing or saving
game state. This rules out a missing-link explanation at that observation;
the boarding-height/update interaction still needs diagnosis. No engine
defect or physical lift completion was established in that earlier round. The state-setting elevator
round-trip fixture is excluded from natural progression.

Normal/highres 82,075 PASS49 SKIP and server/server-highres 77,247 PASS7 SKIP
per configuration, zero FAIL, plus eleven normal/eleven strict input checks
remain the sealed R323 results. Source and all four clean Go1.26.9 ARM64
products remain `6128598105694a289f5a45cfa6127de878b59081`, so they are reused
without rerunning unchanged code. All 6,637 source inputs,3,281 protected files,
1,562 personal files, cloned stock files and R319 through R324 seals rehash
unchanged. Two optional final script reads race orderly quit and support no
completion claim. A flow-style preparation-prefix parsing issue is corrected
for fresh cases; the extra real movement in cases5/7 and all earlier scenario
hashes remain intact. No original GAME.EXE, golden approval or personal write
occurs. Full campaigns/endings and broader gameplay/UI/audio/networking/ABI/
stability evidence remain open. The estimate stays79%; all seven areas remain
partial and 100% is incomplete.

### Retained R324 findings

R324 continues the natural Warrior and Conjurer campaigns through original
keyboard/mouse input and menu Save/Load. Source and all four products remain at
`6128598105694a289f5a45cfa6127de878b59081`. The sealed R323 four-configuration
suites, eleven normal input tests and eleven race/checkptr2/actual-cgocheck2
tests are reused without rerunning unchanged code. Normal/highres retain
82,075 PASS and 49 SKIP each; server/server-highres retain 77,247 PASS and seven
SKIP each, with zero FAIL events.

Fifty-two owned captures comprise thirty-six game exits 0, fifteen bounded
negative exits 2 and one quarantined capture timeout. The fifty-one completed
games have no forced termination. In highres ordinary case39, original
PlayerDeath precedes a capture idle-timeout SIGTERM; no final game result or
successful new Save exists. That failed attempt remains immutable and excluded
from continuation seeds. All fifty-two owned groups verify and settle
quiescent, without truncation. Later affected ordinary captures allow the
existing 120-second native failure wait to finish by using a 150-second capture
idle bound; their game limits and assertions are unchanged.

Thirty-five new living manual Saves independently pass native Player.plr
health, active Map, current owned Path and changed-SHA checks. Exit-0 Conjurer
case4 enters dialog without updating the manual Save, so it is excluded. Two
superseded prepared cases remain NOT RUN; two missing-route preparations fail
before asset cloning or any game. All fifty-four prepared scenarios pass the
real Go-AST registered-input verifier. The audit validates 130 observed world
arrival gates with both server and client distance at most twelve units.

The Warrior traverses the original narrow lava route and the correct fourth
arena gate opening. Passive original recovery restores health10 to20. Actual
Sword input breaks the originally named potion crate, whose real contents are
RedApples, and the antidote crate. Pixel pickups independently produce two
apples and one CurePoisonPotion on server and client. Original lever use starts
the elevator/cage sequence and brings the same Troll extent12419 from its
offscreen position to the arena. Later snapshots show it DEAD and its original
death callback counter at1. Native saved Door2 and ZipperDoor lock codes are0;
actual traversal of the arena exit and the later gauntlet remains unverified.

Earlier targeting and southern-lane attempts remain negative or mixed,
including original PlayerDeath after a Troll kill. Living Saves with poison1
are explicitly incomplete treatment results. The original thing.bin defines
Troll drawable Z15; mouse offset y7 aims eight pixels above its standing ground
projection. With the unchanged R323 actual native-cursor action and a cleared
western lane, highres virtual and ordinary normal/highres runs defeat the
Troll, consume both real apples, move away from its poison cloud and use the
actual visible inventory antidote cell. Their manual Saves have HP9/20,8/20
and8/20 respectively, XP2875.7322 and active poison0. The ordinary runs begin
from a prior virtual menu Save, rather than an uninterrupted fresh chapter.
Separately, a new process Loads the earlier treated case33 Save, preserves the
Troll death/counter1 and poison0, moves and manually re-saves with HP18/20.

The Conjurer's original Necromancer sequence completes with four arrivals,
summon flag1, hit callback1 and dialog-done1. The original Spider callback
caller handle changes from0 to a positive script ID; its continued liveness
after the setpiece is not claimed. The Necromancer disappears through the
original sequence, player NO_UPDATE clears, and actual road movement and menu
Save succeed. The Henrick wolf/Charm NPC demonstration reaches its original
sixty pulses, releases the player, and likewise permits actual movement/Save.
Both sequences pass virtual normal and ordinary normal/highres cases. Native
player spells remain null: these NPC scenes do not prove player Charm learning
or pet ownership and AI-stack mastery.

Actual Bridge_Guard Talk/Done adds ReturnBridgeGuardsBoots2 beside
LocateAldwyn2. Walking onto the original GreenElevatorPit5942 physically
transports the player from the riverbank to its original GreenElevator5943 in
the den. An independent native-map reader confirms reciprocal extents and the
destination; actual transport and living manual Save pass virtual normal and
ordinary normal/highres. A later virtual run rests and traverses five cave
waypoints, saving native HP18/20, poison0 and unchanged XP743.4713. The later
live snapshot is HP8/20 after nearby-enemy damage; no den enemy kill is claimed.
The earlier den-entry Save is retained as the safer alternative. Aldwyn,
returning the boots, player spell learning and the complete chapter remain
open. Magic-shop approach/interior Saves do not establish a lift ride.

| R324 receipt | Scope |
| --- | --- |
| `ev_efe2a04c3730c34b820a47715bc7baa42e2576417878118fb54c713e4a89381e` | Ordinary normal Troll DEAD, real food/antidote use and poison0 living manual Save |
| `ev_8f439aac0b94439b3e27f4608320cc9450df050291c8dd0e0c4acd5b78559081` | Ordinary highres equivalent combat and treated living manual Save |
| `ev_3d4bc4bdba30c3baba745c55370a36865ce19b59f721c008f70d255b4d612fa3` | New-process Load preserves earlier Troll death/counter and treated player, movement/re-Save |
| `ev_f65e266b5a9500d8d647d402601ff909527ea0c536b01e9c598ba84272c32704` | Ordinary normal Necromancer sequence, released player and movement/Save |
| `ev_3ac412a9813cc707fb706a51dd713ac98b27dc91d495cd7eab8e3597321fba8d` | Ordinary highres Henrick NPC demonstration, released player and movement/Save |
| `ev_c51094d3b67ee5306b1f1dcdfbdc5fa0cee8e5732de5ea12e3d7b9692211e002` | Ordinary normal original green-pit transport and living manual Save |
| `ev_576527fa83fedd9cc309c81facf12dce8f2cbab1b3923b3f5382ccd8ea3521ad` | Ordinary highres original green-pit transport and living manual Save |

The final audit rehashes all 6,637 frozen source inputs, 3,281 protected files,
1,562 personal files, cloned stock inputs and the sealed
R319/R320/R320-runtime/R321/R322/R323 artifacts unchanged. Four final optional
script reads race orderly game shutdown; required earlier observations remain
available and no proof relies on those missing reads. No original GAME.EXE,
state-setting API, golden approval or personal-install write is used. Earlier
route labels are retained as raw evidence, with the actual angle convention
corrected in continuation notes:0 north,0.5 east,1 south,1.5 west. Full natural
three-class endings, later gauntlet, broader gameplay/UI/audio, original-peer,
native-generator and sustained stability remain incomplete. All seven areas
remain partial and the weighted estimate stays79%, not measured coverage.

### Retained R323 findings

R323 adds actual mouse targeting of the nearest original unnamed hostile monster
of an exact requested type. Each frame reads the current world, hostile
relationship, native line of sight, viewport and drawable. The bounded action
pins only wire/extent/script-ID values, approaches the ground position and
requires the exact original native cursor target before one ordinary left
press/release. Airborne art height affects aiming, while movement still follows
the ground. Death, disappearance and unavailable drawing remain distinct.
It invokes no damage, forced selection, script, use-packet or state-setting API.
The older named-target behavior is retained and passes its actual MasterUrchin
combat/menu-Save regression.

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,075 | 49 | 0 |
| Highres | 82,075 | 49 | 0 |
| Server | 77,247 | 7 | 0 |
| Server, highres | 77,247 | 7 | 0 |

Every R322 terminal identity, action and multiplicity remains, adding three
passing observation tests per configuration. Eleven combined input tests pass
normally and with race/checkptr2/actual-cgocheck2. Four clean ARM64 Go1.26.9
products at `6128598105694a289f5a45cfa6127de878b59081` have matching Go/REA
revision, tags, architecture and unmodified VCS metadata, with help exit 0.

Forty-eight completed owned game captures produce thirty-four exits 0 and
fourteen bounded negative exits 2. All owned groups verify and settle quiescent
without forced termination or truncation. Thirty-three cases have independently
verified living manual Saves. One exit-0 diagnostic naturally enters Con02a but
fails its manual-Save goal; process success does not qualify that goal. A prior
y-8 arena-two prototype stays unexecuted. Two separate preparations reject a
missing required manual-slot map before starting a game. The first also has a
failed CLI preflight. All forty-nine fully prepared scenarios pass the real
Go-AST action dispatcher and reviewed-input verifier. Negative, mixed and
unexecuted cases are retained, and none becomes a successful new seed.

The Conjurer genuinely defeats three original bats, two lower spiders and two
urchins through ordinary staff input. Whole-world snapshots independently show
each original actor alive and subsequently DEAD. Urchin y-8 aiming times out;
y14 accounts for its original height and permits the native cursor to select
it. Disappearance-only results are never counted as kills. Original menu Saves
and later new-process Loads preserve the deaths and progression.

Actual chest collision and a real dropped-key click produce SilverKey
server1/client1 and a living menu Save. A later original door consumes it,
observing server0/client0, and menu Save completes alive. Ordinary melee breaks
the original apple/key barrels. An actual RedApple click raises native health
from 8 to 13; the consumed food and living Save are verified. Ordinary in-game
rest also demonstrates natural recovery without a health override.

Actual ExitSwitch native Use changes the original switch enabled-to-disabled
and ExitElevator disabled-to-enabled, awards its original 100 XP and menu-saves.
Short platform walking and walking across a full cycle do not prove a ride.
Walking to the cycling platform center and stopping does: the original physical
carry reaches its paired shaft at 4015,2749 and menu Save succeeds. New processes
restore that position and traverse the exit cave. Passive existing script GET
snapshots observe the original collapse guards; original source variable offsets
are used when interpreting their values.

The original exit naturally loads Con02a. The first attempt opens the Save flow
during the transition fade: deferred AUTOSAVE/WORKING files have active map
con02a, while SAVE0001 remains the previous Con01a byte-for-byte. Two preparation
guards correctly reject it as a manual Con02a seed. The initial casing-only
explanation is retained as a rejected hypothesis. A fresh case waits for the
original fade, performs actual new-map movement and then completes the original
manual Save. Native FileInfo.Map and the archived map both identify con02a. A
new process uses the original Solo/row1/Load menu, moves and manually re-saves.
Both manual saves have HP8/20, XP743.4713 and poison0. The first new-map manual
save has no journal entries; the resumed/re-saved case has LocateAldwyn2. Full
Conjurer objectives and campaign completion remain open.

The Warrior crosses all three original crusher segments with actual run/Space
input, activates the first arena lever and defeats both original spiders. Real
in-game recovery raises health8 to20. Original cyclic flames remain active after
Save/Load, and an actual run/Space route crosses that section with HP20. The
second arena's original lever and both Urchin deaths pass in virtual normal,
ordinary normal and ordinary highres. Saved XP is2194.678, with health18,15 and14
respectively. Both ordinary cases start from a previous virtual actual-menu Save;
this is not an uninterrupted ordinary/highres fresh chapter.

Later unsafe routes and blocked cursor/wall approaches remain negative. The
unlocked inner gate and correct fire-hall lane eventually pass with HP18 and a
living menu Save. In the third arena, walking onto the original floor lever
already disables it and enables all three Urchins. A redundant native-Use wait
dies and remains negative. A fresh case immediately fights after actual contact:
all three original named Urchins are DEAD, and the original menu saves HP16/20,
XP2701.7498, poison0 and War2Gauntlet2. The fourth arena, later gauntlet hazards
and actual Horrendous completion/War2Gauntlet4 remain unverified.

| R323 receipt | Scope |
| --- | --- |
| `ev_043213500a872aec0b5b066bf3be3a6bce7aba89e8d88272ee52fc2f00094b3b` | Original unnamed Urchins DEAD and living manual Save |
| `ev_7be9484d679f3d48a08254e761ea70ba778dd12a2112fee2d286be89c48e26a8` | Exit-room combat, native ExitSwitch use and living manual Save |
| `ev_fb16f51baa3e80fb03c4f951356c29a75af80ac950085c278582251cfca8b14e` | Original physical exit lift carry and living manual Save |
| `ev_c0bdc135ad7d2e1b923a4e421af34028cd69902821dfa1808ffaadc9de323c6f` | Natural Con02a entry, settled fade, movement and verified manual Save |
| `ev_0aefb7f4ff8772eaffd16d53447dde7f1e141f95285b8467b3d17bdaaea86f7c` | New-process original Con02a manual Load, movement and re-Save |
| `ev_ead8d402c81d35219911f4fc52024806a016107d096c8178d02e3a28ce4f9275` | Actual floor-lever contact, three original Urchin deaths and living manual Save |

All 6,637 frozen source inputs, 3,281 protected files, 1,562 personal files and
R319/R320/R320-runtime/R321/R322 artifact seals retain their hashes. Every
verified manual Player.plr differs from its seed and names its current owned
case path. No original GAME.EXE execution or existing golden change occurs.
Full three-class natural campaigns/endings and the broader gameplay, UI/audio,
original-peer, native-generator and sustained-stability evidence remain open.
All seven areas remain partial; the estimate stays 79% and 100% is incomplete.

### Retained R322 findings

R322 adds actual mouse use of an original named map object. Each frame resolves
the current server object, its static/dynamic network code and current client
drawable. One real left press/release requires both that exact native cursor
target and CursorUse13. The action retains primitive observations between
frames and changes no cursor selection, object, script, network or world state.

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,072 | 49 | 0 |
| Highres | 82,072 | 49 | 0 |
| Server | 77,244 | 7 | 0 |
| Server, highres | 77,244 | 7 | 0 |

Every R320 terminal identity, action and multiplicity remains, adding two
passing named-use observation tests per configuration. The combined eight
NPC/monster/movement/use-input tests pass normally and with
race/checkptr2/actual-cgocheck2. Four clean ARM64 Go1.26.9 products at
`5a3af918e738e1b66896f905ef40608d7588b247` have matching Go/REA revision,
tags, architecture and unmodified VCS metadata, with help exits0.

Twenty-three fresh owned captures produce thirteen game exits0 and ten
negative exits2. Every owned group is verified and settles quiescent, without
forced termination or truncation. A private jump prototype used unsupported
`run-end` and fails before gameplay. Its artifacts remain unchanged. The
replacement verifier parses the real Go action dispatcher and rejects that
old prototype; every other capture uses registered, reviewed actual-input
or read-only observation actions. No failed save becomes a successful seed.

The original EntranceSwitch requires its raised art position, y-20. Exact
native Use input changes that switch from enabled to disabled and its original
EntranceElevator from disabled to enabled. Virtual normal, ordinary normal
and ordinary highres independently confirm both states and complete original
menu Save. All three save HP20/20, XP100, poison0 and LocateAldwyn2. Their
approach seed is an earlier virtual actual-menu Save; this does not establish
an uninterrupted ordinary/highres fresh chapter.

A new process loads the activated-lift Save, physically rides the original
elevator, passes eight upper-cave arrival checks and saves through the original
menu. Another actual route walks to the barrel-fire barrier, holds ordinary
mouse movement, presses Space and releases the button. It reaches the west
side alive, then menu-saves HP11/20, XP100, poison0 and LocateAldwyn2. Later
processes restore that saved position and continue. A genuine chest collision
releases its SilverKey; actual clicking observes server1/client1. Actual door
contact later consumes it, observing server0/client0. These later cases die
during further movement or after Save and remain negative overall: earned-key
persistence, exit-switch use and the next Conjurer chapter are still unverified.

The Warrior advances through the original corridor, wooden door, stairs and
next paired door. New processes continue to the gauntlet front and naturally
open the instructor dialogue. Actual Done input allows the original timed
redshirt/chicken demonstration; the original menu saves the living player
with HP15/20, XP1849.1724, poison0 and War2Gauntlet2. The actual gauntlet trial
and its completion remain open. Stair probes after earlier saves are explicitly
later than those saved checkpoints; no saved destination is inferred.

| R322 receipt | Scope |
| --- | --- |
| `ev_8713c3d6a911a1d25545b58340df7efcfe930b95f115f74e9a47c75315b727ae` | Ordinary normal native Use, elevator activation and menu Save |
| `ev_234a1bec1214371feaec3cb6de10d17b8beb1aed57772fe198f687d5f61235ca` | Ordinary highres native Use, elevator activation and menu Save |
| `ev_0a7ed867567d626c65f6efbed4b5fa5437d81eb1c5680614dde6954ea5ab732c` | Original physical lift, eight upper-cave arrivals and menu Save |
| `ev_3e2e91c113b6aa63b749b97feefeaf9a10d1f731bdea4946e75c23f3e635fce4` | Actual jump/fire crossing, living west-side arrival and menu Save |
| `ev_3afa7a1506b32573f5114a56f16fc5f0be61766137f8097547bd8a322211dc1b` | Original instructor Done, timed demonstration and living menu Save |

All 6,635 frozen source inputs, 3,281 protected files, 1,562 personal files and
R319/R320/R320-runtime/R321 artifact seals retain their hashes. Original
executable execution and existing golden changes remain zero. Read-only
snapshots observe unnamed bats, urchins and spiders near later failed routes;
they do not by themselves prove a cause of death. Next, actual native-drawable
targeting must support those original unnamed monsters. Full campaigns,
UI/audio/original-peer/native-generator/stability gates remain open. The
estimate stays79%; 100% remains incomplete.

### Retained R321 findings

R321 runs the unchanged clean R320 products through 14 fresh owned captures.
Ten games exit 0 and four stop on bounded E2E expectations with exit 2; every
owned group settles quiescent without forced termination or truncation. The
exit-0 distant-button probe supplies save/load evidence but fails to activate
the elevator, so process success does not qualify that gameplay result.

The earned Gearhart recommendation restores through original Solo/row1/Load
in a new process. Actual QuarterMaster input then loads War02a, the original
briefing closes through a real click, and the original menu saves chapter two.
Separate ordinary normal and ordinary highres cases complete all 18 return
segments, rewards, chapter entry and menu Save. Both start from the previous
virtual combat save; ordinary/highres combat is still unverified. New processes
load the resulting virtual and ordinary normal/highres chapter saves and pass
actual movement checks. Saved XP remains 1849.1724; new chapter journals are
initially empty until their own story events add an objective.

Actual movement into the original gauntlet trigger opens Dant's dialogue.
Real Done input continues to Horrendous, whose own Done awards War2Gauntlet2.
The original menu preserves that objective with HP13/20, poison0 and
XP1849.1724. The earlier uninterrupted world-arrival wait fails in this
automatic-dialogue segment and remains negative evidence.

A fresh Conjurer uses the original class/color/briefing and Captain dialogue,
then passes nine actual-mouse cave-arrival gates. The original BigSpider Bear
is alive before traversal and DEAD after the rockfall route. The player remains
HP20/20; an actual first menu Save contains LocateAldwyn2 and restores in later
processes. A close button approach and menu Save also pass. Entrance-button
activation and upper-cave arrival remain unverified: two pixel-click variants
leave the elevator disabled, and the existing type-only hover gate times out.
Two Button4 objects lie inside the viewport, so that selector cannot uniquely
identify the entrance switch. The next input action must resolve its map name
and require the normal native Use cursor before clicking.

| R321 receipt | Scope |
| --- | --- |
| `ev_15732f5c1e025f1f08c948114f400c507a2e2f053f685bd85240cca4a706773a` | Ordinary normal return, rewards and chapter Save |
| `ev_5df00bb9455a176cdd44b0f6d112d69fccb836ea2cd506a5025dacc636004c89` | Ordinary highres return, rewards and chapter Save |
| `ev_d456a56ea321e22fa3fc0ddb97b04616aba433199c3160ec32ca13e9751f5fc8` | Actual gauntlet introduction and saved objective |
| `ev_f3e0db4459b52264e29ff1e20a0dd2cc7401901f1d708f0b23f848962a9f5e07` | Fresh Conjurer cave/rockfall route and first Save |

All 6,633 frozen source inputs, 3,281 protected files, 1,562 personal files and
the R319/R320/R320-runtime artifact seals retain their hashes. The unchanged
code reuses the passing R320 four-configuration suites and strict input union
below. Earlier sealed NOT RUN cases stay unexecuted; R321 uses fresh clones.
Original executable execution and existing golden changes remain zero. Full
three-class campaigns/endings and the UI/audio/original-peer/native-generator/
stability gates remain open; the estimate stays 79% and 100% remains incomplete.

### Retained R320 findings

R320 adds actual-mouse movement toward a requested world position, using the
current viewport and client/player observations on each frame. Each move has a
600-tick bound and requires both server and client within 12 world units. It
releases the right button as soon as the server arrives, waits for the client,
and releases movement when player observations are unavailable. The next step
allows five ordinary ticks for the real input release to settle. No world,
collision, health, AI, camera, script or map-transition state is modified.

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,070 | 49 | 0 |
| Highres | 82,070 | 49 | 0 |
| Server | 77,242 | 7 | 0 |
| Server, highres | 77,242 | 7 | 0 |

All R319 terminal identities, actions and multiplicities remain, adding two
passing movement-observation tests. Those tests and the four existing named
NPC/monster-input tests pass race/checkptr2/actual-cgocheck2. They cover blocked
movement remaining incomplete, updated directions, server/client arrival
separation, button release and unavailable-player/stale-value handling.
Four clean ARM64 Go 1.26.9 products at
`82786243dde467b894f6de75d99bc853788b118f` have matching Go/REA revision,
tags, architecture and unmodified VCS metadata, with help exits 0.

An approved runtime follow-up now loads the R319 actual-menu post-combat save
in a new process and completes the virtual normal return. All 18 movement
segments observe server and client arrival within 12 world units, including
the cave egress and previously unlocked RubyKey gate. Genuine Gearhart dialog
and Done input award the SponsorshipLetter and 500 XP. The original
Escape/Save/row1/Save/Yes menu writes HP11/20, XP849.1723, poison0 and all four
journal entries complete, including ReturnToGearhart4. A registered whole-object
snapshot and inventory assertion independently observe the earned letter.

The subsequent genuine QuarterMaster dialog consumes that letter, awards
1,000 XP and naturally loads War02a. The live player reaches the original
PlayerStart `(5117.5,5071.5)`, and the captured screen shows the second-chapter
briefing. The natural transition's WORKING player contains HP12/20 and
XP1849.1724. The game exits naturally with code 0; REA verifies the owned group
and quiescent settlement without truncation or forced termination. Receipt:
`ev_dd2aa288bd96e162bc42bfb7d3a0abb6dd058afaeecfe82fd82108ddd7dd6671`.
Briefing dismissal, continued chapter-two play and a chapter-two menu save/load
remain unverified. The cloned seed AUTOSAVE still contains the initial
HP20/XP0/FirstQuest2 and supplies no chapter-two persistence evidence.

Four fresh follow-up cases are prepared and explicitly NOT RUN: the complete
chapter-entry/save route under virtual normal and ordinary normal/highres,
plus a shorter new-process Load of the earned Gearhart reward. Two automatic
approval reviews fail with model capacity; a matrix review and its permitted
single retry time out. All fail before process creation. The earlier sealed
R320 prepared case remains unrun and unchanged. Successful evidence and
unexecuted inputs remain separate; no alternate host execution bypasses review.

All 6,633 frozen inputs, 3,281 pre-existing protected files, 1,562 personal
installation files and both prior R319/R320 artifact seals retain their hashes.
The unchanged code and products reuse the passing suites above; this runtime
follow-up makes no source changes or redundant whole-suite rerun.
The 100% objective remains incomplete and the weighted estimate remains 79%.
Approved reward-reload and chapter-entry/save validation is next; ordinary and
highres return/combat, full campaign, UI/golden, audio, original Windows-peer,
native-generator and stability gates remain open.

### Retained R319 findings

R319 adds an actual-mouse approach/strike action for moving named monsters.
Each frame observes the current player, drawable, viewport, distance and normal
cursor selection. It queues ordinary right-button movement followed by one
left-button press/release. It retains no object pointers between frames,
releases movement if the target disappears, and does not strike a dead target.
No world, AI, health, camera, cursor-selection or script state is written.

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,068 | 49 | 0 |
| Highres | 82,068 | 49 | 0 |
| Server | 77,240 | 7 | 0 |
| Server, highres | 77,240 | 7 | 0 |

Every R318 terminal identity, outcome and multiplicity remains, with two new
passing input-lifetime tests per configuration. Those tests and both existing
NPC-input tests also pass race/checkptr2/actual-cgocheck2. Four clean Go 1.26.9
ARM64 products at `eefe8aaf4b07acff308ba766c928b5188da82d2b` have matching
Go/REA revision, tags, architecture and unmodified VCS metadata.

Two genuine-input virtual normal runs defeat War01A's MasterUrchin and leave
the Warrior alive with HP10/20. The first two named-target calls produce real
weapon input; the next two observe death and issue no attack. A separate
registered whole-object snapshot confirms DEAD, and the actual journal shows
OrchardQuest complete with ReturnToGearhart active. Receipts:
`ev_1b5a2125b04c7c5dee18151a0b706db2122e3de8b19b123a93c71eaf3fb8418b`
and the private R319 `warrior-master-save-v2-virtual` capture.

The second run uses the original Escape/Save/row1/Save/Yes UI. Its saved
player has HP10/20, XP349.1723, poison0, OrchardQuest4 and ReturnToGearhart2.
A new process loads that exact owned save through original Solo/row1/Load
input, restoring position `(1185.000,2687.355)`, HP10 and the dead MasterUrchin.
This is genuine combat persistence, not a direct save or script helper.
The later return route in that process fails and remains negative evidence.

The ordinary full route stalls at a market obstacle before gate contact and
correctly fails its consumed-key assertion. Fixed-duration cave-return inputs
also stall and fail the bounded Gearhart cursor gate. These establish neither
ordinary combat success nor a Gearhart reward or next chapter. A separate
pre-combat attempt writes an alive HP10 UI save, then the player dies and its
live-player assertion fails. The overall source-case failure causes the next
ordinary preparation to be rejected before game launch; it is marked unrun.
No failed case is promoted to success by its file or process exit alone.

All 6,631 frozen source/module/test inputs, the union of 3,280 pre-existing
protected files and all 1,562 personal installation files retain their hashes.
The previous R318 seal also rehashes exactly. The oracle is unchanged at
3,022 code/667 data ranges. Existing canonical PNG failures and their unapproved
proposal remain unchanged. Ordinary/highres combat, observed-arrival return
navigation, Gearhart/QuarterMaster rewards, complete three-class campaigns,
Windows-peer/audio/UI/stability gates remain open. The weighted estimate stays
79%; the overall 100% goal remains active.

### Retained R318 findings

R318 restores the compiled ScriptObject file lifecycle. The section handler now
publishes its actual `nc.obj` path to the native prefab qualifier, and the
nonempty-input merger creates and truncates its output as the original `w+b`
contract requires. The authored 183-byte SCRIPT03 becomes the exact expected
195-byte qualified stream. A clean dedicated server generates, saves and loads
it twice, crosses an intervening stock map, preserves identical fixed-seed MAP
and NXZ bytes, and exits naturally with code 0. Saved/live functions retain
`%0` and both placement offsets `(1842,1657)`, independently reconciled with
saved waypoint coordinates after the region-loader origin is removed. Floors,
walls, linked waypoints, groups and native NXZ decompression also verify.
Receipt: `ev_0943605fb4e9fe93dc7b4615aa0b50153cc87a829223b41ca844568fb67af499`.

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,066 | 49 | 0 |
| Highres | 82,066 | 49 | 0 |
| Server | 77,238 | 7 | 0 |
| Server, highres | 77,238 | 7 | 0 |

Every R317 terminal identity, outcome and multiplicity remains, with 18 new
passing parent/subtest events per configuration. The 57-event focused union
also passes with race detection, checkptr level 2 and actual cgocheck2. The
native-merger subprocess separately checks two nonempty inputs, exact output
bytes, creating an absent destination, truncating stale data, unchanged inputs
and both live-VM coordinate pairs. The prior C-only ASan attempt timed out;
no successful ASan result is claimed.

The remaining type/ZIP/game-flag diagnostic proposal is integrated. Type JSON
and the complete native-blob ZIP are encoded on the main loop before HTTP can
retire or mutate borrowed storage. Nonzero game-flag changes run their native
hooks on that loop; copied scalar response formatting remains unchanged.
Canceled requests, actual type retirement, native-blob retirement, original
JSON/ZIP bytes and mutation ordering have registered-route coverage. Eight
real native blobs also produce a CRC-valid diagnostic ZIP on the new server.
This does not establish global scalar synchronization or atomicity across all
native workers.

Four clean Go 1.26.9 ARM64 products at code
`9e14fe95701bb8a7e2c88b664024d723649d53a6` have matching Go/REA revision,
architecture, tags and `vcs.modified=false`, with natural help exits 0. The
original oracle verifies 3,022 code and 667 data ranges, preserving every
previous record. All 6,629 frozen inputs, 3,278 pre-existing protected files,
and 1,562 personal installation files retain their exact hashes. GAME.EXE was
not executed. Exact file-flow and negative-case scope is recorded in
[native-theme-reader.md](native-theme-reader.md#r318-compiled-script-file-lifecycle).

Separately pinned clean R317 normal products now provide actual Wizard
Wiz01A objective completion and natural Wiz02A entry in virtual and ordinary
time. Real barrel interaction opens the fire approach; SapphireKey contact,
apprentice dialog, robe/book pickup, the cycling elevator, Horvath's reward,
lab portal and original map exit follow genuine input. Saved XP is 1761.1278;
the earlier post-introduction value is 111, correcting prior private prose.
Actual row-0 Load in a new process restores the natural AUTOSAVE, equipped
robe, three spells and MeetHorvath journal. Ordinary earlier cave/pickup stages
use genuine menu-save checkpoints and are not one uninterrupted fresh chapter.
Boundary receipts: `ev_4215649beb2d8c17e051c53906f320027796cedc9019affdab12b2e8d1a35f4e`
and `ev_f135e22d604835fe60d02269ca38d97917d9517f29df0d3a1d1711201858ee5e`.
New-process Load receipts: `ev_cc49d10fa5377816535399cd72e03d347d4023bbf5610ada34f6bcd67138db82`
and `ev_11a2ece90609a887750fe4ec2bc29994a36a5ff1708070058865642f29b398b9`.
These results are not attributed to the newer R318 products.

R317 Warrior products also pass actual menu Save/movement/Load and original
RubyKey-gate consumption in ordinary normal/highres runs. A separate ordinary
process restores an ordinary-produced save with the key and journal. The
cave bats are defeated, but MasterUrchin and subsequent objectives remain
incomplete. All 2,046 Wizard and 7,980 Warrior sealed artifacts rehash exactly;
failed inputs and deaths are retained. No gameplay state, HP, lock, objective,
script callback or personal save was injected or edited.

Two additional two-prefab definitions, including opposite North/South markers,
fail during generation before successful load. Their natural game exit 0 is
not a successful semantic result; controller exits are 1 and both captures
are retained. Runtime merging of those definitions remains open despite the
passing actual native-merger regression. Existing two canonical PNG failures,
the unapproved golden proposal, full three-class campaigns/endings, original
Windows-peer interoperability, broader audio/UI and sustained stability also
remain open. The weighted estimate remains 79%; the 100% goal stays active.

### Retained R317 findings

R317 corrects the native qualifier metadata used by compiled prefab functions.
The pinned `asm` reader assigned both coordinates to X and left Y zero. The
live server now decodes the ordinal suffix and two signed 32-bit coordinates
from the original CRT contract, preserving the full function name. Static
GAME.EXE instructions and independently read PE bytes establish delimiter
skipping, string ordinals, decimal prefixes, wrapping arithmetic and separate
builtin coordinate delivery. Incomplete malformed qualifiers retain the
upstream fallback; no safe original result is invented for `atoi(NULL)`.

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,048 | 49 | 0 |
| Highres | 82,048 | 49 | 0 |
| Server | 77,220 | 7 | 0 |
| Server, highres | 77,220 | 7 | 0 |

All R316 terminal outcomes and multiplicities remain, with 29 new passing
events per configuration. All 6,627 frozen source/test/module inputs remain
unchanged. The 24-event real SCRIPT03/builtin focus changes from 19 FAIL and
5 PASS before the correction to 24 PASS after it, including race, checkptr2
and actual cgocheck2. The integrated lifetime/input/qualifier union passes
78 events normally and under those strict checks.

The registered legacy-script HTTP route now traverses and encodes all borrowed
function and value storage on the game loop. Additive tests cover cancellation,
retirement before HTTP output, and byte-identical empty-VM JSON. Flag, type and
memmap diagnostics still require separate lifetime review; no global race-free
claim is made. A bounded named-NPC input action observes a live drawable,
moves the actual mouse, waits for the existing cursor selection and clicks
once. It changes no world, camera, script, AI or dialog state. Its moving-target
and vanished-target tests pass. Clean R317 virtual Warrior progression uses
that actual input for the guard, QuarterMaster and Gearhart, then obtains the
RubyKey with HP20 and a natural zero exit. Receipt:
`ev_8d04a853d45351d7c6d56bfb5ff4e76d6bca499a77b8e87d0a72d90c8669899d`.
Ordinary normal/highres runs remain in progress at this checkpoint.

Four clean Go 1.26.9 ARM64 products at code
`53d4af42cf3d77bb3ddbc0f96ff1168acfd7ed86` have zero help exits and matching
Go/REA architecture, revision, tag and `vcs.modified=false` metadata. The
original oracle verifies 3,019 code and 665 data ranges, adding five of each
while retaining every prior record. GAME.EXE was read statically and never run.

A new 671-byte authored AreaMap includes a structurally decoded 183-byte
SCRIPT03 object. The old clean `62c46ea` server qualifies the named waypoints
and naturally completes two generations, but saves an empty ScriptObject and
returns null for the live script list. That zero process exit does not pass
the script-payload gate. The saved section is exactly `010000000000` and the
failure is retained for R318 file-path investigation. Receipt:
`ev_60e18b9935645bbba965c776a53882be22adaf535c1eb834c71d5c3ce44f053b`.

Separately pinned clean `62c46ea` Wizard runs now execute original Wiz01a
spell pickup/casting, the Urchin encounter, Horvath's automatic introduction
and reward dialog, and the LesserHeal/potion/journal rewards in both clocks.
Receipts: `ev_92e57a952506bd88f079813abf89af3a701be1bcf1112251f8f08fdb90875527`
and `ev_26f5704fdcb311be584163b71833354e74f5d7fac5e5f3c4531382372a946577`.
End results agree, but exact positions and intermediate HP differ. The stock
fallback can contribute to the encounter; no all-lightning-hit claim is made.
The old autosave contains the initial state, so post-reward persistence remains
a separate actual-menu Save/Load gate being pursued in R318.

All 3,275 pre-existing protected files and all 1,562 personal files
(571,159,691 bytes) remain unchanged. Existing manual-save PNG comparisons
remain failing and the two-image proposal is still unapproved. Full three-class
campaigns/endings, original Windows-peer compatibility, broader audio/UI and
sustained stability evidence remain open. The weighted estimate stays 79%;
the 100% goal is active. REA sampling and before/after filesystem limits remain.

### Retained R316 findings

R316 fixes a reproduced generated-waypoint ownership failure and ten further
world diagnostic operations. A required AreaMap with authored floors, walls,
linked waypoints and groups saved correctly, then panicked with `incorrect free`
during map teardown. Its C allocations now transfer into the tracked allocator
when generated waypoints become world-owned. The same 464-byte library and
encrypted theme pass generation, saving, loading, an intervening stock map,
fixed-seed repetition and natural shutdown on the clean corrected server.
The entire generated MAP is byte-identical to the artifact saved before the
old teardown failure. The exact scope is in
[native-theme-reader.md](native-theme-reader.md#r316-authored-geometry-and-waypoint-ownership).

All 6,622 frozen source/test/module inputs remain fixed through four integrated
configurations. Every R315 outcome and multiplicity is retained, with 43 new
passing parent/subtest events in each configuration:

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 82,019 | 49 | 0 |
| Highres | 82,019 | 49 | 0 |
| Server | 77,191 | 7 | 0 |
| Server, highres | 77,191 | 7 | 0 |

The combined 49-event focus passes with race detection, checkptr level 2 and
actual `GOEXPERIMENT=cgocheck2`. Real registered HTTP routes test cancellation,
ordered/atomic object flags, unchanged 302/400 responses, and original JSON bytes
after borrowed native values have been retired. Object lookup and collections,
map indexes, teams, players, walls, waypoints and drawables now finish traversal
and encoding on the main loop. Script/flag/type/memmap diagnostics remain outside
this repair; no global race-freedom claim is made. The new server also answers
nine affected diagnostic variants after each of two generated-map loads.

Four clean Go 1.26.9 ARM64 products at code
`62c46ea5e21f59a80bae57ed3b68541311b0f465` have zero help exits and matching
Go/REA architecture, revision, tag and `vcs.modified=false` metadata. The oracle
still verifies all 3,014 code and 660 data ranges; no original record changed
and original GAME.EXE was never executed.

Separately pinned clean `0666df6` normal products extend Wiz06a's actual input
route. The stock Invisibility reward/cast is followed by walking into the
unlocked guardhouse, clicking its SilverKey, and returning by original
113→114→115→116→117 links. The locked Gate consumes the key and the player
reaches city waypoint120 at `(2914.665,3845.857)` with HP20. The exact same input
passes in virtual and ordinary clocks with natural zero exits. Receipts:
`ev_f01b10016ff40a4fda6a52acd26069584a14907bdc0a237cf5fe8cbf4e8019c2`
and `ev_f1cdfce91c15ff3b12d08fd68a9b3e7a46d7ecee2d11967d4941ed90e1f72ea4`.
The next stock gate requires GoldKey, held by Guard1 in both decoded original
inventory and a live snapshot. Two staff-input attempts prove no hit, defeat
or loot. Wiz06b, completed objectives, post-key persistence and endings remain
unverified. Deaths and wrong-click/return assertions remain negative evidence.

Latest normal/highres `0666df6` manual-save runs still fail both unchanged PNG
comparisons: 943 live-load pixels and 889 fresh-load pixels. Each actual PNG
is byte-identical to R308. All four runs pass the observed save/overwrite/state
restoration and movement checks. The separate provided Host Load passes real
Load and movement in virtual and ordinary time, but a persistent transparency
defect is not causally established. No rendering implementation or expected
image was changed, and the prepared two-image proposal remains unapproved.

All 3,271 pre-existing protected files and all 1,562 personal files
(571,159,691 bytes) remain unchanged. Owned R316 process groups are verified
quiescent and their selected file hashes are rechecked. REA process sampling,
before/after filesystem observations and environment/image-binding limitations
remain. The weighted maturity estimate remains 79%; the full 100% goal is active.

### Retained R315 findings

R315 restores the original CRT stream at the two legacy C ABI shims. The
previous unrestricted Go random values put required prefab rooms outside the
map. The same authored south-marker AreaMap/theme now completes generation,
save, load, byte-identical fixed-seed repetition and natural dedicated shutdown.
Saved/live object positions and an exact native NXZ round trip agree. Static
original comparison, stream ownership and bounded prefab scope are documented in
[native-theme-reader.md](native-theme-reader.md#r315-crt-stream-and-required-prefab).

All 6,617 frozen source/test/module inputs remain fixed through four integrated
configurations. Every R314 terminal outcome and multiplicity is retained, with
six new passing regression events in each configuration:

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 81,976 | 49 | 0 |
| Highres | 81,976 | 49 | 0 |
| Server | 77,148 | 7 | 0 |
| Server, highres | 77,148 | 7 | 0 |

Counts include parent/subtest events. The seven-event focused union also passes
with race detection, checkptr level 2 and actual `GOEXPERIMENT=cgocheck2`.
The oracle verifies 3,014 code and 660 data ranges, preserving every earlier
record. Four clean Go 1.26.9 ARM64 products at code
`0666df6e687bffb82986dd6356c3607020130f0d` have zero help exits and matching
Go/REA revision, tag and `vcs.modified=false` metadata.

Separately pinned R314 normal/highres products now have actual ordinary-clock
exit handoff evidence for four authored directions, with destination readiness,
server/client equipment preservation and natural exits. This extends the old
generation-only evidence; it does not attribute those gameplay runs to the new
CRT products. The exact matrix is in
[native-theme-reader.md](native-theme-reader.md#r315-actual-exit-movement).

R315 campaign exploration uses only the clean R314 normal client, explicit
virtual time, headless input and audio disabled. Actual stock book pickup,
reward UI, icon drag into quickbar spell39 and A-key Invisibility casting pass.
The running approach still dies; changing the post-cast input to walking reaches
`(2909.500,4717.415)` with HP20/20 and both guards stationary. The stock notice
countdown nevertheless changes from 0 to 58, and the next north input stalls
at map geometry. This establishes neither complete non-detection nor Wiz06b
entry, objective completion or post-reward persistence/reload. The wrong-book
Lock44 assertion failure and other incomplete route attempts remain negatives.
Casting/run and walking receipts are
`ev_aed2478cce80e6270b4b02b7d467a45b14c2a90468c2312d0640839916fe8206`
and `ev_e1e535798a12d015b410b899addddaaa9ea6e30cf61e699fc52690a394007dac`.
No new ordinary-clock/highres campaign route or natural ending is claimed.

During this run the old `/private/tmp` checkout lost 6,500 tracked files and Git
metadata for an unknown cause. Its failing audit is preserved. A fresh independent
workspace clone at `a2b6d0c` matched all 6,984 pre-change tracked files against
the intact snapshot before applying this correction. All 3,270 pre-existing
protected files verify unchanged there. All 1,562 personal files
(571,159,691 bytes) and original Save/settings remain unchanged. Existing manual
save/load PNG discrepancies remain open; no expected image was changed.

These bounded advances retain the weighted maturity estimate at 79%.

### Retained R314 findings

R314 fixes two reproduced lifecycle failures. Dedicated classic generation no
longer invokes the client progress renderer; its rendered path restores original
signed tick comparisons and writes the progress packet without unaligned access.
The whole-world `GET /debug/nox/objects` collection response now traverses and
encodes objects inside the server loop, returning only owned JSON bytes to the
HTTP goroutine. The previous observer could follow an inventory freed during
campaign autosave. Request cancellation cannot leave the server loop blocked on
the response channel. Other optional debug endpoints are outside this correction.

The exact final source, test and module inputs total 6,615 files. All four
integrated configurations pass with every R313 outcome and multiplicity retained:

| Configuration | PASS events | SKIP events | FAIL events |
| --- | ---: | ---: | ---: |
| Normal | 81,970 | 49 | 0 |
| Highres | 81,970 | 49 | 0 |
| Server | 77,142 | 7 | 0 |
| Server, highres | 77,142 | 7 | 0 |

Counts include parent and subtest events. Each configuration adds 18 passing
events: 13 whole-body progress events and five snapshot lifetime/cancellation
tests. The focused union also passes 34 events with race detection, checkptr
level 2 and actual `GOEXPERIMENT=cgocheck2`; the five API tests separately pass
under race detection. All 3,267 existing protected tests/fixtures/YAML retain
their hashes. All 1,562 personal installation files (571,159,691 bytes), original
Save/settings and existing PNG expectations are preserved.

Four clean ARM64 products are built from
`dd142aa82dcbcc0061683c99b3ee4c44b5ae484e`; their help exits are zero and REA
build metadata agrees with Go on Go 1.26.9, architecture, tags, revision and
`vcs.modified=false`. The clean dedicated product passes five consecutive
ordinary-clock authored exit generations and natural shutdown. Saved destinations,
positions and four exact NXZ round trips are documented in
[native-theme-reader.md](native-theme-reader.md#r314-dedicated-progress-and-saved-exits).

The corrected clean normal client also repeats the Con11a greeting with exactly
three collection requests at observed phases 0, 1 and 2. Each response is complete
HTTP 200 JSON with 291 objects; game and wrapper exit naturally with code 0.
This bounded check uses an explicit virtual clock and unchanged phase inputs.
It confirms a working corrected response through that transition, not general
race freedom. Receipt:
`ev_71ff2c248f06b3aacc02a4fe374b86d8bccf8a915bfbeddd680eae92500c00ae`.

Two further R314 runtime matrices use the separately pinned clean R313 code
`0eaa6e6d5a606aeace14d06d88697f1e70bdbda8`; their timings and identities are
not attributed to the newly built products.

The ordinary-clock network profile passes normal and highres pairs, each with
one dedicated server and two sequential independent client processes. The same
server observes players `0,1,0,1,0`, a connected So_Beach-to-Estate transfer and
a fresh client connection. Three native UDP downloads per pair produce six
files matching stock hashes. All three movement phases advance the authoritative
unit and matching client drawable. Normal and highres Estate transport/reload
take 3.579s and 3.611s; the retained ordinary input schedules reach movement in
29.859s and 29.908s under a separately declared 120s stage bound. The original
20s virtual end-to-movement assertion is not reused. Controller runtimes are
329.534s and 330.366s, with 646 and 648 raw samples.

A controlled stock Spider, placed and targeted through the existing Lua API,
executes its actual attack callback and changes player HP from 150 to 140.
The owning client receives exact health packet bytes `43 8c 00`; no direct damage
or HP writes supply the result. Six product processes exit naturally with code 0.
An inherited process sandbox allows only loopback outbound traffic; child-only
proxies reject optional external lookups. Machine settings are untouched.
Normal/highres receipts are
`ev_dd828c3c819d9cd7250d716494989bad26980d7ef366003ab2b433dbb6ba2ea0`
and `ev_cf359ecc6f1883030a07e70ce524b9b8cfdef67a1676425b53cc5b6ac8b79be6`.
The remote-client server-TickHook metrics remain zero, so they provide no client
FPS/clock-ratio measurement. This does not independently read the rendered HUD
scalar or establish original Windows-peer compatibility.

Wiz06a normal and highres pass the Captain's two stock dialogs under both
virtual and explicit ordinary clocks. Retained input remains blocked for 30
iterations in the second dialog, then its ordinary Done input releases movement.
The same input moves the player about 28.32 units. Actual AUTOSAVE and WORKING
files each retain journal `Wiz6SneakIn`, type 2; the read-only decoder does not
certify every other save section/checksum or demonstrate reload. Ordinary receipts
are `ev_cdd3209167c9b98e580e0a7032e5e4b79a613a02d4e0eb0a9a9f9f66b349974b`
and `ev_f17b121dfadaecfbe904611ae7b3501c4d463ac6e56cc3cc7049ca8ea83c078f`.
Con11a normal and highres virtual runs reach the untouched greeting trigger;
two ordinary Done inputs advance stock global 48 through `0,1,2` and release
the guard phase. Controlled direct map entry is explicit preparation in both
chapters. Full routes, later battles and story-triggered endings remain open.
The earlier geometry-precondition failure and unsafe debug-observer crash remain
retained negatives, distinct from the passing stock callbacks.

These bounded advances leave the weighted engineering estimate at 79%.

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
theme/AreaMap inputs absent from the retained stock map inventory. The R310
theme reader now transports native names, file handles and child theme
pointers. Conditional-expression readers also retain native files and player
fields. The generation entry, save and placement boundaries preserve native
objects. An authored single-room theme now completes ordinary-clock
generation, map/NXZ save, generated-map load and natural shutdown. Broader
classic input combinations and callback/layout classification remain open; see
[native-theme-reader.md](native-theme-reader.md).

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

## R310 theme reader validation checkpoint

Revision `98080022714fdfa52fb03acf22cfec0bc6c0bdc8` preserves native theme-name,
encrypted-file and child-parser pointers, restores the original 60-byte
algorithm value buffer, and reads the skeleton table through its native
side slots. The original integer declarations remain available as token
adapters. The fixed 1,116-byte theme layout and all previous oracle records
remain intact.

| Configuration | PASS events | SKIP events | Passed packages | FAIL |
| --- | ---: | ---: | ---: | ---: |
| Normal | 81,915 | 47 | 40 | 0 |
| Highres | 81,915 | 47 | 40 | 0 |
| Server | 77,087 | 7 | 36 | 0 |
| Server, highres | 77,087 | 7 | 36 | 0 |

Each suite retains every R307 terminal outcome and multiplicity and adds
nine passing parent/subtest events. The source-bound C fixture exercises
236 bounded cases in each of O0, O2 and ASan+UBSan. Actual CGo reads temporary
encrypted themes through native pointers and registered high-bit name
tokens, verifies file closure, room/hall graph cleanup, partial EOF writes
and untouched guards. The focused union passes 157 events with race,
checkptr level 2 and cgocheck2.

All 3,249 pre-existing test/fixture/YAML files and all 1,562 personal files
(571,159,691 bytes) retain their hashes. The original executable hash and
both existing manual-save PNGs remain unchanged. The oracle verifies 2,999
code and 659 data ranges. All four clean ARM64 products build, return help
successfully, and have matching REA/Go build metadata. The new checks cover
bounded theme parsing; conditional expressions and complete generation
remain open. The weighted maturity estimate remains 79%.

## R313 integrated regression checkpoint

The combined native mapgen, SDL input, host reset, audio ownership and
ordinary-clock changes pass the four Darwin ARM64 configurations below.
All 6,611 compiled-source, test and module inputs retain their frozen hashes
before and after each suite.

| Configuration | PASS events | SKIP events | Passed packages | FAIL |
| --- | ---: | ---: | ---: | ---: |
| Normal | 81,952 | 49 | 40 | 0 |
| Highres | 81,952 | 49 | 40 | 0 |
| Server | 77,124 | 7 | 36 | 0 |
| Server, highres | 77,124 | 7 | 36 | 0 |

Every R310 terminal outcome and multiplicity is retained, with 37 new passing
events per configuration. The client configurations additionally skip two new
opt-in native OpenAL tests; the separate real OpenAL null-driver gate executes
them and passes all 70 events normally and under strict settings. Only the
same two existing server test families need native runtime address-label
normalization; fixed low addresses and event multiplicities stay literal.

The final mapgen/timing/host-reset union passes 195 events with race,
checkptr level 2 and `GOEXPERIMENT=cgocheck2`, including the added complete
buffer allocation/free regression. The timing/connection union passes 42
events, and the tagged native SDL input union passes 17 events in normal,
highres and strict modes. All 3,252 pre-existing protected tests, fixtures and
YAML files retain their hashes; existing expected values and PNGs are unchanged.
The oracle verifies 3,007 code and 659 data ranges while retaining every
previous record. These are parent/subtest events, not a coverage denominator.

All four ARM64 products build from clean code commit
`0eaa6e6d5a606aeace14d06d88697f1e70bdbda8` and return help with exit 0.
REA and Go agree on Go 1.26.9, Mach-O ARM64, build tags, revision and clean
status. The clean dedicated server also completes the authored initial
classic generation/save/load scenario and exits naturally with code 0;
its map/NXZ bytes match the source-bound development run. Receipt:
`ev_f18e248ce22c30ceb14f10a0d408ee43250cc281bc5f764c4bcb0d6634a3e5af`.
The ordinary audio soak and two broader generator variants retain their
separate development-artifact identities. Source, suite, build and runtime
records are retained under `tmp/opennox-resume-r313-20261018` in the local
workspace. Subsequent changes in this checkpoint are documentation only.

## R313 input, audio ownership and ordinary-clock checkpoint

The weighted maturity estimate remains **79%**. The following controlled
results close specific defects and evidence gaps; they do not establish a
complete gameplay, campaign or compatibility denominator.

The SDL keyboard adapter now rejects out-of-range and unmapped scancodes
before forwarding events. Unmapped keys previously inserted zero, the game
keyboard batch terminator, and discarded later mapped presses or releases.
Real native SDL dummy-driver queues reproduced that loss in both menu and
letterboxed gameplay layouts. A separate direct-boundary regression caught
invalid table indices. The isolated candidate passed all 47 SDL package
events normally and with highres, race, checkptr level 2 and cgocheck2,
including all 30 existing resize/context/attribute events. The integrated
tree subsequently passed the 17 additive input events normally, with highres,
and under the strict settings. Its production source and both additive files
are byte-identical to the isolated candidate. These tests cover native event
transport and software input state; Cocoa/Retina/compositor, physical input,
relative mouse and IME behavior remain outside this result. Receipt:
`ev_0efcfae293ed46a62d55733f0a5e6390b49b4b27e4228ee3147738737ae99403`.

Sample and stream teardown now delete an immutable copy of the originally
allocated OpenAL buffer IDs. Their mutable recycling lists could contain
duplicates while omitting a still-owned ID. The identical native partial-queue
regression failed both baseline closes and passed the correction. All 70
native `TestOpenAL` events passed both normally and with race, checkptr level 2
and cgocheck2, using OpenAL Soft's silent null driver. The integrated tree
also passes all 70 events in both settings, retaining existing movie-buffer
ownership, drain, seek, statistics and teardown checks. Production source
matches the isolated candidate exactly; its additive regression differs only
by a comment.

`NOX_E2E_CLOCK=ordinary` retains the production clock, Sleep, RNG, millisecond
ticks, local calendar, time-based effects and frame-limit flag policy while
the scenario counts input iterations separately. The default remains virtual.
The integrated timing, real-Sleep, checkpoint, connection-gate and frame-limit
union passes 42 events. An earlier 34-event command misspelled the
connection-gate test prefix and did not execute those tests; the 42-event run
closes that omission. The initial mapgen/timing/host-reset focused union
passes 191 events normally and under the strict settings; the final buffer
regression raises the strict union to 195. These overlapping
event counts are not a coverage percentage.

A source-bound normal ARM64 development artifact over `eb7f60ff`, with
explicit ordinary-time and buffer-ownership overlays, completed a private
headless host run. Raw active-start-to-child-exit events span 317.336 wall
seconds within a 900-second execution bound; the controller reports 317.530
seconds measured after shutdown and 425.627 seconds overall. All 416
metric/state samples are retained. The 227 stable post-scenario samples span
232.724 seconds and recompute to 30.001203 game FPS, within the declared
27–31 FPS bound. Platform milliseconds advance at 0.999933604 times wall time,
within the declared 0.95–1.05 bound.

The run preserves every original host gameplay/audio assertion, exercises
real input pickup/drop, two hammer attacks, a stock hostile Spider's first
attack, walk/run and TriLevel/Estate transitions with host/drawable rebinding.
Actual OpenAL Soft 1.25.2 core output selects the MacBook Air speakers and
consumes native buffers in all six effect phases and after transitions.
Ordinary WindowClosed input gives child and controller exit 0 without
termination cleanup. Context/device cleanup completes without OpenAL
warnings. All 21 controller checks and 27 retained audit checks pass.
Receipt:
`ev_eb6a9b4e44e7ac1dc27fe5e94dd352bf6fbd3a0c90f0a2bea2251efae6b080d3`.

The 309 paired host/drawable samples share netcode 2. Their largest non-atomic
position distance is 8.569 units; each of the four real-input movement phases
exceeds eight units. This is one host process, not independent socket-peer
replication. Goroutines range from 8 to 12 and end at 9; threads range from 7
to 8 and end at 8. Final Go heap is 207,903,448 bytes, peak 274,121,080 bytes;
child maximum RSS is 396,623,872 bytes. These bounded observations establish
neither general leak freedom nor human-heard audio fidelity. E2E state/input
isolation, controlled combat fixtures and owned optional-movie withholding
remain part of the profile. It is not a clean combined release run.

## R313 controlled campaign endings and host lifetime

Clean R310 normal/highres products at `98080022714fdfa52fb03acf22cfec0bc6c0bdc8`
pass all six controlled class/configuration ending-dispatch runs. Each loads
the final map directly, explicitly calls an unchanged stock function and
selects the corresponding movie once:

| Class | Final map | Explicit function | Selected movie | Normal/highres |
| --- | --- | --- | --- | --- |
| Warrior | War11a | EndScene3 | warrior.vqa | pass/pass |
| Conjurer | Con11a | TheEnd1 | conjurer.vqa | pass/pass |
| Wizard | Wiz11a | FinalFadeOut | wizard.vqa | pass/pass |

All six captures exit naturally with code 0, settle without termination
cleanup and retain their complete selected capture budgets. Movies are
deliberately dismissed through ordinary input. These virtual-clock,
audio-disabled runs do not establish objective/route completion, story-driven
final triggers, boss combat, full movie playback or persistent ending bits.

The unchanged Warrior scenario's post-EndGame Screen exposed a real lifetime
defect: `serverPlayers.ResetAll` cleared player slots but retained cached host
player/unit references before Coop teardown freed their storage. ResetAll now
clears both references at their owner. A new lifecycle regression fails the
baseline and passes the candidate and strict settings. Both normal/highres
source-bound candidates pass the byte-identical previously failing scenario,
retaining both Screen actions; its newly created PNGs are observations with
zero golden comparisons. The integrated owner fix and regression exactly
match the validated candidate. Receipts:
`ev_6c0aa52060a4e678df892febcc7c13a014f3b01bfbf2ee928a72bed967a44bf0`
and `ev_da86516f50727d6298233075cfb7681434e733336d6e7a8417b65bc6195ca322`.

Direct War11a entry also establishes that the observed autosave wait was
waiting for normal stock dialogs. Two ordinary Done inputs complete the
unchanged autosave gate on both clean products; no pause/unfreeze correction
is needed for this case. Full three-class routes, chapter save/reload and
legitimate progression remain open, including the earlier Con11a/Wiz06a
full-route questions. All 3,252 pre-existing test/fixture/YAML files and all
1,562 personal installation files retain their hashes. The original Windows
executable was not executed; existing candidate PNGs remain separate.

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

The R308 investigation reproduced the same 889-pixel failure on clean
`2fe0db447822fd0e00e891440c0891c6ac0ea4ab` normal and highres products.
Both actual PNGs are byte-identical to the retained failure. A private passive
overlay around the single real `DrawPlayer` call identifies the central actor
as Jack, netcode 2. At frames 70 through 74 its native status is `0x10`, buffs,
fade and flags70 are zero, and incoming/outgoing alpha enable is false. The
cached alpha byte is 28; it does not activate blending. The overlay's final
PNG remains byte-identical to the release output. This observation applies
to the copied manual-save case and does not resolve every Host Load report.

The PNGs were introduced in `51a8df7f0` on October 1. The following day's
weapon/armor material corrections (`0bba5d6bf`, `3c7abc0fa`) and five-bit
RGB555 arithmetic corrections (`255653c28`, `716088799`) were not reflected
in those images. A diagnostic overlay restoring only the historical material
code makes 425 pixels match the old fresh-load golden; the other 464 remain.
Restoring the historical RGB555 split as well makes both unchanged original
live-load and fresh-load PNG comparisons pass. These are explicitly modified
historical controls, not current-release successes. The current rendering
code passes 175 original-contract test events, including the retained
65,536-word original color transcript.

Two proposed PNGs were tested in separate copied data trees with the current
normal and highres release products. All four seed/live-load and fresh-process
load runs pass exact pixels, state restoration, 329.469 units of real input
movement and natural exit. Seed runs also verify overwrite refusal, confirmed
overwrite and unchanged autosave on refusal. Public input scripts, the
100-tick waits, scalar comparisons and renderer code remain unchanged.
The canonical PNGs remain preserved pending the user's decision under the
existing test/fixture preservation condition; candidate results are not counted
as a pass of the unchanged public golden. REA candidate receipts are
`ev_b3f921ca0e880152d4dd2239eb25212f5ca8d01abdde109425439cd8ea6dee82`,
`ev_a1ee2635f49dc4f223b9478199f37faa671ef0a28530702b23f8047e9dd1d19b`,
`ev_3b0fd6ca44f15a3f0743baa557391fd850e546b0caebb08a13324e4e28794733`
and `ev_d6ac32ddaa87b90f11229ef53720bcebc089b6150c0057063518ce5d7ef4cc59`.

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

## Native UDP download, transition and fresh reconnection

Clean Go 1.26.9 ARM64 products from `e6673a55be5ef3357e4391482cdc925dbeb391f8`
pass a separate normal/highres profile with stock client maps withheld only
from owned data copies. Each pair transfers So_Beach to client 1, transfers
Estate while that client remains connected, then transfers Estate to a fresh
client process on the same running server. All six UDP transfers produce
`.nxz` and `.map` bytes matching the preserved originals. No game-code change
is required; the two new scenarios retain the existing live-peer scenario.

The initial unchanged scenario downloaded correctly but did not move. Separate
diagnostics establish real stock chat-help/MOTD dismissal and the remaining
observer state. The current native map request deliberately calls
`PlayerGoObserver`; normal Action input is needed after each transfer. Both
final profiles verify received Action packets, observer-to-normal reports,
normal unit flags and independent server/drawable movement. Earlier negative
captures and their unchanged inputs remain preserved.

| Profile | So_Beach walk | Connected Estate walk | Fresh Estate walk | Handoff |
| --- | ---: | ---: | ---: | ---: |
| Normal | 174.839 units | 84.637 units | 104.067 units | 11.248 s |
| Highres | 114.602 units | 180.572 units | 149.653 units | 11.459 s |

All six walk phases retain matching netcodes; their maximum observed
replication distance is 1.647 units. Each server observes `0,1,0,1,0` players,
retains identity across its connected handoff, and survives for the fresh
connection. All six processes exit naturally with code 0, without termination
cleanup. REA retains complete capture budgets and owned-process settlement.
The profile, fixed bounds, input prerequisites and remaining limits are in
[native-map-transfer.md](native-map-transfer.md).

These are headless virtual-tick clients with real Sleep and ordinary-clock
dedicated servers, with audio disabled. They do not establish original
Windows-peer interoperability, combat, WAN, ordinary-time clients or SDL
compositor parity. The overall weighted maturity estimate remains 79%.

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
