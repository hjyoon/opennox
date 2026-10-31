# Original dead-target callback reference

Both gzip files execute the unchanged GAME.EXE callback at `0x00532040`,
SHA256 `0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Startup control word `0x0e3f` and independent nearest `0x023f` each run 10,800
inputs. Their input/output rows are identical. No external service or
predicate, pointer, write or return outcome is substituted.

All 42 original body bytes execute; six trailing NOP bytes belong to the
48-byte sealed interval. The callback reads the candidate ID, the explicit
context's ID and, on a match, the candidate DEAD flag. A matching dead
candidate sets the global DWORD at `0x00833C1C` to 1. Other cases retain its
existing value, including noncanonical values. Every case returns the
candidate pointer. The candidate's UpdateData link deliberately points to
a different record with a different target ID.

Inputs cover six candidate IDs, six context IDs, ten flag words, six
initial global words and five classes. Full heap, stack, global neighbors,
original code, nonvolatile registers and FPU state are guarded. The global
is in the PE32 virtual BSS rather than file-backed data; its address is
sealed in the original instructions, and its contents are explicit inputs.
The file-backed data verifier is unchanged.

The public C comparator and both references were frozen before the engine
change. It checks the native candidate return and entire candidate,
explicit-context, alternate-update and global guard regions. The same
native allocations pass a Go push control before the old C entry crashes
at candidate.low32+36. Raw crash and initial preparation failures remain
preserved outside the repository. The first producer version had the
wrong Unicorn write-hook signature; the executed version only corrects
that signature and retains every planned input.

Reproduce with Capstone and Unicorn, using fresh output paths:

```sh
python3 original_monster_dead_target_native_532040.py \
  --game /path/to/GAME.EXE --control 0xe3f \
  --fixture /new/output.json.gz --proof /new/proof.json
```

Use `0x23f` for the independent nearest control. These leaf checks do not
certify the retained PE32 FIGHT parent or overall porting 100%.
