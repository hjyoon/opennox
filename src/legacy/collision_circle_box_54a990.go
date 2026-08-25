package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func circleBoxCollisionNative54A990(center types.Pointf, radius float32, box *server.Object) (float64, types.Pointf) {
	var normal types.Pointf
	result := C.sub_54A990(
		(*C.float2)(unsafe.Pointer(&center)),
		C.float(radius),
		asObjectC(box),
		(*C.float2)(unsafe.Pointer(&normal)),
	)
	return float64(result), normal
}
