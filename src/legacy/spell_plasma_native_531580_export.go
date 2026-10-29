package legacy

/*
#include "spell_plasma_native_531580.h"

_Static_assert(_Generic(&nox_xxx_plasmaSmth_531580, int (*)(void*): 1, default: 0), "Plasma create native ABI");
_Static_assert(_Generic(&nox_xxx_plasmaShot_531600, int (*)(void*): 1, default: 0), "Plasma update native ABI");
_Static_assert(_Generic(&sub_5319E0, void (*)(void*): 1, default: 0), "Plasma destroy native ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_plasmaSmth_531580
func nox_xxx_plasmaSmth_531580(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellPlasmaCreate531580((*server.DurSpell)(record)))
}

//export nox_xxx_plasmaShot_531600
func nox_xxx_plasmaShot_531600(record unsafe.Pointer) C.int {
	return C.int(GetServer().SpellPlasmaUpdate531600((*server.DurSpell)(record)))
}

//export sub_5319E0
func sub_5319E0(record unsafe.Pointer) {
	GetServer().SpellPlasmaDestroy5319E0((*server.DurSpell)(record))
}
