This is a local copy of `github.com/opennox/go-openal` at commit
`164a70f24e7c` (BSD-licensed; see `LICENSE`). It is kept here to fix
macOS default-device handling in the cgo binding without changing Linux
audio behavior. On macOS it links only OpenAL Soft via pkg-config, avoiding
the mixed OpenAL Soft/Apple-framework linkage in the original binding.
