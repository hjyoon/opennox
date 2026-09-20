package legacy

/*
// The shared header still contains unrelated Win32-only layout assertions.
// These boundaries use fixed-width eight-byte records on every target.
#define _Static_assert(...)
#include "GAME4_1.h"
#undef _Static_assert
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func directionToAngleExportCall509E00(data *server.DirectionInitData) int32 {
	return int32(C.nox_xxx_xferDirectionToAngle_509E00((*C.uint32_t)(unsafe.Pointer(data))))
}

func indexedDirectionExportCall509E20(direction int32, out *server.IndexedDirectionVector509E20) int32 {
	return int32(C.nox_xxx_xferIndexedDirection_509E20(
		C.int(direction),
		(*C.int2)(unsafe.Pointer(out)),
	))
}

func directionIndexToAngleExportCall509E90(index int32) int32 {
	return int32(C.nox_xxx_mathDirection4ToAngle_509E90(C.int(index)))
}

func directionOctantExportCall509EA0(direction int32) int32 {
	return int32(C.nox_xxx_math_509EA0(C.int(direction)))
}

func normalizeVectorExportCall509F20(point *types.Pointf) {
	C.nox_xxx_utilNormalizeVector_509F20((*C.float2)(unsafe.Pointer(point)))
}

//export nox_xxx_xferDirectionToAngle_509E00
func nox_xxx_xferDirectionToAngle_509E00(data *C.uint32_t) C.int {
	return C.int(server.DirectionToAngle509E00(
		(*server.DirectionInitData)(unsafe.Pointer(data)),
	))
}

//export nox_xxx_xferIndexedDirection_509E20
func nox_xxx_xferIndexedDirection_509E20(direction C.int, out *C.int2) C.int {
	return C.int(server.IndexedDirection509E20(
		int32(direction),
		(*server.IndexedDirectionVector509E20)(unsafe.Pointer(out)),
	))
}

//export nox_xxx_mathDirection4ToAngle_509E90
func nox_xxx_mathDirection4ToAngle_509E90(index C.int) C.int {
	return C.int(server.DirectionIndexToAngle509E90(int32(index)))
}

//export nox_xxx_math_509EA0
func nox_xxx_math_509EA0(direction C.int) C.int {
	return C.int(server.DirectionOctant509EA0(int32(direction)))
}

//export nox_xxx_utilNormalizeVector_509F20
func nox_xxx_utilNormalizeVector_509F20(point *C.float2) {
	server.NormalizeVector509F20((*types.Pointf)(unsafe.Pointer(point)))
}
