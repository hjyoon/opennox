package legacy

/*
#include "spell_turn_undead_native_531310.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellTurnUndeadCreate_531310
func nox_xxx_spellTurnUndeadCreate_531310(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTurnUndeadCreate531310((*server.DurSpell)(record)))
}

//export nox_xxx_spellTurnUndeadDelete_531420
func nox_xxx_spellTurnUndeadDelete_531420(record unsafe.Pointer) {
	GetServer().SpellTurnUndeadDestroy531420((*server.DurSpell)(record))
}
