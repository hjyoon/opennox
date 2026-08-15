package legacy

/*
#include "monster_generator_collide_4ebe10.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_collideMonsterGen_4EBE10
func nox_xxx_collideMonsterGen_4EBE10(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().MonsterGeneratorCollide4EBE10(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		srv.NoxScriptC().ScriptCallback,
	)
}
