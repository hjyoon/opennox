package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export sub_547DB0
func sub_547DB0(unit *C.nox_object_t, point *C.float2) C.int {
	if (*server.Object)(unsafe.Pointer(unit)).PointInShapeNative547DB0((*types.Pointf)(unsafe.Pointer(point))) {
		return 1
	}
	return 0
}

func pointOccupancyRuntime54EF00() pointOccupancyOriginalRuntime54EF00 {
	if override := pointOccupancyRuntimeOverride54EF00; override != nil {
		return *override
	}
	return pointOccupancyOriginalRuntime54EF00{
		Each:  func(rect types.Rectf, fn func(*server.Object) bool) { GetServer().S().Map.EachObjInRect(rect, fn) },
		Load:  func() uint32 { return memmap.Uint32(0x5D4594, 2491708) },
		Store: func(value uint32) { *memmap.PtrUint32(0x5D4594, 2491708) = value },
	}
}

//export sub_54EF00
func sub_54EF00(point *C.float2) C.int {
	r := pointOccupancyRuntime54EF00()
	return C.int(server.PointOccupiedNative54EF00((*types.Pointf)(unsafe.Pointer(point)), r.Each, r.Load, r.Store))
}

//export sub_54EF60
func sub_54EF60(unit *C.nox_object_t, point *C.float2) {
	r := pointOccupancyRuntime54EF00()
	(*server.Object)(unsafe.Pointer(unit)).PointOccupiedCandidateNative54EF60((*types.Pointf)(unsafe.Pointer(point)), r.Store)
}
