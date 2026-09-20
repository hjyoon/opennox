package legacy

/*
#include "monster_generator_collide_4ebe10.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var monsterGeneratorCollideCall4EBE10 = func(source, target *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().MonsterGeneratorCollide4EBE10(
		source,
		target,
		(*types.Pointf)(collision),
		srv.NoxScriptC().ScriptCallback,
	)
}

//export nox_xxx_collideMonsterGen_4EBE10
func nox_xxx_collideMonsterGen_4EBE10(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	monsterGeneratorCollideCall4EBE10(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
