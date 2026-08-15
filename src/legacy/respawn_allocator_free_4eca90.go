package legacy

/*
extern void* nox_alloc_respawn_1568020;

static void* nox_respawn_allocator_free_load_4ECA90(void) {
	return nox_alloc_respawn_1568020;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type respawnAllocatorFreeRuntime4ECA90 struct {
	LoadAllocator func() unsafe.Pointer
	FreeClass     func(unsafe.Pointer)
}

func respawnAllocatorFreeNative4ECA90(runtime respawnAllocatorFreeRuntime4ECA90) {
	server.RespawnAllocatorFree4ECA90(server.RespawnAllocatorFreeHooks4ECA90[unsafe.Pointer]{
		LoadAllocator: runtime.LoadAllocator,
		FreeClass:     runtime.FreeClass,
	})
}

// Sub_4ECA90 frees the native-width respawn allocation class without clearing
// the legacy pointer global, matching the original session-shutdown boundary.
func Sub_4ECA90() {
	respawnAllocatorFreeNative4ECA90(respawnAllocatorFreeRuntime4ECA90{
		LoadAllocator: func() unsafe.Pointer {
			return C.nox_respawn_allocator_free_load_4ECA90()
		},
		FreeClass: func(allocator unsafe.Pointer) {
			alloc.AsClass(allocator).Free()
		},
	})
}
