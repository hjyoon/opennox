package legacy

/*
#include "spell_duration_cancel_4fe9d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellDurationCancelLegacy4FE9D0(record unsafe.Pointer) byte {
	return GetServer().S().Spells.Dur.SpellDurationCancel4FE9D0((*server.DurSpell)(record))
}

func spellDurationCancelExportCall4FE9D0(record unsafe.Pointer) byte {
	return byte(C.nox_xxx_spellCancelSpellDo_4FE9D0(record))
}

//export nox_xxx_spellCancelSpellDo_4FE9D0
func nox_xxx_spellCancelSpellDo_4FE9D0(record unsafe.Pointer) C.uint8_t {
	return C.uint8_t(spellDurationCancelLegacy4FE9D0(record))
}
