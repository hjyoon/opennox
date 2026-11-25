package legacy

/*
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func allocateDrawMemory(count int, size uintptr, high bool) unsafe.Pointer {
	if high && count > 0 && size != 0 && uintptr(count) <= ^uintptr(0)/size {
		p, _ := alloc.CallocHigh(count, size)
		return p
	}
	return C.calloc(C.size_t(count), C.size_t(size))
}

func freeDrawMemory(ptr unsafe.Pointer) {
	if !alloc.TryFreePtr(ptr) {
		C.free(ptr)
	}
}
