# Original raw Skull and ArrowTrap entries, before native changes

These references execute the original PE32 instructions from GAME.EXE with
whole-file SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
They do not read a native implementation or derive expected results from it.
The five public C entries still use their old arguments. Their ordinary Go
dispatcher callbacks are separate; no new native-port or actual unchanged-C
runtime crash credit is assigned by these original references.

| Original body | Complete bytes | SHA-256 |
| --- | ---: | --- |
| `54F9A0` update | 215 | `b79f13c5294ff503c6c585b8c967c3e3b463019bd943a1e60a2065db0b47c862` |
| `54FA80` projectile | 300 | `c580ca147c534ab48e9186e4ad3f126f9a513a31bde673b0e45a29fdc8309bdd` |
| `54FBB0` state wrapper | 51 | `42356876554beba9c143631ba4fdc958744d9a73fe567ff437c1d75423e0d5e0` |
| `54FBF0` rectangle scan | 94 | `215b664efef388461519e9d1e18a74d13ec447b2a1b7d7574a05fc32e82cc9d8` |
| `54FC50` candidate callback | 91 | `e4f6e35e82ae0c21b86ac7fc68dc9c9cb89ee879116cdd01d4bb14a262421585` |

The five bodies total **751 bytes**. Intraprocedural CFG reachability stops
the last body at `54FCAB`; independent one-byte RET stubs at `54FCB0` and
`54FCC0` are excluded. The first 773-byte overcount and its correction are
retained under `ROOT/skull-update-native-entry-port`. The real 17-byte
`419A70` FISTP helper also executes; it already has a complete earlier seal
and supplies no new body-byte credit.

The independent generator supplies **753 inputs**: enabled-byte/state/timer
and unsigned wrap cases, cold/warm/zero/large cache results, signed direction
words, raw numeric projectile IDs, allocation and pointer failures, shared
record aliases, exact low-bit candidate gates, callback mutations and
reentries. It reads no native source. Its encoded input SHA-256 is
`2fa912b92071a20903b0596adce24a80942874efeb6e80fe5b8d8f6def6dcc36`;
decoded SHA-256 is
`8a004ff539f042e37bfa144916780ab9042d5a6ed48889229bb6d6e18bce8825`.

Both PC53 profiles execute every byte of the five bodies and converter:
startup-compatible rounding toward zero (`0x0E3F`) and nearest (`0x023F`).
Each includes **53 original reentrant calls, 402 original enumeration
callback calls, 276 real FISTP calls and 26 memory faults**. All complete
64-KiB heap, mapped image and caller guards pass. Successful returns preserve
stack position, nonvolatile registers and FPU stack/control. There are no
substitutions for original FP arithmetic or conversion instructions.

Only ten world services are replaced: numeric name lookup, constructor,
creation, balance, audio, trap FX, enumeration, enemy, map and direction
tests. Enumeration invokes the actual original `54FC50`; nested plans run
the actual requested original body. Every service boundary and nested
return records the complete heap digest, shared marker/caches, FPS, exact
arguments and result. Final outputs retain those records, result/fault and
complete heap digest. Private local scan state cannot replace the original
live shared marker: a later reentrant empty scan can clear an earlier hit.

| Profile | Encoded output SHA-256 | Decoded output SHA-256 |
| --- | --- | --- |
| Toward zero | `c5033d70e53a39874dd667354eb9ead31d970e1ca63ab73b3603b6334dedde93` | `f9269f877bc70d69e4defb25eecb5c200ad742df9d468d469fa152c4e06c288b` |
| Nearest | `08d9557293ef26581f612b6177b1df60f0e579addb9db47cf0c5b0e755d658de` | `a7515ca2c9beb823f28871bf1c3256db6cd9a625c8a7078f7b4276110dccf76b` |

Independent reproduction matches the complete compressed outputs, decoded
outputs and proof bytes. **303 complete outputs differ between profiles**.
Proofs and every preparation failure remain under
`ROOT/skull-update-native-entry-port`; the earlier Capstone FSTP access-flag
omission and incorrect controller interpreter path remain failed attempts.

The code oracle adds five complete bodies and ten file-backed data records,
totalling **104 new data bytes**. The existing binary32 `4.0` constant and
2048-byte direction table retain their earlier identities and seals. Three
file-backed out-of-table neighborhoods are separately bound, including an
explicit executable-section data record for the eight bytes read by signed
direction `-32768`. All older oracle records and their order stay unchanged.

This is a declared initialized-PE fixture. Signed direction `32767` reads
the mapped virtual-zero tail at `005F6650`; that tail and mutable globals
`00834B08..00834B1F` are not file-backed ranges and are not inserted as such
in the code oracle. The latter are explicitly initialized from the inputs.
Malformed-index and record-alias references do not certify arbitrary native
named-field aliasing or represent a running original campaign snapshot.
Actual C/Go layout and service controls, unchanged-C reproductions, native
consumers and implementation are subsequent gates.

Reproduce with the original asset and the pinned Unicorn/Capstone environment:

```sh
python generate_skull_raw_inputs_54f9a0.py --output /private/tmp/skull-input-reproduction.json.gz
python original_skull_raw_54f9a0.py --exe "$NOX_ORIGINAL_DATA/GAME.EXE" --control 0xe3f --output /private/tmp/skull-tozero.json.gz --proof /private/tmp/skull-tozero-proof.json
python original_skull_raw_54f9a0.py --exe "$NOX_ORIGINAL_DATA/GAME.EXE" --control 0x23f --output /private/tmp/skull-nearest.json.gz --proof /private/tmp/skull-nearest-proof.json
```

Output files use exclusive creation. The producer reads the pinned sibling
input fixture, so compare the independently generated file before running.
