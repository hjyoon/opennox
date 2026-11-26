package legacy

/*
#include "GAME4_2.h"

*/
import "C"

import "unsafe"

const mapgenBufferSize = 0x2000

var mapgenBuffer unsafe.Pointer

func mapgenAcquireBuffer(register func(unsafe.Pointer) uint32) (unsafe.Pointer, uint32) {
	buffer := nox_mapgen_alloc_memory(1, mapgenBufferSize)
	if buffer == nil {
		return nil, 0
	}
	token := register(buffer)
	if token == 0 {
		nox_mapgen_release_memory(buffer)
		return nil, 0
	}
	return buffer, token
}

//export nox_xxx_mapgenAllocBuffer_5213E0
func nox_xxx_mapgenAllocBuffer_5213E0() C.int {
	nox_xxx_mapgenFreeBuffer_521400()
	buffer, token := mapgenAcquireBuffer(func(ptr unsafe.Pointer) uint32 {
		return registerMapgenPointer(ptr)
	})
	mapgenBuffer = buffer
	legacyWords.dword_5d4594_2487556 = C.uint32_t(token)
	return C.int(bool2int(buffer != nil))
}

//export nox_xxx_mapgenFreeBuffer_521400
func nox_xxx_mapgenFreeBuffer_521400() {
	if mapgenBuffer != nil {
		forgetMapgenPointer(mapgenBuffer)
		nox_mapgen_release_memory(mapgenBuffer)
		mapgenBuffer = nil
	}
	legacyWords.dword_5d4594_2487556 = 0
}

//export nox_mapgenPrepareWorkspaceNative
func nox_mapgenPrepareWorkspaceNative(theme *C.uint8_t) C.int {
	if nox_xxx_mapgenAllocBuffer_5213E0() == 0 {
		sub_520F80()
		return 0
	}
	if sub_520EA0(theme) == 0 {
		nox_xxx_mapgenFreeBuffer_521400()
		return 0
	}
	return 1
}
