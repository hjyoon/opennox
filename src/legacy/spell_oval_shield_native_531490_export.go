package legacy

/*
#include "spell_oval_shield_native_531490.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_531490
func sub_531490(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellOvalShieldCreate531490((*server.DurSpell)(record)))
}

//export sub_5314F0
func sub_5314F0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellOvalShieldUpdate5314F0((*server.DurSpell)(record)))
}

//export sub_531560
func sub_531560(record unsafe.Pointer) {
	GetServer().SpellOvalShieldDestroy531560((*server.DurSpell)(record))
}
