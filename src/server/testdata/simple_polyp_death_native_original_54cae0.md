# Original ImpEgg, Polyp and Potion raw C references

The independent input generator and original producer are frozen before the
raw C entries change. GAME.EXE SHA256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
There are 149 contexts per PC53/to-zero (`0xe3f`) and PC53/nearest (`0x23f`)
profile. The complete ImpEgg (33), Polyp (150), Potion (31) and actual
`00419A70` FISTP converter (17) execute all 231 instruction bytes in each
profile. Alignment NOPs are excluded. Polyp's 150 bytes and the converter's
17 bytes have prior complete seals; only 64 body bytes lack an older seal.
This new raw-memory/reentry reference does not relabel those old seals.

Each profile records 31 reentrant original calls, 110 real converter calls
and 15 memory faults: three flags, two source-X, four source-Y, two update
loads and four duration stores. All 64 KiB heap and loaded-image bytes,
caller guards, valid return stack/nonvolatile registers and FPU control/TOP
are checked. There are 73 different complete outputs between profiles.
Only six named world services are replaced: name lookup, numeric object
allocation, CreateAt, balance-record lookup, audio and delayed deletion.
The balance service returns the declared original binary32 record through
x87. Multiplication, signed DWORD FPS, binary32 spill and FISTP execute in
the original code; no FP arithmetic or conversion is substituted.

ImpEgg reloads and returns the complete flags DWORD after audio, setting
only bit `0x40`. Polyp retains the live shared type cache and constructor
return, reads source Y then X before caching the cloud update pointer, and
retains that pointer across CreateAt/balance callbacks and reentry. Failed
allocation skips source/update reads while preserving audio then deletion.
Potion forwards the source to audio then deletion without reading it.
Inputs include nil/unmapped pointers, a source whose Y is mapped but X is
unmapped, a partial flags-only mapping, source/cloud/update aliases, live
cache/FPS/position/update changes, signed FPS extremes, exceptional
lifetimes, and same/cross-entry reentry. World services are deliberately
opaque: accepting invalid arguments there supplies no transitive service
safety or full-game behavior certificate.

Reproduce the generator with `--output NEW_PATH`. Run the producer with
`--exe GAME.EXE --control 0xe3f --output NEW_PATH --proof NEW_PATH`, then
repeat with `0x23f`. Input gzip OS byte is `0x13`, output OS byte `0xff`,
mtime zero; compressed and decoded bytes are both pinned. All outputs use
exclusive creation. Neither program reads native implementation or native
expected output. This reference supplies no native port, original C fault
positive control, product/E2E, audible-fidelity or campaign completion credit.
