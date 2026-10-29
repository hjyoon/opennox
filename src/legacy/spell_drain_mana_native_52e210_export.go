package legacy

/*
#include "spell_drain_mana_native_52e210.h"

_Static_assert(_Generic(&nox_xxx_spellDrainMana_52E210, int (*)(void*): 1, default: 0), "Drain Mana native duration ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellDrainMana_52E210
func nox_xxx_spellDrainMana_52E210(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellDrainManaUpdate52E210((*server.DurSpell)(record)))
}
