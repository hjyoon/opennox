package legacy

/*
#include "spell_precheck_4fd0e0.h"
*/
import "C"

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func spellPrecheckExportCall4FD0E0(unit *server.Object, spellID int32) int32 {
	return int32(C.sub_4FD0E0(asObjectC(unit), C.int32_t(spellID)))
}

//export sub_4FD0E0
func sub_4FD0E0(unit *C.nox_object_t, spellID C.int32_t) C.int32_t {
	return C.int32_t(GetServer().S().SpellPrecheck4FD0E0(
		asObjectS((*nox_object_t)(unit)),
		spell.ID(int32(spellID)),
	))
}
