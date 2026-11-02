# Original Chakram retarget public-entry contract

The complete GAME.EXE `004EB250` driver (226 instruction bytes) and
`004EB340` callback (150 bytes) execute without substituting instructions,
arithmetic, FPU comparisons, selection or velocity results. Their fourteen
and ten alignment bytes are excluded. The original image SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The already manifested constants at `00583C10` (binary64 widening of
binary32 0.1) and `00583C18` (binary32 400/160000) remain original bytes.

The independent producer generates 272 inputs without consulting results.
Each PC53 FPU profile executes all 376 instruction bytes on valid returns,
plus nine original faults. Startup toward-zero `0x0e3f` and nearest `0x023f`
differ in 112 complete outputs; both goldens are retained. The three declared
world services are rectangle enumeration `00517C10`, buff lookup `004FF350`
and map visibility `00537110`. They preserve the C calling convention and
may mutate declared records/scratch or reenter the original driver. All
nested driver/callback instructions execute as original machine code.

Vectors exercise integer class/flag masks, visibility and invisibility,
cached exclusions/update data, live coordinates and speed, point aliasing,
shared globals across reentrancy, live world mutations, equal/range gates,
finite rounding, subnormals, zero, infinities/NaNs and original dereference
fault prefixes. They are a declared sample, not exhaustive gameplay proof.
Whole original image, all 64 KiB of heap, five scratch words and guarded
caller storage are checked. Valid returns preserve nonvolatile registers
and FPU CW/TOP. The driver legitimately overwrites its first argument slot
with the binary32 Y spill; this exact declared slot is allowed, while its
return address and remaining caller storage remain guarded. An initial
preparation attempt rejected that valid spill; its source and failure log
remain under `ROOT/chakram-retarget-native-entry-port`.

The native comparator freezes both public-C and direct-Go contracts before
engine changes. It checks full ordered service/callback/nested state
prefixes, native pointer identity, original return/fault boundaries, every
tracked field bit and all 64 KiB of native records/canaries. It does not
claim instruction-by-instruction observation of native arithmetic. The
callback EAX is unobserved; the driver result remains a native pointer.
A runtime override supplies only declared world-service behavior and FPU
profile, never a selected target or computed velocity as a replacement.

Actual old C before evidence uses complete native records above 4 GiB,
independent C/native layout and owner/update/LastHit chain controls, real
rectangle/buff/map operations and the existing typed C visibility service.
The driver SIGSEGVs at its narrowed source. The callback instead loses valid
integer class bits via a numeric float-to-uint32 cast and rejects a valid
target before the narrowed point is reached. That is a semantic failure,
not a claimed callback SIGSEGV. The executed before test binary, original
source hashes and both child logs are retained.

Canonical toward-zero gzip SHA-256:
`e03c702b1cbe7aaaec0c71746bc48376bcb8fdfb30ce50c8ffb8d579cbd332eb`.
Canonical decoded SHA-256:
`ee3ad3ef672f9c805230c5b81940e8d27c6eddaa50b81d30de2a5ed8025e0a31`.
Original execution and pre-change reference preparation do not imply a
completed raw C port, transitive world-service proof, campaign completion
or audible/product coverage.
