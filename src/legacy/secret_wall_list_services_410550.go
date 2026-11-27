package legacy

/*
#include <stdlib.h>
#include <string.h>

static void secret_wall_native_foreign_store_pointer_410550(void* slot, void* value) {
    memcpy(slot, &value, sizeof(value));
}
*/
import "C"

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// The live native slot holds the secret wall list head.
func secretWallNativeHeadCell410550() unsafe.Pointer {
	return unsafe.Pointer(&worldOwnerStorage.secret_wall_head)
}

func secretWallNativeForeignLoadWord410550(slot unsafe.Pointer) uint16 {
	return binary.NativeEndian.Uint16(unsafe.Slice((*byte)(slot), 2))
}

func secretWallNativeForeignStorePointer410550(slot, value unsafe.Pointer) {
	C.secret_wall_native_foreign_store_pointer_410550(slot, value)
	runtime.KeepAlive(slot)
	runtime.KeepAlive(value)
}

// The observer handles diagnostic replies before the live service runs.
func secretWallNativeServiceDefault410550(site uint32, args []secretWallNativeArg410550) secretWallNativeReply410550 {
	defer func() { runtime.KeepAlive(args) }()
	switch site {
	case 0x410551:
		return secretWallNativeReply410550{Pointer: unsafe.Pointer(secretWallFirstNative410780())}
	case 0x410569, 0x4107b8, 0x4107d2, 0x4107eb:
		return secretWallNativeReply410550{Pointer: unsafe.Pointer(secretWallNextNative410790((*server.SecretWall)(args[0].Pointer)))}
	case 0x41073d, 0x4107e0, 0x4107f6:
		// Real C free is only valid for C-owned allocations. Pinned Go graph
		// controls must handle Free through the observer without deallocation.
		C.free(args[0].Pointer)
		return secretWallNativeReply410550{}
	default:
		panic("unknown secret-wall410550 service site")
	}
}
