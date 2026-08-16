package legacy

/*
#include "random_reachable_point_4ed970.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export sub_4ED970
func sub_4ED970(radius C.float, center, output *C.float2) *C.float2 {
	result := GetServer().S().RandomReachablePointAroundInto4ED970(
		float32(radius),
		(*types.Pointf)(unsafe.Pointer(center)),
		(*types.Pointf)(unsafe.Pointer(output)),
	)
	return (*C.float2)(unsafe.Pointer(result))
}
