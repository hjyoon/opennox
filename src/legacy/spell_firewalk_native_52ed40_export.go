package legacy

/*
#include "spell_firewalk_native_52ed40.h"

_Static_assert(_Generic(&nox_xxx_firewalkTick_52ED40, int (*)(void*): 1, default: 0), "Firewalk native duration ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_firewalkTick_52ED40
func nox_xxx_firewalkTick_52ED40(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellFirewalkUpdate52ED40((*server.DurSpell)(record)))
}
