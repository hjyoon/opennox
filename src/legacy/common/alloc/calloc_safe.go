//go:build safe

package alloc

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

// CallocC keeps the CRT's allocation result, including a freeable zero-size
// owner. Register successful allocations so the safe free uses the same owner.
func CallocC(count, size uintptr) unsafe.Pointer {
	ptr := C.calloc(C.size_t(count), C.size_t(size))
	if ptr != nil {
		allocMu.Lock()
		allocs[ptr] = count * size
		allocMu.Unlock()
	}
	return ptr
}
