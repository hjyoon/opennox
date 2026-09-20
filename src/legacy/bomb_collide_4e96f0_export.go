package legacy

/*
#include "bomb_collide_4e96f0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var bombCollideCall4E96F0 = func(bomb, other *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().BombCollide4E96F0(
		bomb,
		other,
		collision,
		server.BombCollideRuntime4E96F0{
			ScriptCallback: srv.NoxScriptC().ScriptCallback,
			DamageClear: func(obj *server.Object, damage int32) {
				Nox_xxx_unitDamageClear_4EE5E0(obj, int(damage))
			},
		},
	)
}

//export nox_xxx_collideBomb_4E96F0
func nox_xxx_collideBomb_4E96F0(
	bomb, other *C.nox_object_t,
	collision *C.float,
) {
	bombCollideCall4E96F0(
		asObjectS((*nox_object_t)(bomb)),
		asObjectS((*nox_object_t)(other)),
		unsafe.Pointer(collision),
	)
}
