# Original lateral-dodge probe

Original GAME.EXE SHA256:
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
No executable or asset is included.

The independent inputs and native getter/consumer/golden were frozen before
engine changes. Actual original execution covers 2,361 root contexts and four
ancillary RNG contexts per FPU profile, including three raw faults. All 550
instruction bytes execute: root547C50 337, float RNG416030 81, integer
RNG415FA0 67 and attack-pop5471B0 65. Fifteen alignment bytes are excluded.
The 18,465 bytes of direction/RNG/constants/switch data remain guarded.
All these ranges already exist in the oracle manifest; no duplicate ranges
or additional coverage credit are added.

Only Trace535250, Obstacles50B580, Tile411160, Push50A260 and Pop50A160
are declared world services. Original RNG and the attack-pop child execute.
Original heap, code/data, nonvolatile registers and FPU guards remain checked.
The root overwrites its own first by-value stack argument for distance only
at547CAB/547CBC; every other caller byte remains unchanged. Native physical
stack layout is not compared with the original 32-bit calling convention.

The production e3f and nearest23f outputs differ in 1,569 root contexts.
They remain separate; native gameplay comparison uses e3f. Original v3 adds
Frame/FPS load observations at547D61/547D67. Removing only these observations
reproduces every v2 result exactly. Earlier producer-preparation failures,
including the overly strict caller-stack guard, remain private and preserved.

The historical JSON key `speed` denotes raw object+176, which independent
C/Go layouts identify as **Shape.Circle.R**. Actual SpeedCur is object+544.
Neither that wire key nor original input/output values were rewritten after
this distinction was found. An additional original control sets only +544
to zero: all 2,361 root and four ancillary results reproduce the exact same
gzip bytes. The old Go helper's SpeedCur substitution and its collapsed
zero-speed expectation are therefore incorrect. Three older fixtures bind
their existing numeric values to radius; a fourth corrects the disproved
zero-current-speed expectation. Old source copies and failures are retained;
test identities and the remaining coordinate/RNG/assertion values stay fixed.

Native comparison checks both public C and direct Go paths, the same mutable
local ray through world calls/retries, cached entry direction/origin, live
position/radius reloads, five-attempt short-circuiting, original Pop child,
null pushes and post-Push frame/FPS reads with DWORD wrap. It compares ordered
services and boundary-visible stores, service/RNG writes and the entire
C-owned native allocation including padding and high uintptr argument bits.
Direction indices0..255 are credited; other signed-short indices are not.
The native arithmetic preserves precision53/ToZero, separate multiply/add
spills and x87 negative indefinite NaN bits without changing thread FPU state.

Before porting, the actual old public C SIGSEGVs only after valid above4GiB
C/Go layouts, distinct radius/speed fields, whole initial-record guards and
existing typed sound/action getters pass. C11 _Generic selects an exactly
typed legacy-int or native-object-pointer wrapper; no incompatible
function-pointer cast is used. A separate immutable literal original case
also fails the old Go helper's radius-dependent destination. Their exact
binaries and raw failures remain private.

Separate production composition uses native objects and the default
RNG/clock/Pop/Push bindings with declared geometry observers. It checks real
Logic RNG consumption, independent Other RNG preservation, original literal
coordinates, wrapped timer, actual initial-IDLE replacement and whole native
records. End/Cancel handlers are inactive. This is not full map geometry,
all callback/network behavior, campaign completion or physical audio proof.

Sealed production gzip SHA256:
`c0cec2ea8230ea72deb7cba8a0ed850bf895b57bf2af04d92edf9c8cebae9166`.
Decoded SHA256:
`c6513a855625a0deb4e0d978230bcc123689e3489b8eebbd1fce06ffe9b79579`.
Private records: `ROOT/monster-check-dodgeables-native-entry-port`.
