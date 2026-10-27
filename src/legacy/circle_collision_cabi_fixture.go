package legacy

/*
#include "GAME5.h"
#include "common/alloc/classes/alloc_class.h"
extern void* nox_alloc_hit_2491548;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func circleCollisionCABIReleaseFixturePool() {
	if C.nox_alloc_hit_2491548 != nil {
		C.nox_xxx_allocHitArray_5486D0()
		C.nox_free_alloc_class((*C.nox_alloc_class)(C.nox_alloc_hit_2491548))
		C.nox_alloc_hit_2491548 = nil
	}
}

func circleGeometryCABI54A990(center types.Pointf, radius float32, box *server.Object, normal *types.Pointf) float64 {
	return float64(C.sub_54A990((*C.float2)(unsafe.Pointer(&center)), C.float(radius), asObjectC(box), (*C.float2)(unsafe.Pointer(normal))))
}

func circleNearestCABI550280(second bool, origin types.Pointf, lo, hi float32, allowLo, allowHi int32, point types.Pointf, normal *types.Pointf) bool {
	if second {
		return C.sub_5502F0((*C.float2)(unsafe.Pointer(&origin)), C.float(lo), C.float(hi), C.int(allowLo), C.int(allowHi), (*C.float2)(unsafe.Pointer(&point)), (*C.float2)(unsafe.Pointer(normal))) != 0
	}
	return C.sub_550280((*C.float2)(unsafe.Pointer(&origin)), C.float(lo), C.float(hi), C.int(allowLo), C.int(allowHi), (*C.float2)(unsafe.Pointer(&point)), (*C.float2)(unsafe.Pointer(normal))) != 0
}

func circleObjectCABI550D00(kind string, first, second *server.Object, move int32, point types.Pointf) int {
	switch kind {
	case "circle-circle":
		C.nox_xxx_collisionCheckCircleCircle_550D00(asObjectC(first), asObjectC(second))
	case "circle-box":
		C.sub_54AD50(asObjectC(first), asObjectC(second), C.int(move))
	case "wall-response":
		return int(C.sub_550380(2, asObjectC(first), (*C.float2)(unsafe.Pointer(&point))))
	}
	return 0
}
