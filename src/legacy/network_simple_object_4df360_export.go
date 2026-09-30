package legacy

/*
#include "defs.h"
#include "GAME3_2.h"
*/
import "C"

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

// GAME.EXE 00419A70 uses FISTP with round-to-nearest-even. A C integer cast
// truncates instead, and cannot represent the x87 invalid-conversion result.
func netSendSimpleObjectFloatToInt4DF360(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

var netSendSimpleObjectCall4DF360 = func(recipient int32, obj *server.Object) int32 {
	return GetServer().S().NetSendSimpleObject4DF360(recipient, obj, netSendSimpleObjectFloatToInt4DF360)
}

func netSendSimpleObjectCEntry4DF360(recipient int32, obj *server.Object) int32 {
	return int32(C.nox_xxx_netSendSimpleObject2_4DF360(C.int(recipient), asObjectC(obj)))
}

//export nox_xxx_netSendSimpleObject_native_4DF360
func nox_xxx_netSendSimpleObject_native_4DF360(recipient C.int, obj *C.nox_object_t) C.int {
	return C.int(netSendSimpleObjectCall4DF360(int32(recipient), asObjectS(obj)))
}
