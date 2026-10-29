package legacy

/*
#include "spell_energy_bolt_native_52e820.h"

_Static_assert(_Generic(&nox_xxx_spellEnergyBoltStop_52E820, int (*)(void*): 1, default: 0), "Lightning create native ABI");
_Static_assert(_Generic(&nox_xxx_spellEnergyBoltTick_52E850, int (*)(void*): 1, default: 0), "Lightning update native ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellEnergyBoltStop_52E820
func nox_xxx_spellEnergyBoltStop_52E820(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellEnergyBoltCreate52E820((*server.DurSpell)(record)))
}

//export nox_xxx_spellEnergyBoltTick_52E850
func nox_xxx_spellEnergyBoltTick_52E850(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellEnergyBoltUpdate52E850((*server.DurSpell)(record)))
}
