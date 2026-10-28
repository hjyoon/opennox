package legacy

/*
#include "spell_blink_native_530310.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellBlink2_530310
func nox_xxx_spellBlink2_530310(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellBlinkCreate530310((*server.DurSpell)(record)))
}

//export nox_xxx_spellBlink1_530380
func nox_xxx_spellBlink1_530380(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellBlinkUpdate530380((*server.DurSpell)(record)))
}

//export nox_xxx_spellTeleportCreateWake_530560
func nox_xxx_spellTeleportCreateWake_530560(target, origin, destination unsafe.Pointer) {
	GetServer().SpellTeleportCreateWake530560((*server.Object)(target), (*types.Pointf)(origin), (*types.Pointf)(destination))
}

func Get_nox_xxx_spellTeleportCreateWake_530560() unsafe.Pointer {
	return C.nox_xxx_spellTeleportCreateWake_530560
}
