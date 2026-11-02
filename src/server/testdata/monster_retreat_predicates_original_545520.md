# Original retreat predicate native-entry reference

The reference is frozen before changing any engine body or public C ABI, at
`616b61d7b3d94c3871d1b5b3721fa90005e87b27`. Producer, independent inputs,
full original runs, guard proofs and failed preparations are retained under
`ROOT/monster-retreat-predicates-native-entry-port`.

The untouched `GAME.EXE` has SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Execution includes the complete `545520` (83 bytes), `545580` (40), `5455B0`
(43), and real `4FF350` buff getter (35): **201 instruction bytes**, plus
four constant-data bytes at `583068`. All instruction addresses execute on
valid inputs. The 39 alignment bytes receive no instruction credit. These
ranges are already covered by the existing oracle manifest; no duplicate
range or coverage increment is required.

Each PC53/ToZero (`0x0e3f`) and PC53/nearest (`0x023f`) run executes 2,073
independently generated root contexts and one ancillary null-buff context.
The two complete outputs are byte-identical. Nineteen root contexts fault;
null health is ignored by the cast predicate, while null update is not.
The parent executes the actual original health, cast and buff bodies with
their original short circuit. There are no substituted instructions,
arithmetic operations, service results, or world-service boundaries.

Inputs cover zero maximum, full uint16 health values, adjacent binary32
thresholds, signed zero, subnormals, infinities and both kinds/signs of NaN,
all original status/buff gates, null records, health/object aliases, and
independent seeded bit patterns. Whole heap and original image bytes,
guarded caller bytes, valid-return nonvolatile registers, CW and TOP are
checked. `545544`/`54554D` deliberately overwrite a by-value caller argument
with Cur/Max; only those observed DWORD stores are admitted.

Frozen input gzip SHA-256:
`523f59cb9b2f5d3fc24ac1c32d5750189bd75f6e42663012e0838b067a04a3a8`.
Golden gzip SHA-256:
`3c8ca86a5bb03ca5ac48a495c98b534fa32cf2d2670db84e1ce87e58c1e175d7`.
Golden decoded SHA-256:
`8d721b4b0a416310851bc2feae147e8a28ed07eba1aebe02b7e6224bc3788db7`.
The consumer validates both golden hashes and compares each raw public C
and direct-Go entry, including fault admission and the complete native
allocation. The separately frozen `_Generic` C getter supports the original
legacy-int ABI and the subsequent native-pointer ABI without being edited.

The raw `5456D0` C update remains a separate unported entry. The current
runtime dispatches its ordinary native Go update instead, and no direct
active C or Go reference to that raw symbol is found outside its definition
and declaration. That static classification is not a transitive closure or
runtime reachability certificate. This reference does not establish
campaign completion, audible audio output, or overall 100% porting.
