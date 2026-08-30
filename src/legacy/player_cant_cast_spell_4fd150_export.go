package legacy

/*
#include "GAME4.h"
*/
import "C"

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func playerCantCastSpellExportCall4FD150(unit *server.Object, spellID, bypassModeRules int32) int32 {
	return int32(C.nox_xxx_checkPlrCantCastSpell_4FD150(
		asObjectC(unit),
		C.int(spellID),
		C.int(bypassModeRules),
	))
}

//export nox_xxx_checkPlrCantCastSpell_4FD150
func nox_xxx_checkPlrCantCastSpell_4FD150(
	unit *C.nox_object_t,
	spellID C.int,
	bypassModeRules C.int,
) C.int {
	return C.int(GetServer().S().CheckPlayerCantCastSpell4FD150(
		asObjectS((*nox_object_t)(unit)),
		spell.ID(spellID),
		int(bypassModeRules),
	))
}
