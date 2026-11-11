//go:build !windows || !amd64

package alloc

import "unsafe"

func mallocHighMapping(size uintptr) (unsafe.Pointer, func()) { return nil, nil }
