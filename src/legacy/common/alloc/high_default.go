//go:build !(windows && amd64) && !(linux && (amd64 || arm64))

package alloc

import "unsafe"

func mallocHighMapping(size uintptr) (unsafe.Pointer, func()) { return nil, nil }
