# Original Skeleton death/debris/drop C reference

The independent generator and producer are frozen before engine changes.
GAME.EXE SHA256 is `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The four complete bodies are 54A310 (124 bytes), 54A390 (297), 54A4C0
(644), and 54A750 (124): 1,189 bytes, excluding 27 bytes of alignment NOPs.
The manifest seals each complete body separately. No original assets change.

Both PC53/to-zero (0xe3f) and PC53/nearest (0x23f) execute every instruction
across 169 independent inputs. Each has seven actual memory-access failures
and three actual CPU divide errors. Three complete outputs differ between
profiles. Only the sixteen named external world services are replaced;
nested calls execute the original four bodies. Fresh Unicorn instances keep
prior CPU exceptions from contaminating later cases. Full PE image and 64 KiB
heap, caller guard, valid-return stack/nonvolatile registers, FPU control and
valid TOP are checked. The original debris function reuses its own argument
stack word for the part count; only that precise caller write is allowed.

Cases cover coop admission, allocation failure before source access, object
aliases, invalid native roles, live position/class/use/cache/FPS/divisor
mutations, both drop branches, all four modifier-name lookups including nil,
all-descriptor-nil application, charge-byte truncation and invalid use data,
failed part allocation, signed modulo including divide overflow, floating
spills, and seventeen reentrant world boundaries. Integer RNG/float services
retain original min/max arguments; debug source strings/line numbers are
caller metadata and not native ABI arguments.

The original modifier helper initializes four pointer slots only. Its callee
4E4990 copies a fifth DWORD from uninitialized stack storage. The original
producer observes both 0xcccccccc and a reused part-count value (2); neither
is a defined modifier parameter. That fifth word is explicitly excluded from
semantic equality; the native port initializes Field16 to zero. This repairs
an uninitialized read and does not claim byte identity for undefined stack
contents. The callee itself is an external world boundary here; only its four
initialized modifier pointers are compared.

Reproduce with the generator's --output NEW_PATH, then the producer's --exe
GAME.EXE --control 0xe3f (and 0x23f) --output NEW_PATH --proof NEW_PATH. Outputs
are exclusive creations, gzip mtime is zero, input OS byte is 0x13 and output
OS byte is 0xff. Compare the complete compressed and decoded bytes. Producer
and generator read no native implementation or native expected results.

This reference alone does not prove native C execution, default-world
composition, actual audible fidelity, or any campaign cell's completion.
