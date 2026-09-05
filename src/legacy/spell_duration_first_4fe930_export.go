package legacy

/*
#include "spell_duration_first_4fe930.h"
*/
import "C"

import "unsafe"

func spellDurationFirstLegacy4FE930() unsafe.Pointer {
	return unsafe.Pointer(GetServer().S().Spells.Dur.SpellDurationFirst4FE930())
}

func spellDurationFirstExportCall4FE930() unsafe.Pointer {
	return C.nox_xxx_spellCastedFirst_4FE930()
}

//export nox_xxx_spellCastedFirst_4FE930
func nox_xxx_spellCastedFirst_4FE930() unsafe.Pointer {
	return spellDurationFirstLegacy4FE930()
}
