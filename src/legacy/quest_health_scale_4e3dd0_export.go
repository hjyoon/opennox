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
		Difficulty:       func() float64 { return float64(C.sub_4E3CA0()) },
		StoreDamageScale: func(value float32) { C.sub_4E4080(C.float(value)) },
		StoreHealthScale: func(value float32) { C.sub_4E40C0(C.float(value)) },
		HealthScale:      func() float64 { return float64(C.sub_4E40F0()) },
		SetHP: func(obj *server.Object, value uint16) int32 {
			return int32(C.nox_xxx_unitSetHP_4E4560(asObjectC(obj), C.ushort(value)))
		},
	}
}

var questHealthScaleCall4E3DD0 = func() int16 {
	return GetServer().S().QuestHealthScale4E3DD0(questHealthScaleRuntime4E3DD0())
}

//export nox_xxx_questHealthScale_native_4E3DD0
func nox_xxx_questHealthScale_native_4E3DD0() C.short {
	return C.short(questHealthScaleCall4E3DD0())
}

func questHealthScaleCEntry4E3DD0() int16 { return int16(C.sub_4E3DD0()) }
