package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_castPull_52BFA0
func nox_xxx_castPull_52BFA0(id C.int, second, owner, caster *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(Nox_xxx_castPull_52BFA0(spell.ID(id), asObjectS((*nox_object_t)(second)), asObjectS((*nox_object_t)(owner)), asObjectS((*nox_object_t)(caster)), (*server.SpellAcceptArg)(arg), int(level)))
}

//export nox_xxx_castPush_52C000
func nox_xxx_castPush_52C000(id C.int, second, owner, caster *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(Nox_xxx_castPush_52C000(spell.ID(id), asObjectS((*nox_object_t)(second)), asObjectS((*nox_object_t)(owner)), asObjectS((*nox_object_t)(caster)), (*server.SpellAcceptArg)(arg), int(level)))
}
