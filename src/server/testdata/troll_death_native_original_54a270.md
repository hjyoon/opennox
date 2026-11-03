# Original Troll death C-entry reference

This freezes the original before changing the raw public C entry. No native
implementation or expected native output is used by either producer.

The original GAME.EXE SHA-256 is
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The complete `54A270` body is 154 bytes, SHA-256
`c21a25775cdac2ec64520866b21071063b02b89aaa5818ebfb511a685cc3f79e`.
The original `419A70` converter is 17 bytes, SHA-256
`10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186`.
Both ranges are already sealed in the code manifest; no duplicate is added.
Six bytes of padding after the complete callback are excluded.

The independent input generator reuses only input columns of the thirty
existing historical cases, plus cache/resolve results, allocation failure,
invalid source/cloud/update pointers, aliases, world mutations, binary64
lifetime and signed FPS boundaries, and five original reentrant executions.
All 136 cases execute with PC53/to-zero (`0xe3f`) and PC53/nearest (`0x23f`).
Each profile executes all 171 original instruction bytes and records eleven
original memory faults: six source-Y reads, one cloud-update read and four
final cached-update stores. Forty-nine complete outputs differ between profiles.

Only resolve `4E3AA0`, constructor `4E3450`, create `4DAA50`, audio `501960`
and binary64 balance lookup `419D40` are declared world services. FPS and the
cached type index are read/written by the original instructions themselves.
Reentrant calls execute the original body and converter. The entire image,
64 KiB heap, guarded caller, valid-return stack/nonvolatile registers, and
FPU control/TOP are checked. Original writes outside the cached type scalar
or the two declared duration records are rejected. Faults remain original
instruction accesses; no synthetic fault result replaces execution.

The constructor result may be nil, another valid object, an alias of the
source, or a declared bad pointer. The update pointer is cached before
source position reads and world effects. Later replacements do not redirect
the final store. Allocation failure returns one without reading the source.
Raw nil/update faults remain distinct from established ordinary admission
and its nil-update store guard.

Generate inputs with `--output NEW_PATH`; execute the original producer with
`--exe GAME.EXE --control 0xe3f --output NEW_PATH --proof NEW_PATH`, and repeat
with `0x23f`. Outputs are created exclusively. Compare bytes to the frozen
files. Gzip metadata is explicit: input OS byte `0x13`, original output OS
byte `0xff`, mtime zero. This retains the initial canonical bytes across
Python versions. Historical failed first-return-only disassembly and gzip
metadata assumptions remain recorded under ROOT/troll-death-native-entry-port.

This reference does not prove a native C entry port, actual default-world
composition, original audible fidelity, or campaign completion.
