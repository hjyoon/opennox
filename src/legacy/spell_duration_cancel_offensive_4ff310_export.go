package legacy

/*
#include "spell_duration_cancel_offensive_4ff310.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func spellDurationCancelOffensiveLegacy4FF310(caster *server.Object) {
	GetServer().S().Spells.Dur.SpellDurationCancelOffensive4FF310(caster)
}

func spellDurationCancelOffensiveExportCall4FF310(caster *server.Object) {
	C.sub_4FF310(asObjectC(caster))
}

//export sub_4FF310
func sub_4FF310(caster *C.nox_object_t) {
	spellDurationCancelOffensiveLegacy4FF310(
		asObjectS((*nox_object_t)(caster)),
	)
}
