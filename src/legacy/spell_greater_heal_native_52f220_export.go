package legacy

/*
#include "spell_greater_heal_native_52f220.h"

_Static_assert(_Generic(&sub_52F220, int (*)(void*): 1, default: 0), "Greater Heal native create ABI");
_Static_assert(_Generic(&sub_52F2E0, int (*)(void*): 1, default: 0), "Greater Heal native update ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_52F220
func sub_52F220(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellGreaterHealCreate52F220((*server.DurSpell)(record)))
}

//export sub_52F2E0
func sub_52F2E0(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellGreaterHealUpdate52F2E0((*server.DurSpell)(record)))
}
