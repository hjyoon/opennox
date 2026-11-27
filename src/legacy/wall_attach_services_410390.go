package legacy

/*
#include <stdint.h>
#include <stddef.h>
#include "GAME1.h"
#include "GAME1_1.h"

_Static_assert(sizeof(int) == 4, "wall-attach410390 C int width");
_Static_assert(sizeof(float) == 4, "wall-attach410390 C float width");
_Static_assert(sizeof(int2) == 8 && offsetof(int2, field_4) == 4, "wall-attach410390 local int2 layout");

*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"
)

// Observer handling is checked by the caller before any production helper.
// The service calls share the implementations used by the C entry points.
func wallAttachNativeServiceDefault410390(site uint32, args []wallAttachNativeArg410390) wallAttachNativeReply410390 {
	defer func() { runtime.KeepAlive(args) }()
	switch site {
	case 0x4103a0:
		return wallAttachNativeReply410390{Pointer: nox_xxx_wall_4105E0(int32(args[0].Word), int32(args[1].Word))}
	case 0x4103b0:
		return wallAttachNativeReply410390{Pointer: nox_xxx_wallCreateAt_410250(int32(args[0].Word), int32(args[1].Word))}
	case 0x4103e3:
		return wallAttachNativeReply410390{Pointer: unsafe.Pointer(nox_xxx_polygonIsPlayerInPolygon_4217B0((*C.int2)(args[0].Pointer), C.int(args[1].Word)))}
	case 0x4103fa:
		return wallAttachNativeReply410390{Pointer: unsafe.Pointer(sub_421990((*C.int2)(args[0].Pointer), C.float(math.Float32frombits(args[1].Word)), C.int(args[2].Word)))}
	default:
		panic("unknown wall-attach410390 service site")
	}
}
