package legacy

/*
#include "spell_duration_selective_cancel_4feb10.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func spellCancelDurSpellLegacy4FEB10(spellID int32, caster *server.Object) int32 {
	return GetServer().S().Spells.Dur.SpellCancelDurSpell4FEB10(spellID, caster)
}

func spellCancelDurSpellExportCall4FEB10(spellID int32, caster *server.Object) {
	C.nox_xxx_spellCancelDurSpell_4FEB10(C.int32_t(spellID), asObjectC(caster))
}

//export nox_xxx_spellCancelDurSpell_4FEB10
func nox_xxx_spellCancelDurSpell_4FEB10(spellID C.int32_t, caster *C.nox_object_t) {
	_ = spellCancelDurSpellLegacy4FEB10(
		int32(spellID),
		asObjectS((*nox_object_t)(caster)),
	)
}
