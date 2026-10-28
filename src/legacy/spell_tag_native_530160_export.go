package legacy

/*
#include "spell_tag_native_530160.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellTagCreature_530160
func nox_xxx_spellTagCreature_530160(record unsafe.Pointer) C.int {
	return C.int(GetServer().S().SpellTagCreate530160((*server.DurSpell)(record)))
}

//export sub_530250
func sub_530250(record unsafe.Pointer) C.uint {
	return C.uint(server.SpellTagUpdate530250((*server.DurSpell)(record)))
}

//export sub_530270
func sub_530270(record unsafe.Pointer) {
	GetServer().S().SpellTagDestroy530270((*server.DurSpell)(record))
}
