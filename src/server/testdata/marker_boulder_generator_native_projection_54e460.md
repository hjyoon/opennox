# Native transport and complete consumers before three raw engines

The initial native 64-KiB heap fixture was frozen at `4a66bcefa`, following
the independent original initial-memory producer/data at `4cd52e87d`.
All three original C entry arguments are still integers at this preparation.

`generate_special_death_boundaries_54e460.py` hash-binds and executes the
unchanged original-only producer. It inserts observations before successful
world events and at the final return/fault, without replacing instructions,
original writes, arguments or world plans. Both profiles reproduce their
existing complete output bytes exactly and add 3,743 whole-memory observations
each. The original producer continues to verify the whole heap/image/helper
region, caller, stack, nonvolatile registers and FPU invariants. Invalid-store
records include the original attempted store; they are not native write
permissions or native instruction/bus-trace credit.

The complete native consumer checks the ordered world trace with exact JSON
integers, all final world scalars/marker slots, actual fault kinds/sites and
every complete heap projection at its corresponding original observation.
Only exact observed original write destinations may change native bytes;
world mutation plans update the guarded expected memory separately. This
preserves write prefixes even when a later callback overwrites a marker slot.
Reentrant plans invoke the same native Go method or real C symbol, rather
than substituting original outputs for their execution.

Native image fields have separate C-owned, fully guarded allocations: the
packed names are widened into native pointer slots, the mass byte/divisor
region preserves its actual overlap, and the shared index/FPS/frame stay
DWORDs. The other native heap fields retain their previously frozen explicit
correspondences. Unobserved collision bytes remain guarded, without an
arbitrary cross-structure alias certificate.

The immutable original PE image is identified by SHA-256
`66eb0df5e37c97beb6c9b68fafa5a1d966cd4271347bd6c3453215684ff1a1c5`.
The independently recorded five mutable regions form a case-independent
patch-to-full-image digest certificate. The native consumer must project its
actual bytes onto a certified patch and independently match every region;
unknown patches fail. This compositional image projection is distinct from
hashing unrelated native process/code memory as a PE32 image. No original
executable image copy is added to the repository.

Before engine changes, the complete consumer compiles through a dynamic
interface and explicitly reports absent-engine branches without leaf credit.
An additional 400-case transport test checks every profile/input against its
first original observation and all native guards. The public C ABI and absent
native method are checked by a separate preparation gate. Actual original
worlds include 34 reentries and sixteen memory/two division faults per
profile; their full native execution remains a subsequent gate.

Reproduce each profile with the original executable and an output directory:

```sh
python generate_special_death_boundaries_54e460.py \
  --exe /path/to/GAME.EXE --control 0xe3f \
  --original-output original.json.gz --original-proof original.proof.json \
  --output boundaries.json.gz --proof boundaries.proof.json
```

Use `0x23f` for nearest rounding. Unicorn JIT needs the host's execution
permission; the initial restricted-host failure is retained separately.
