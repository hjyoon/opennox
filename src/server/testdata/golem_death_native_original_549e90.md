# Original MechGolem/Golem death C reference

Generator and producer are frozen before either raw C entry is changed.
GAME.EXE SHA256 is `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
The complete MechGolem body at 549E90 is 265 bytes; Golem at 549FA0 is
426 bytes. All 691 instruction bytes execute in each PC53/to-zero (0xe3f)
and PC53/nearest (0x23f) profile across 164 independent inputs. Thirteen
alignment NOP bytes are excluded. Six complete outputs differ by profile.

Each profile records ten memory failures (four origin, two object, one
name, two mass-table and one rock-table) and two actual CPU divide errors.
Twenty-two reentrant calls execute the original bodies, with ordered
external-service argument/results and completed fault prefixes retained.
Full PE image and 64 KiB heap, caller stack guard, valid-return stack and
nonvolatile registers, FPU control word and valid TOP are checked. Golem
reuses its incoming argument stack word for the part count; only its two
specific original count stores are allowed there.

Only eleven named external world services are replaced. The producer reads
the original packed name tables, byte masses, adjacent signed divisor and
debug-data word before the rock table from the original image. Inputs may
stop the Mech table early or mutate declared tables, mass bytes, positions,
flags, shared cache, divisor, FPS and Coop state at world boundaries.
Names and the iteration cursor are cached where the original caches them;
mass/cache/flags/divisor and mode are reloaded at their original sites.
RNG debug strings/line numbers select the part-count draw without adding a
draw. Floating service results are binary64; the original performs each
binary32 spill under the selected control word.

Cases include empty tables/counts, allocation short circuits before origin
reads, invalid sources/results, object aliases, unsigned masses 128/255,
live part-table and mass changes, both decay ranges, DWORD multiply/wrap,
signed remainder/divide overflow, exceptional floating values, cached
source identity with live positions, and cross/same-entry reentry.

The two malformed indices 0xffffffff and 0x7fffffff project to the DWORD
before the rock table. Its value 0x00632e63 points into the zero-filled
original loaded image, so the original constructor service receives an
empty string and returns nil in this fixture. Native controlled table
fixtures represent that observed empty name explicitly. These projections
do not certify arbitrary out-of-table named-field aliases in the native
packed-data adapter. The cache is an extracted C global, separate from its
retired blob slot; byte masses and the following divisor remain packed.

Reproduce with generator `--output NEW_PATH`, then producer `--exe GAME.EXE
--control 0xe3f` (and 0x23f) `--output NEW_PATH --proof NEW_PATH`. All result
files are exclusive creations. gzip mtime is zero, input OS byte is 0x13
and output OS byte 0xff. Compare compressed and decoded bytes. No native
implementation or native expected output is read by either program.

This reference alone supplies no native port, default-world execution,
natural death, audible-fidelity or campaign completion credit.
