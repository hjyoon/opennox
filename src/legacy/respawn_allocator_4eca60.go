package legacy

/*
extern void* nox_alloc_respawn_1568020;

static void* nox_respawn_allocator_load_4ECA60(void) {
	return nox_alloc_respawn_1568020;
}

static void nox_respawn_allocator_store_4ECA60(void* value) {
	nox_alloc_respawn_1568020 = value;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type respawnAllocatorRuntime4ECA60 struct {
	NewClass       func(name string, recordSize uintptr, capacity int) unsafe.Pointer
	StoreAllocator func(unsafe.Pointer)
}

func respawnAllocatorLoad4ECA60() unsafe.Pointer {
	return C.nox_respawn_allocator_load_4ECA60()
}

func respawnAllocatorStore4ECA60(value unsafe.Pointer) {
	C.nox_respawn_allocator_store_4ECA60(value)
}

func respawnAllocatorNative4ECA60(runtime respawnAllocatorRuntime4ECA60) int {
	ok := server.RespawnAllocator4ECA60(
		unsafe.Sizeof(respawnRecord4EC5E0{}),
		server.RespawnAllocatorHooks4ECA60[unsafe.Pointer]{
			NewClass: runtime.NewClass,
			NonZero: func(allocator unsafe.Pointer) bool {
				return allocator != nil
			},
			StoreAllocator: runtime.StoreAllocator,
		},
	)
	return bool2int(ok)
}

// Nox_xxx_allocItemRespawnArray_4ECA60 creates the native-width respawn
// allocation class and stores its opaque handle in the legacy pointer global.
func Nox_xxx_allocItemRespawnArray_4ECA60() int {
	return respawnAllocatorNative4ECA60(respawnAllocatorRuntime4ECA60{
		NewClass: func(name string, recordSize uintptr, capacity int) unsafe.Pointer {
			return alloc.NewClass(name, recordSize, capacity).UPtr()
		},
		StoreAllocator: respawnAllocatorStore4ECA60,
	})
}
