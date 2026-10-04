package legacy

/*
#include "monster_look_at_5125a0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var monsterLookAtCall5125A0 = server.MonsterLookAt5125A0

func monsterLookAtCResult5125A0(unit *server.Object, direction int32) uintptr {
	return uintptr(unsafe.Pointer(C.nox_xxx_monsterLookAt_5125A0(
		(*C.nox_object_t)(unsafe.Pointer(unit)), C.int(direction),
	)))
}

//export nox_server_monster_look_at_5125a0
func nox_server_monster_look_at_5125a0(unit *C.nox_object_t, direction C.int) C.uintptr_t {
	return C.uintptr_t(monsterLookAtCall5125A0(asObjectS((*nox_object_t)(unit)), int32(direction)))
}
