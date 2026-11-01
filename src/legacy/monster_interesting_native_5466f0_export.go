package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func monsterInterestingSoundRuntime5466F0() monsterInterestingSoundOriginalRuntime5466F0 {
	if override := monsterInterestingSoundRuntimeOverride5466F0; override != nil {
		return *override
	}
	return monsterInterestingSoundOriginalRuntime5466F0{
		Reactive: monsterReactiveRuntime545DA0(),
		Tile: func(point *types.Pointf) int {
			return Nox_xxx_tileNFromPoint_411160(*point)
		},
		Path: func(unit *server.Object, point *types.Pointf) bool {
			return GetServer().Sub_50B810(unit, point)
		},
		Trace: func(from, to types.Pointf) bool {
			return GetServer().S().MapTraceRayAt(from, to, nil, nil, 9)
		},
	}
}

//export sub_5466F0
func sub_5466F0(unit *C.nox_object_t) C.int {
	runtime := monsterInterestingSoundRuntime5466F0()
	return C.int((*server.Object)(unsafe.Pointer(unit)).MonsterInterestingSoundNative5466F0(runtime.Reactive, runtime.Tile, runtime.Path, runtime.Trace))
}
