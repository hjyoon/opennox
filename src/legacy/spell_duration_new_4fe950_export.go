package legacy

/*
#include "spell_duration_new_4fe950.h"
*/
import "C"

import "unsafe"

func spellDurationNewLegacy4FE950() unsafe.Pointer {
	return unsafe.Pointer(GetServer().S().Spells.Dur.SpellDurationNew4FE950())
}

func spellDurationNewExportCall4FE950() unsafe.Pointer {
	return C.nox_xxx_newSpellDuration_4FE950()
}

//export nox_xxx_newSpellDuration_4FE950
func nox_xxx_newSpellDuration_4FE950() unsafe.Pointer {
	return spellDurationNewLegacy4FE950()
}
