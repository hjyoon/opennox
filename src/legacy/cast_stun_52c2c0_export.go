package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var castStunCall52C2C0 = func(caster *server.Object, arg *server.SpellAcceptArg, power int32) int32 {
	return GetServer().S().CastStun52C2C0(caster, arg, power, server.CastStunRuntime52C2C0{
		BuffApply:   buffApplyExportCall4FF380,
		Attribution: recordPlayerAttributionRuntime4E7540,
	})
}

//export nox_xxx_castStun_52C2C0
func nox_xxx_castStun_52C2C0(spellID C.int, second unsafe.Pointer, caster *C.nox_object_t, fourth *C.nox_object_t, arg unsafe.Pointer, power C.int) C.int {
	return C.int(castStunCall52C2C0(
		asObjectS((*nox_object_t)(caster)),
		(*server.SpellAcceptArg)(arg),
		int32(power),
	))
}
