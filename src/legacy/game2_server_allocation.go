package legacy

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

// Native-pointer fixtures can scope the allocation service around a complete
// public/reference call. Nil keeps the ordinary product CRT ownership policy.
var game2ServerAllocationService func(uintptr, uintptr) unsafe.Pointer
var game2ServerReleaseService func(unsafe.Pointer)

func g2ServerAllocate(count, size uintptr) unsafe.Pointer {
	if game2ServerAllocationService != nil {
		return game2ServerAllocationService(count, size)
	}
	return C.calloc(C.size_t(count), C.size_t(size))
}

func g2ServerRelease(p unsafe.Pointer) {
	if game2ServerReleaseService != nil {
		game2ServerReleaseService(p)
	} else {
		C.free(p)
	}
}
