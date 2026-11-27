package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func questHealthScaleRuntime4E3DD0() server.QuestHealthScaleRuntime4E3DD0 {
	return server.QuestHealthScaleRuntime4E3DD0{
		Ready:            memmap.PtrUint32(0x5D4594, 1563932),
		DamageInit:       memmap.PtrFloat32(0x5D4594, 1563908),
		HealthInit:       memmap.PtrFloat32(0x5D4594, 1563916),
		DamageCoeff:      memmap.PtrFloat32(0x5D4594, 1563920),
		HealthCoeff:      memmap.PtrFloat32(0x5D4594, 1563924),
		Difficulty:       func() float64 { return questParameterFloat(questDifficulty) },
		StoreDamageScale: setQuestDamageScale,
		StoreHealthScale: setQuestHealthScale,
		HealthScale:      func() float64 { return questParameterFloat(questHealthScale) },
		SetHP: func(obj *server.Object, value uint16) int32 {
			// Quest scaling observes only the low WORD of the setter result.
			return int32(uint16(unitSetHPNative4E4560(obj, value)))
		},
	}
}

var questHealthScaleCall4E3DD0 = func() int16 {
	return GetServer().S().QuestHealthScale4E3DD0(questHealthScaleRuntime4E3DD0())
}

//export sub_4E3DD0
func sub_4E3DD0() C.short {
	return C.short(questHealthScaleCall4E3DD0())
}

func questHealthScaleCEntry4E3DD0() int16 { return int16(C.sub_4E3DD0()) }
