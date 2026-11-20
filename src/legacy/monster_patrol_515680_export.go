package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_monsterGoPatrol_515680
func nox_xxx_monsterGoPatrol_515680(unit *C.nox_object_t, request *C.nox_monster_patrol_request_t) {
	server.MonsterGoPatrolFromRecord515680(asObjectS(unit), (*server.MonsterPatrolRequest515680)(unsafe.Pointer(request)))
}
