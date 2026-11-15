package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

//export nox_xxx_spellCastCleansingFlame_52D5C0
func nox_xxx_spellCastCleansingFlame_52D5C0(id C.int, second, recipient, caster *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(Nox_xxx_spellCastCleansingFlame_52D5C0(spell.ID(id), asObjectS((*nox_object_t)(second)), asObjectS((*nox_object_t)(recipient)), asObjectS((*nox_object_t)(caster)), (*server.SpellAcceptArg)(arg), int(level)))
}
