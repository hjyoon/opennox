package legacy

/*
#include "spell_teleport_markers_native_5305d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_5305D0
func sub_5305D0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTeleportToMarkCreate5305D0((*server.DurSpell)(record)))
}

//export sub_530650
func sub_530650(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTeleportToMarkUpdate530650((*server.DurSpell)(record)))
}

//export nox_xxx_castTele_530820
func nox_xxx_castTele_530820(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTeleportPopCreate530820((*server.DurSpell)(record)))
}

//export sub_530880
func sub_530880(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellTeleportPopUpdate530880((*server.DurSpell)(record)))
}
