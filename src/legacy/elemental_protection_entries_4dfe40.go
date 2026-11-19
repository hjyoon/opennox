package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_checkFireProtect_4DFE40
func nox_xxx_checkFireProtect_4DFE40(unit *nox_object_t) C.double {
	if unit == nil {
		return 0
	}
	return C.double(fireProtectionCall4DFE40(GetServer().S(), asObjectS(unit)))
}

//export nox_xxx_checkElectrProtect_4DFF40
func nox_xxx_checkElectrProtect_4DFF40(unit *nox_object_t) C.double {
	if unit == nil {
		return 0
	}
	return C.double(electricProtectionCall4DFF40(GetServer().S(), asObjectS(unit)))
}

//export nox_xxx_getPoisonDmg_4E0040
func nox_xxx_getPoisonDmg_4E0040(unit *nox_object_t) C.double {
	if unit == nil {
		return 0
	}
	return C.double(GetServer().S().PoisonProtection4E0040(asObjectS(unit), server.PoisonProtectionRuntime4E0040{
		PoisonProtectEngage: C.nox_xxx_checkPoisonProtectEnch_4DFDE0,
	}))
}
