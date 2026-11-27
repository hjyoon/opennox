package legacy

import (
	"runtime"
	"unsafe"
)

// nativePointerFromWord decodes a full-width address passed as a native word.
// The native caller owns the referent for the duration of the call.
func nativePointerFromWord(word uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&word))
}

// loadUnalignedNativePointer reads a pointer from a packed foreign slot.
func loadUnalignedNativePointer(address unsafe.Pointer) unsafe.Pointer {
	var value unsafe.Pointer
	width := unsafe.Sizeof(value)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&value)), width), unsafe.Slice((*byte)(address), width))
	return value
}

// storeForeignPointer copies exactly one pointer representation to memory
// owned by the native allocator. The previous bytes can be arbitrary.
func storeForeignPointer(address, value unsafe.Pointer) {
	width := unsafe.Sizeof(value)
	copy(unsafe.Slice((*byte)(address), width), unsafe.Slice((*byte)(unsafe.Pointer(&value)), width))
	runtime.KeepAlive(value)
}
