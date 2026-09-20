package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var monsterCollideCall4E83B0 = func(monster, other *server.Object, collision unsafe.Pointer) unsafe.Pointer {
	_ = collision // Present in the registered collide ABI; 004E83B0 does not read it.
	return server.MonsterCollideScript4E83B0(
		monster,
		other,
		GetServer().NoxScriptC().ScriptCallback,
	)
}

//export nox_xxx_collideMonsterEventProc_4E83B0
func nox_xxx_collideMonsterEventProc_4E83B0(monster, other *C.nox_object_t, collision *C.float) unsafe.Pointer {
	return monsterCollideCall4E83B0(
		asObjectS((*nox_object_t)(monster)),
		asObjectS((*nox_object_t)(other)),
		unsafe.Pointer(collision),
	)
}
