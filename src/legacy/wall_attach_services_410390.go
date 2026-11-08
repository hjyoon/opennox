package legacy

/*
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include "GAME1.h"
#include "GAME1_1.h"

_Static_assert(sizeof(int) == 4, "wall-attach410390 C int width");
_Static_assert(sizeof(float) == 4, "wall-attach410390 C float width");
_Static_assert(sizeof(int2) == 8 && offsetof(int2, field_4) == 4, "wall-attach410390 local int2 layout");

static void* wall_attach_polygon_query_410390(void* point, int32_t selector) {
	return (void*)nox_xxx_polygonIsPlayerInPolygon_4217B0((int2*)point, (int)selector);
}
static void* wall_attach_polygon_nearby_410390(void* point, uint32_t radius_bits, int32_t selector) {
	float radius;
	memcpy(&radius, &radius_bits, sizeof(radius));
	return (void*)sub_421990((int2*)point, radius, (int)selector);
}
static void wall_attach_foreign_store_pointer_410390(void* slot, void* value) {
	memcpy(slot, &value, sizeof(value));
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func wallAttachNativeForeignStorePointer410390(slot, value unsafe.Pointer) {
	C.wall_attach_foreign_store_pointer_410390(slot, value)
	runtime.KeepAlive(slot)
	runtime.KeepAlive(value)
}

// Observer handling is checked by the caller before any production helper.
// These four existing bodies remain external; their internal algorithms are
// not reconstructed or credited by the raw410390 caller port.
func wallAttachNativeServiceDefault410390(site uint32, args []wallAttachNativeArg410390) wallAttachNativeReply410390 {
	defer func() { runtime.KeepAlive(args) }()
	switch site {
	case 0x4103a0:
		return wallAttachNativeReply410390{Pointer: nox_xxx_wall_4105E0(int32(args[0].Word), int32(args[1].Word))}
	case 0x4103b0:
		return wallAttachNativeReply410390{Pointer: nox_xxx_wallCreateAt_410250(int32(args[0].Word), int32(args[1].Word))}
	case 0x4103e3:
		return wallAttachNativeReply410390{Pointer: C.wall_attach_polygon_query_410390(args[0].Pointer, C.int32_t(int32(args[1].Word)))}
	case 0x4103fa:
		return wallAttachNativeReply410390{Pointer: C.wall_attach_polygon_nearby_410390(args[0].Pointer, C.uint32_t(args[1].Word), C.int32_t(int32(args[2].Word)))}
	default:
		panic("unknown wall-attach410390 service site")
	}
}
