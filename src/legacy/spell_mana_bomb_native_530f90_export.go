package legacy

/*
#include "spell_mana_bomb_native_530f90.h"

_Static_assert(_Generic(&nox_xxx_manaBomb_530F90, int (*)(void*): 1, default: 0), "Mana Bomb native create ABI");
_Static_assert(_Generic(&nox_xxx_manaBombBoom_5310C0, int (*)(void*): 1, default: 0), "Mana Bomb native update ABI");
_Static_assert(_Generic(&sub_531290, int (*)(void*): 1, default: 0), "Mana Bomb native destroy ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_manaBomb_530F90
func nox_xxx_manaBomb_530F90(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellManaBombCreate530F90((*server.DurSpell)(record)))
}

//export nox_xxx_manaBombBoom_5310C0
func nox_xxx_manaBombBoom_5310C0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellManaBombUpdate5310C0((*server.DurSpell)(record)))
}

//export sub_531290
func sub_531290(record unsafe.Pointer) C.int {
	GetServer().SpellManaBombDestroy531290((*server.DurSpell)(record))
	// The duration manager discards the original incidental EAX.
	return 0
}
