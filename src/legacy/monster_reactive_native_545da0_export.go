package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func monsterReactiveRuntime545DA0() server.MonsterReactiveRuntime545DA0 {
	if runtime := monsterReactiveRuntimeOverride545DA0; runtime != nil {
		return *runtime
	}
	return server.MonsterReactiveRuntime545DA0{
		Frame:    func() uint32 { return GetServer().S().Frame() },
		TickRate: func() uint32 { return GetServer().S().TickRate() },
		Random:   func(min, max int) int { return GetServer().S().Rand.Logic.IntClamp(min, max) },
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			return unit.MonsterPushAction(action)
		},
	}
}

//export sub_545DA0
func sub_545DA0(unit *C.nox_object_t) C.int {
	return C.int(bool2int((*server.Object)(unsafe.Pointer(unit)).MonsterRecentSound545DA0(monsterReactiveRuntime545DA0())))
}

//export nox_xxx_monsterLookAtDamager_5466B0
func nox_xxx_monsterLookAtDamager_5466B0(unit *C.nox_object_t) C.int {
	return C.int(bool2int((*server.Object)(unsafe.Pointer(unit)).MonsterLookAtDamager5466B0(monsterReactiveRuntime545DA0())))
}
