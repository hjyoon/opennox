package legacy

import (
	"runtime"
	"unsafe"
)

// Original00409B50 assumes the cdecl DF=0 contract. It scans with ECX initially
// FFFFFFFF, then copies forward in DWORD units and finally BYTE units. These
// ordered loads/stores deliberately retain destructive overlap semantics.
func sub409B50Native(source unsafe.Pointer) uint32 {
	remaining := ^uint32(0)
	var offset uintptr
	for remaining != 0 {
		value := *(*byte)(unsafe.Add(source, offset))
		remaining--
		if value == 0 || remaining == 0 {
			break
		}
		offset++
	}
	count := ^remaining
	// Resolve the shared destination only after the entire source scan. The
	// getter returns the live C-owned alias, without allocating or inferring
	// a string capacity. No nil guard or safety truncation is introduced.
	destination := sub409B80Native()
	for i := uint32(0); i < count>>2; i++ {
		off := uintptr(i) * 4
		value := *(*uint32)(unsafe.Add(source, off))
		*(*uint32)(unsafe.Add(destination, off)) = value
	}
	tail := uintptr(count &^ uint32(3))
	for i := uint32(0); i < count&3; i++ {
		off := tail + uintptr(i)
		value := *(*byte)(unsafe.Add(source, off))
		*(*byte)(unsafe.Add(destination, off)) = value
	}
	// Offsets are integers, never converted pointer values. Every unsafe.Add
	// above is immediately needed by a real access; there is no dead one-past
	// pointer at the terminal NUL of an exactly sized Go source allocation.
	runtime.KeepAlive(source)
	return count
}
