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

//export nox_xxx_castFumble_52C060
func nox_xxx_castFumble_52C060(id C.int, second, owner, caster *C.nox_object_t, arg unsafe.Pointer) C.int {
	// The original C ABI has five arguments. The Go selector's extra level is unused.
	return C.int(Nox_xxx_castFumble_52C060(spell.ID(id), asObjectS((*nox_object_t)(second)), asObjectS((*nox_object_t)(owner)), asObjectS((*nox_object_t)(caster)), (*server.SpellAcceptArg)(arg), 0))
}
