# Original raw retreat-to-master entry

Reference is frozen before porting public `5456D0`, against clean engine
`1fc4dd99bb8de3677ff25606257c87a5dfafa9dc`. Independent inputs, producers,
original runs, instruction/memory guards and native before controls live in
`ROOT/monster-retreat-master-native-entry-port`.

The original GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Each FPU profile executes **942 parent contexts plus one ancillary null-buff
context**, including **17 raw faults**. All instruction addresses in the
192-byte parent and real 83/40/43-byte health/retreat/cast predicates and
35-byte buff getter execute on valid inputs: **393 instruction bytes**.
The two four-byte original constants are binary32 one and thirty. The parent
has no trailing alignment bytes before `545790`; the helpers' 39 alignment
bytes receive no instruction credit.

Only original Push `50A260` and Pop `50A160` are declared service boundaries.
Their whole before/after states, ordered calls, mutation inputs and returned
native item identities are captured. Health, cast, buff, distance arithmetic,
short circuits and all payload stores execute original instructions, without
instruction, floating-point or result substitution. Whole original image,
heap and guarded caller bytes, valid-return nonvolatile registers, CW and
TOP are checked. The health helper's by-value scratch stores remain inside
the parent's private stack frame.

Startup PC53/ToZero and PC53/nearest differ only on `distance/50/6e-07` and
`distance/6e-07/50`; their full goldens remain separate. The native comparison
uses the startup reference. Inputs cover admission/health/status/buff gates,
self/no/live owners, raw non-Monster classes, null data/health/unit, strict
distance edges, unspilled subtraction, subnormal/nonfinite bits, independent
random bit patterns, rejected pushes, shared returned slots, switched MUD
and owner identities, raw payload reloads and partial effects on faults.

Input gzip SHA-256:
`32e572c27c6dcaab897d70e00506e73bcd73230032aa66755e14f03e28f52cd9`.
Startup golden gzip SHA-256:
`f7e863caa08ddcc57f5bafc7f345087fcb5f85b09bc5a2839bbc5e2e93559086`.
Startup decoded SHA-256:
`61bf4dd194efdf884b7dbc62861cd8b8d58590f35c7b48d20d13c2ab72cb806f`.
Nearest golden gzip SHA-256:
`1d66844f650df233e5f1f89308438757027d249604caaae2c0f8ad0dab9543dd`.

The frozen `_Generic` getter supports legacy-int and typed-native void C
entry ABIs without a later edit. The consumer compares actual C and direct
Go paths, ordered service states, final records and every native allocation
byte, preserving untouched high uintptr words and fault prefixes. Engine
changes must not edit these references or expectations. Full production
Push/Pop composition is a separate test, not substituted-service coverage.
This reference supplies no campaign, audio or overall 100% completion credit.
