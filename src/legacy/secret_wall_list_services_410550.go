package legacy

/*
#include <stdint.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include "GAME1.h"
extern void* dword_5d4594_251560;

static void* secret_wall_native_head_cell_410550(void) { return &dword_5d4594_251560; }
static void* secret_wall_native_foreign_load_pointer_410550(const void* slot) {
    void* value;
    memcpy(&value, slot, sizeof(value));
    return value;
}
static uint16_t secret_wall_native_foreign_load_word_410550(const void* slot) {
    uint16_t value;
    memcpy(&value, slot, sizeof(value));
    return value;
}
static void secret_wall_native_foreign_store_pointer_410550(void* slot, void* value) {
    memcpy(slot, &value, sizeof(value));
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// This always names the actual production C owner; the four-byte historical
// memmap descriptor is not the storage used by the current C list helpers.
func secretWallNativeHeadCell410550() unsafe.Pointer {
	return C.secret_wall_native_head_cell_410550()
}

func secretWallNativeForeignLoadPointer410550(slot unsafe.Pointer) unsafe.Pointer {
	value := C.secret_wall_native_foreign_load_pointer_410550(slot)
	runtime.KeepAlive(slot)
	return value
}

func secretWallNativeForeignLoadWord410550(slot unsafe.Pointer) uint16 {
	value := uint16(C.secret_wall_native_foreign_load_word_410550(slot))
	runtime.KeepAlive(slot)
	return value
}

func secretWallNativeForeignStorePointer410550(slot, value unsafe.Pointer) {
	C.secret_wall_native_foreign_store_pointer_410550(slot, value)
	runtime.KeepAlive(slot)
	runtime.KeepAlive(value)
}

// These helper bodies remain external. A handled diagnostic reply is checked
// before this function; no returned scalar is recast as a native pointer.
func secretWallNativeServiceDefault410550(site uint32, args []secretWallNativeArg410550) secretWallNativeReply410550 {
	defer func() { runtime.KeepAlive(args) }()
	switch site {
	case 0x410551:
		return secretWallNativeReply410550{Pointer: unsafe.Pointer(C.nox_xxx_wallSecretGetFirstWall_410780())}
	case 0x410569, 0x4107b8, 0x4107d2, 0x4107eb:
		return secretWallNativeReply410550{Pointer: unsafe.Pointer(C.nox_xxx_wallSecretNext_410790((*C.nox_secret_wall_t)(args[0].Pointer)))}
	case 0x41073d, 0x4107e0, 0x4107f6:
		// Real C free is only valid for C-owned allocations. Pinned Go graph
		// controls must handle Free through the observer without deallocation.
		C.free(args[0].Pointer)
		return secretWallNativeReply410550{}
	default:
		panic("unknown secret-wall410550 service site")
	}
}
