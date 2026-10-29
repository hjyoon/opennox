package legacy

/*
#include "spell_force_of_nature_native_52ef30.h"

_Static_assert(_Generic(&sub_52EF30, int (*)(void*): 1, default: 0), "Force of Nature native create ABI");
_Static_assert(_Generic(&sub_52EFD0, int (*)(void*): 1, default: 0), "Force of Nature native update ABI");
_Static_assert(_Generic(&sub_52F1D0, int (*)(void*): 1, default: 0), "Force of Nature native destroy ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_52EF30
func sub_52EF30(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellForceOfNatureCreate52EF30((*server.DurSpell)(record)))
}

//export sub_52EFD0
func sub_52EFD0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellForceOfNatureUpdate52EFD0((*server.DurSpell)(record)))
}

//export sub_52F1D0
func sub_52F1D0(record unsafe.Pointer) C.int {
	GetServer().SpellForceOfNatureDestroy52F1D0((*server.DurSpell)(record))
	// The duration manager discards the original pointer-valued EAX.
	return 0
}
