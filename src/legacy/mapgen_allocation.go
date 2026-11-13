package legacy

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

// The native Mapgen graph and its transferred fixture records share the
// existing scoped GAME4 allocation owner. Without a fixture service these
// calls retain the real CRT calloc/free behavior.
//
//export nox_mapgen_alloc_memory
func nox_mapgen_alloc_memory(count, size C.size_t) unsafe.Pointer {
	c, s := uintptr(count), uintptr(size)
	// Preserve native calloc's zero/overflow result before a diagnostic high
	// allocator converts count to Go int. Such allocations carry no high
	// pointer witness and must retain their CRT owner.
	if c == 0 || s == 0 || c > uintptr(^uint(0)>>1) || c > ^uintptr(0)/s {
		return C.calloc(count, size)
	}
	return game4SupportAllocate(c, s)
}

//export nox_mapgen_release_memory
func nox_mapgen_release_memory(ptr unsafe.Pointer) {
	game4SupportReleaseMemory(ptr)
}

func mapgenHighFixtureMemory() func() {
	return game4SupportHighFixtureMemory()
}
