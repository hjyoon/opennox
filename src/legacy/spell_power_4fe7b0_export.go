package legacy

/*
#include "spell_power_4fe7b0.h"
*/
import "C"

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func spellPowerLegacy4FE7B0(spellID spell.ID, caster *server.Object) int32 {
	return GetServer().SpellPower4FE7B0(spellID, caster)
}

func spellPowerExportCall4FE7B0(spellID int32, caster *server.Object) int32 {
	return int32(C.nox_xxx_spellGetPower_4FE7B0(C.int32_t(spellID), asObjectC(caster)))
}

//export nox_xxx_spellGetPower_4FE7B0
func nox_xxx_spellGetPower_4FE7B0(spellID C.int32_t, caster *C.nox_object_t) C.int32_t {
	return C.int32_t(spellPowerLegacy4FE7B0(
		spell.ID(int32(spellID)),
		asObjectS((*nox_object_t)(caster)),
	))
}
