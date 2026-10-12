package handles

import (
	"sync/atomic"
	"unsafe"
)

// FrontierPtr is the last issued key in this monotonically allocated arena.
// It neither allocates nor dereferences the protected opaque handle memory.
// Like NewPtr, it is used only between Init and Release on the owning thread.
func FrontierPtr() unsafe.Pointer {
	off := uintptr(atomic.LoadUint32(&cur))
	if base == 0 || off == 0 {
		return nil
	}
	return unsafe.Pointer(&data[off])
}
