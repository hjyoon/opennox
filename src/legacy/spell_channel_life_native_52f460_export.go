package legacy

/*
#include "spell_channel_life_native_52f460.h"

_Static_assert(_Generic(&sub_52F460, int (*)(void*): 1, default: 0), "Channel Life native update ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

//export sub_52F460
func sub_52F460(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellChannelLifeUpdate52F460((*server.DurSpell)(record)))
}
