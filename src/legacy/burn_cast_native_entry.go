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

//export nox_xxx_castBurn_52C3E0
func nox_xxx_castBurn_52C3E0(id C.int, second, owner, caster *C.nox_object_t, arg unsafe.Pointer) C.int {
	// The original public C entry has five arguments; only the Go selector has a level.
	return C.int(Nox_xxx_castBurn_52C3E0(spell.ID(id), asObjectS((*nox_object_t)(second)), asObjectS((*nox_object_t)(owner)), asObjectS((*nox_object_t)(caster)), (*server.SpellAcceptArg)(arg), 0))
}
