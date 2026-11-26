package cfixture

/*
#include "memmap.h"

static void memmap_native_accessors(uintptr_t base, uintptr_t off, uintptr_t* out) {
	out[0] = (uintptr_t)mem_getPtr(base, off);
	out[1] = (uintptr_t)mem_getPtrSize(base, off, 8);
	out[2] = (uintptr_t)mem_getU8Ptr(base, off);
	out[3] = (uintptr_t)mem_getI8Ptr(base, off);
	out[4] = (uintptr_t)mem_getU16Ptr(base, off);
	out[5] = (uintptr_t)mem_getI16Ptr(base, off);
	out[6] = (uintptr_t)mem_getU32Ptr(base, off);
	out[7] = (uintptr_t)mem_getI32Ptr(base, off);
	out[8] = (uintptr_t)mem_getU64Ptr(base, off);
	out[9] = (uintptr_t)mem_getI64Ptr(base, off);
	out[10] = (uintptr_t)mem_getFloatPtr(base, off);
	out[11] = (uintptr_t)mem_getDoublePtr(base, off);
}
*/
import "C"

import "unsafe"

func MemmapNativeBlobStorage(base uintptr) (unsafe.Pointer, uintptr) {
	var size C.size_t
	ptr := C.nox_legacy_memory_blob(C.uintptr_t(base), &size)
	return unsafe.Pointer(ptr), uintptr(size)
}

func MemmapNativeAccessors(base, off uintptr) (out [12]uintptr) {
	C.memmap_native_accessors(C.uintptr_t(base), C.uintptr_t(off), (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return out
}
