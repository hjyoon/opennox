package legacy

/*
#include "GAME5_2.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var (
	_ = [1]struct{}{}[unsafe.Sizeof(C.float2{})-unsafe.Sizeof(types.Pointf{})]
	_ = [1]struct{}{}[unsafe.Sizeof(C.float4{})-unsafe.Sizeof(types.Rectf{})]
	_ = [1]struct{}{}[unsafe.Offsetof(C.float4{}.field_8)-unsafe.Offsetof(types.Rectf{}.Max)]
)

//export sub_57C790
func sub_57C790(line *C.float4, point, out *C.float2, length C.float) {
	server.ClampPointToSegment((*types.Rectf)(unsafe.Pointer(line)), *(*types.Pointf)(unsafe.Pointer(point)),
		float32(length), (*types.Pointf)(unsafe.Pointer(out)))
}

//export nox_xxx_mathPointOnTheLine_57C8A0
func nox_xxx_mathPointOnTheLine_57C8A0(line *C.float4, point, out *C.float2) C.int {
	return C.int(bool2int(server.PointOnTheLineInto((*types.Rectf)(unsafe.Pointer(line)),
		*(*types.Pointf)(unsafe.Pointer(point)), (*types.Pointf)(unsafe.Pointer(out)))))
}

//export sub_57CDB0
func sub_57CDB0(grid *C.int2, ray *C.float, out *C.float2) C.int {
	cell := image.Pt(int(grid.field_0), int(grid.field_4))
	normal, ok := GetServer().S().ProjectileWallNormal(cell, *(*types.Rectf)(unsafe.Pointer(ray)))
	if ok {
		*(*types.Pointf)(unsafe.Pointer(out)) = normal
		return 1
	}
	return 0
}

//export sub_57F1D0
func sub_57F1D0(point *C.float2) C.char {
	return C.char(server.WallRayTargetQuadrant57F1D0(*(*types.Pointf)(unsafe.Pointer(point))))
}

//export sub_57F2A0
func sub_57F2A0(point *C.float2, x, y C.int) C.int {
	return C.int(server.WallRayRegion57F2A0(*(*types.Pointf)(unsafe.Pointer(point)), int32(x), int32(y)))
}
