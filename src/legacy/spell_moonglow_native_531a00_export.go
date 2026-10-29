package legacy

/*
#include "spell_moonglow_native_531a00.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellCreateMoonglow_531A00
func nox_xxx_spellCreateMoonglow_531A00(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellMoonglowCreate531A00((*server.DurSpell)(record)))
}

//export sub_531AF0
func sub_531AF0(record unsafe.Pointer) {
	GetServer().SpellMoonglowDestroy531AF0((*server.DurSpell)(record))
}
