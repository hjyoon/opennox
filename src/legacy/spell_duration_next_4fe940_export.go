package legacy

/*
#include "spell_duration_next_4fe940.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellDurationNextLegacy4FE940(record unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(server.SpellDurationNextNative4FE940((*server.DurSpell)(record)))
}

func spellDurationNextExportCall4FE940(record unsafe.Pointer) unsafe.Pointer {
	return C.nox_xxx_spellCastedNext_4FE940(record)
}

//export nox_xxx_spellCastedNext_4FE940
func nox_xxx_spellCastedNext_4FE940(record unsafe.Pointer) unsafe.Pointer {
	return spellDurationNextLegacy4FE940(record)
}
