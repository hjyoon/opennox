package legacy

/*
#include "GAME3_3.h"
static void* raw_skeleton_foreign_pointer_4e23c0(uintptr_t bits) { return (void*)bits; }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func rawSkeletonForeignPointer4E23C0(bits uintptr) unsafe.Pointer {
	return C.raw_skeleton_foreign_pointer_4e23c0(C.uintptr_t(bits))
}

func rawSkeletonFallbackService4E23C0(call rawSkeletonService4E23C0) uint32 {
	a := call.Arguments
	switch call.Site {
	case 0x4e23ce:
		return uint32(unitBuffTestLegacy4FF350((*server.Object)(a[0].Pointer), int32(a[1].Word)))
	case 0x4e23ea, 0x4e247b:
		GetServer().S().Audio.EventObj(sound.ID(int32(a[0].Word)), (*server.Object)(a[1].Pointer), int(int32(a[2].Word)), a[3].Word)
		return 0 // Audio has no result consumed by this caller.
	case 0x4e243f:
		return uint32(nox_server_testTwoPointsAndDirection_4E6E50((*C.float2)(a[0].Pointer), C.int(int32(a[1].Word)), (*C.float2)(a[2].Pointer)))
	case 0x4e2452:
		return uint32((*server.Object)(a[0].Pointer).MonsterActionGet50A020())
	case 0x4e249d:
		return uint32(rawDefaultDamageNative4E0B30(a[0].Pointer, a[1].Pointer, a[2].Pointer, int32(a[3].Word), int32(a[4].Word)))
	default:
		panic("undeclared raw Skeleton service site")
	}
}

//export sub_4E23C0
func sub_4E23C0(target, source, weapon *C.nox_object_t, damage, typ C.int32_t) C.int32_t {
	return C.int32_t(rawSkeletonDamageNative4E23C0(unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), int32(damage), int32(typ)))
}
