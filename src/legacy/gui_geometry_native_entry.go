package legacy

/*
#include "GAME1_2.h"
#include "GAME2_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/ntype"
)

//export sub_46AE10
func sub_46AE10(win *C.nox_window, enabled C.int) *C.nox_window {
	return (*C.nox_window)(guiGeometryStyle46AE10(asWindow((*nox_window)(win)), int32(enabled)).C())
}

//export nox_xxx_pointInRect_4281F0
func nox_xxx_pointInRect_4281F0(point *C.int2, bounds *C.int4) C.int {
	return C.int(guiGeometryPointInRect4281F0((*ntype.Point32)(unsafe.Pointer(point)), (*[4]int32)(unsafe.Pointer(bounds))))
}

//export sub_428220
func sub_428220(point *C.float2, bounds *C.float4) C.int {
	return C.int(guiGeometryFloatInRect428220((*[2]float32)(unsafe.Pointer(point)), (*[4]float32)(unsafe.Pointer(bounds))))
}

//export sub_4739E0
func sub_4739E0(vp *C.nox_draw_viewport_t, point, out *C.int2) C.int {
	return C.int(guiGeometryToScreen4739E0((*noxrender.Viewport)(unsafe.Pointer(vp)), (*ntype.Point32)(unsafe.Pointer(point)), (*ntype.Point32)(unsafe.Pointer(out))))
}

//export sub_473A10
func sub_473A10(vp *C.nox_draw_viewport_t, point *C.int2, out *C.uint32_t) C.int {
	return C.int(guiGeometryToWorld473A10((*noxrender.Viewport)(unsafe.Pointer(vp)), (*ntype.Point32)(unsafe.Pointer(point)), (*ntype.Point32)(unsafe.Pointer(out))))
}
