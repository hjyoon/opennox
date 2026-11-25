package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Sub_4EC5B0 resets the legacy respawn registry through the portable ordered
// contract while keeping both pointer-valued globals at the native C width.
func Sub_4EC5B0() {
	server.RespawnReset4EC5B0(server.RespawnResetHooks4EC5B0[unsafe.Pointer]{
		LoadAllocator: respawnAllocatorLoad4ECA60,
		ClearHead: func() {
			respawnAddStoreHead4EC5E0(nil)
		},
		FreeAll: func(allocator unsafe.Pointer) {
			alloc.AsClass(allocator).FreeAllObjects()
		},
		Enable: func() {
			respawnAddStoreAllow4EC5E0(1)
		},
	})
}
