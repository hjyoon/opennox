package legacy

/*
#include "GAME2_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap/nox/blobdata"

func approximateDistance32(dx, dy int32) uint32 {
	return blobdata.ApproxSqrt32(uint32(dx*dx + dy*dy))
}

//export sub_48C690
func sub_48C690(x1, y1, x2, y2 C.int) C.uint {
	return C.uint(approximateDistance32(int32(x2-x1), int32(y2-y1)))
}

//export sub_48C6B0
func sub_48C6B0(dx, dy C.int) C.uint {
	return C.uint(approximateDistance32(int32(dx), int32(dy)))
}

//export sub_48C730
func sub_48C730(value C.uint) C.uint {
	return C.uint(blobdata.ApproxSqrt32(uint32(value)))
}
