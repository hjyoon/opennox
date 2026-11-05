# Original raw Marker, Boulder, MonsterGenerator and destruction credit

These references execute GAME.EXE with SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Inputs and expected outputs are independent of native implementations. The
four actual C entry bodies still use PE32 pointer arguments; their existing
ordinary Go callbacks do not qualify these raw C ABIs.

| Body | Bytes | Inputs per FPU profile |
| --- | ---: | ---: |
| Marker `54E460` | 77 | shared 200 |
| Boulder `54E4B0` | 366 | shared 200 |
| MonsterGenerator `54E630` | 178 | shared 200 |
| Destruction credit `4D61B0` | 58 | 286 |

All 679 instruction bytes execute under PC53 toward-zero (`0x0E3F`) and
nearest (`0x023F`) profiles. Alignment bytes after the final RET are excluded.
Two existing five-byte call-site seals remain unchanged and receive no new
byte credit. The first three bodies execute 34 real original reentries and
16 memory/two signed-division faults in each profile. Three complete outputs
differ by rounding profile. Their world services are declared in the proof;
original floating-point instructions are never substituted.

The credit helper executes without world-service substitutions. Its two
profile outputs are identical. It tests all 256 low flag bytes, counter wrap,
invalid unit/update/player pointers and four declared PE32 aliases. The
original caches the update pointer but reloads its player pointer after
incrementing the player counter. Additional class or nil-data guards would
change this raw contract.

The references check the complete 64-KiB heap, mapped image, helper region,
caller bytes, FPU control and valid-return stack/nonvolatile registers/FPU
TOP. Fault outputs preserve exact observed sites and prior writes. Point FX
spies observe pointers without executing the callee or reading coordinates.
The script boundary receives three original words; event 3 is an explicit
native-service projection, not a fourth original argument.

Six new file-backed data ranges seal 423 bytes: mass bytes, signed divisor,
name pointer table, pointed strings, RNG source paths and DestroyedGenerator.
The index `834ACC`, frame `84EA04` and FPS `85B3FC` are mutable virtual-zero
globals and are explicitly initialized, never fabricated as file-backed data.
Invalid-index and alias transport does not certify arbitrary native aliases.

The sibling producers and input/output gzip files retain exact original
hashes in their proof JSON. Both input generators and both complete original
references reproduce byte for byte before native implementation. Reproduce
using the pinned Unicorn/Capstone environment and exclusive output paths:

```sh
python generate_independent_inputs_54e460.py --output /private/tmp/death-inputs.json.gz
python generate_credit_inputs_4d61b0.py --output /private/tmp/credit-inputs.json.gz
python original_special_death_54e460.py --exe "$NOX_ORIGINAL_DATA/GAME.EXE" --control 0xe3f --output /private/tmp/death-tozero.json.gz --proof /private/tmp/death-tozero-proof.json
python original_credit_4d61b0.py --exe "$NOX_ORIGINAL_DATA/GAME.EXE" --control 0xe3f --output /private/tmp/credit-tozero.json.gz --proof /private/tmp/credit-tozero-proof.json
```

Repeat with `--control 0x23f`. Input files are read beside the producers, so
compare fresh generated inputs with the pinned siblings first. No native
implementation, actual unchanged-C fault, product or campaign credit is
assigned by this original reference. Separate initial-memory/layout/service,
actual-C causality and frozen whole-output consumer gates precede engine work.
