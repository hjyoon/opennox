# Pixel references for the existing PNG goldens

`png_golden_pixels.json` preserves the pixels identified by the existing
276 PNG MD5 expectations: six particle cases and 270 sprite/blend cases,
with 225 distinct hashes. The existing MD5 strings are unchanged.

Go 1.27 changes DEFLATE output, so comparing compressed PNG bytes no longer
tests the renderer consistently. Before changing those comparisons, a
temporary copy of the historical PNG/DEFLATE encoding code, compiled with
Go 1.27.3, reproduces all 276 existing MD5 expectations exactly. Each verified
PNG is then decoded with the standard library. The table records its width,
height and SHA-256 of every pixel's `RGBA()` channels as big-endian uint16s
in row order. The MD5 of the original encoded PNG is the table key.

This derives no replacement golden from a newly accepted renderer output:
every input PNG first matches its pre-existing encoded-byte expectation.
The same retained test main records the standard encoder failures and all
historical encoder successes. Production raster/color code is unchanged.
Original encoded PNGs, temporary encoder sources, provenance, both actual
producer mains and complete results are retained in the port artifacts;
the temporary encoder is not part of the final test implementation.

Evidence directory: `go1273-toolchain-transition` under the retained
`opennox-port-20261105` artifacts. See
`Go1273-historical-PNG-existing-hash-reference-qualified-v1.json` and
`PNG-existing-goldens-pixel-reference-frozen-v1.json`.
