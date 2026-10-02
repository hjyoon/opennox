package legacy

/*
#include "cast_shock_52c5a0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var castShockCall52C5A0 = func(caster, context *server.Object, arg *server.SpellAcceptArg, power int32) int32 {
	return server.CastShock52C5A0(caster, context, arg, power, server.CastShockRuntime52C5A0{
		GlyphTypeCache: Get_dword_5d4594_2487712_ptr(),
		LookupTypeID: func(name string) uint32 {
			return uint32(GetServer().S().Types.IndByID(name))
		},
		BalanceFloat: func(key string) float64 {
			return GetServer().S().Balance.Float(key)
		},
		BalanceFloatInd: func(key string, index int32) float64 {
			return GetServer().S().Balance.FloatInd(key, int(index))
		},
		BuffApply: buffApplyExportCall4FF380,
	})
}

//export nox_xxx_useShock_52C5A0
func nox_xxx_useShock_52C5A0(spellID C.int, second unsafe.Pointer, caster, context *C.nox_object_t, arg unsafe.Pointer, power C.int) C.int {
	return C.int(castShockCall52C5A0(
		asObjectS((*nox_object_t)(caster)),
		asObjectS((*nox_object_t)(context)),
		(*server.SpellAcceptArg)(arg),
		int32(power),
	))
}
