package legacy

/*
#include "spell_teleport_target_native_530a30.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_530A30_spell_execdur
func sub_530A30_spell_execdur(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTeleportToTargetCreate530A30((*server.DurSpell)(record)))
}

//export nox_xxx_castTTT_530B70
func nox_xxx_castTTT_530B70(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTeleportToTargetUpdate530B70((*server.DurSpell)(record)))
}
