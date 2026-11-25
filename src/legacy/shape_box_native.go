package legacy

/*
#include "GAME1_2.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

//export nox_shape_box_calc
func nox_shape_box_calc(shape *C.nox_shape) {
	(*server.Shape)(unsafe.Pointer(shape)).Box.Calc()
}

func shapeBoxCalcCallNative428270(shape *server.Shape) {
	C.nox_shape_box_calc((*C.nox_shape)(unsafe.Pointer(shape)))
}
