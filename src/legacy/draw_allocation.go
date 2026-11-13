package legacy

/*
#include <stdlib.h>
#include "draw_allocation.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_draw_memory_calloc_high_go
func nox_draw_memory_calloc_high_go(count, size C.size_t) unsafe.Pointer {
	p, _ := alloc.CallocHigh(int(count), uintptr(size))
	return p
}

//export nox_draw_memory_release_go
func nox_draw_memory_release_go(ptr unsafe.Pointer) {
	if !alloc.TryFreePtr(ptr) {
		C.free(ptr)
	}
}
