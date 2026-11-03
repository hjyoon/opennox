package legacy

import (
	"encoding/binary"
	"unsafe"
)

// These fixtures overlay native records on C-owned bytes filled with guard
// words. Preserve their pointer bits without asking Go's write barrier to
// trace an arbitrary previous word as a Go heap reference. Values are C
// addresses (or declared fault addresses); this never stores a Go-owned object.
func storeWallFixturePointerBits(buf []byte, at int, value unsafe.Pointer) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		binary.NativeEndian.PutUint32(buf[at:], uint32(uintptr(value)))
	} else {
		binary.NativeEndian.PutUint64(buf[at:], uint64(uintptr(value)))
	}
}
