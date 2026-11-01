# Original reactive monster helpers

Source executable SHA256:
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
No original asset or executable is included.

The reference executes the complete unchanged Sound `545DA0` (179 bytes)
and LookDamager `5466B0` (62 bytes), with the actual status-bit predicate
`5347A0` (23 bytes) and RNG `415FA0` (67 bytes). All 331 instruction bytes
execute. Two startup FPU controls, `0xe3f` and `0x23f`, produce identical
2,836 rows, including four raw nil faults. The two-byte alignment padding
after LookDamager is excluded. Registers, stack boundaries, FPU CW/TOP,
code/RNG-table bytes and guarded heap changes are verified independently.

Only Push is a declared world service. Frame/FPS reads are passive
observations at actual original load instructions. Real RNG execution
records signed DWORD bounds, returned bits and index changes. Independent
inputs cover marker/age/FPS wrap, all Push failure masks, seeds 0/1/4094/4095,
class/status bits, raw NaN/Inf/zero coordinates, reused independent and
embedded action records, and after-Push mutation of update pointers,
coordinates, status and world clocks. Sound retains the entry update
pointer; LookDamager reads live coordinates after Push. Raw nil faults
retain the preceding mutation and service prefixes.

The compressed rows and native consumer were frozen before reactive engine
edits. Gzip SHA256:
`21f8a0a310bf64bf68da1560facc6d3b5370cc9057ea8a3b5c79ec5373f01a9a`.
Decoded SHA256:
`d503d68bda0d36bde27074c7fe7cee3c14d6156940017bfd84aaebf89737db77`.
The native consumer checks public C and Go paths against the same original
rows, independent C/Go layouts, complete Cmalloc records/padding/neighbors,
native-width pointer projections, actual production prand results/index,
ordered services and raw nil faults. Expected words come from the original
execution; none come from the new implementation.

Private producers, inputs, both original outputs, instruction traces,
pre-edit public C SIGSEGVs and seals remain under
`ROOT/monster-react-sound-damager-native-entry-port`. The pre-edit native
faults occur after complete valid above4GiB layout and real predicate
controls pass. Earlier Guard/RNG/predicate manifest records remain unchanged;
only the previously unsealed 62-byte LookDamager range is added.
This reference does not establish full transitive game equivalence,
speaker output, campaign completion or overall 100% porting.
