package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func monsterDodgeRuntime547C50() monsterDodgeOriginalRuntime547C50 {
	if override := monsterDodgeRuntimeOverride547C50; override != nil {
		return *override
	}
	return monsterDodgeOriginalRuntime547C50{
		RandomFloat: func(min, max float32) float64 { return GetServer().S().RandomFloat416030(min, max) },
		RandomInt:   func(min, max int) int { return GetServer().S().Rand.Logic.IntClamp(min, max) },
		Trace: func(ray *[4]float32, flags uint32) int {
			return int(C.nox_xxx_mapTraceRay_535250((*C.float4)(unsafe.Pointer(ray)), nil, nil, C.char(flags)))
		},
		Obstacles: func(unit *server.Object, ray *[4]float32) int {
			return int(C.nox_xxx_mapTraceObstacles_50B580((*C.nox_object_t)(unsafe.Pointer(unit)), (*C.float4)(unsafe.Pointer(ray))))
		},
		Tile: func(point *types.Pointf) int {
			return int(C.nox_xxx_tileNFromPoint_411160((*C.float2)(unsafe.Pointer(point))))
		},
		Pop: func(unit *server.Object) int { return unit.MonsterPopAction() },
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			return unit.MonsterPushAction(action)
		},
		Frame: func() uint32 { return GetServer().S().Frame() },
		FPS:   func() uint32 { return GetServer().S().TickRate() },
	}
}

//export nox_xxx_monsterCheckDodgeables_547C50
func nox_xxx_monsterCheckDodgeables_547C50(unit *C.nox_object_t) C.int {
	r := monsterDodgeRuntime547C50()
	return C.int((*server.Object)(unsafe.Pointer(unit)).MonsterCheckDodgeablesNative547C50(r.RandomFloat, r.RandomInt, r.Trace, r.Obstacles, r.Tile, r.Pop, r.Push, r.Frame, r.FPS))
}
