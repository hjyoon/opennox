package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var castConfuseCall52C1E0 = func(caster *server.Object, arg *server.SpellAcceptArg, power int32) int32 {
	return GetServer().S().CastConfuse52C1E0(caster, arg, power, server.CastConfuseRuntime52C1E0{
		BuffApply:   buffApplyExportCall4FF380,
		Attribution: recordPlayerAttributionRuntime4E7540,
	})
}

//export nox_xxx_castConfuse_52C1E0
func nox_xxx_castConfuse_52C1E0(spellID C.int, second unsafe.Pointer, caster *C.nox_object_t, fourth *C.nox_object_t, arg unsafe.Pointer, power C.int) C.int {
	return C.int(castConfuseCall52C1E0(
		asObjectS((*nox_object_t)(caster)),
		(*server.SpellAcceptArg)(arg),
		int32(power),
	))
}
