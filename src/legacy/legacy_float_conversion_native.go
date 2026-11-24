package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

//export nox_float2int
func nox_float2int(value C.float) C.int {
	return C.int(server.X87TruncateSignedDWORD419A70(float32(value)))
}

//export nox_float2int16
func nox_float2int16(value C.float) C.short {
	return C.short(nox_float2int(value))
}

//export nox_float2int16_abs
func nox_float2int16_abs(value C.float) C.short {
	return nox_float2int16(C.float(math.Abs(float64(value))))
}

//export nox_double2float
func nox_double2float(value C.double) C.float {
	return C.float(storeFloat32TowardZero(float64(value)))
}

//export nox_double2int
func nox_double2int(value C.double) C.int {
	x := float64(value)
	if !(x >= -0x1p63 && x < 0x1p63) {
		return 0
	}
	return C.int(int32(int64(x)))
}
