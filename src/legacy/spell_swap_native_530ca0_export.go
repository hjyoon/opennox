package legacy

/*
#include "spell_swap_native_530ca0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_530CA0
func sub_530CA0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellSwapCreate530CA0((*server.DurSpell)(record)))
}

//export sub_530D30
func sub_530D30(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellSwapUpdate530D30((*server.DurSpell)(record)))
}
