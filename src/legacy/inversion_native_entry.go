package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "github.com/opennox/libs/spell"

//export sub_52BEB0
func sub_52BEB0(id C.int, second, owner, caster *C.nox_object_t) C.int {
	// Acceptance and level exist only in the six-argument Go selector.
	return C.int(Sub_52BEB0(spell.ID(id), asObjectS((*nox_object_t)(second)), asObjectS((*nox_object_t)(owner)), asObjectS((*nox_object_t)(caster)), nil, 0))
}
