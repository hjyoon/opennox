package legacy

/*
#include <stdint.h>
#include "defs.h"

extern void* nox_alloc_respawn_1568020;
extern nox_respawn_record_t* dword_5d4594_1568024;
extern uint32_t nox_xxx_respawnAllow_587000_205200;

static void* nox_respawn_reset_load_allocator_4EC5B0(void) {
	return nox_alloc_respawn_1568020;
}

static void nox_respawn_reset_clear_head_4EC5B0(void) {
	dword_5d4594_1568024 = NULL;
}

static void nox_respawn_reset_enable_4EC5B0(void) {
	nox_xxx_respawnAllow_587000_205200 = 1;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Sub_4EC5B0 resets the legacy respawn registry through the portable ordered
// contract while keeping both pointer-valued globals at the native C width.
func Sub_4EC5B0() {
	server.RespawnReset4EC5B0(server.RespawnResetHooks4EC5B0[unsafe.Pointer]{
		LoadAllocator: func() unsafe.Pointer {
			return C.nox_respawn_reset_load_allocator_4EC5B0()
		},
		ClearHead: func() {
			C.nox_respawn_reset_clear_head_4EC5B0()
		},
		FreeAll: func(allocator unsafe.Pointer) {
			alloc.AsClass(allocator).FreeAllObjects()
		},
		Enable: func() {
			C.nox_respawn_reset_enable_4EC5B0()
		},
	})
}
