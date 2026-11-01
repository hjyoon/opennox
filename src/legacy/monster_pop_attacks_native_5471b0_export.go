package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_monsterPopAttackActions_5471B0
func nox_xxx_monsterPopAttackActions_5471B0(unit *C.nox_object_t) {
	pop := monsterPopAttacksOverride5471B0
	if pop == nil {
		pop = func(unit *server.Object) int { return unit.MonsterPopAction() }
	}
	(*server.Object)(unsafe.Pointer(unit)).MonsterPopAttackActionsNative5471B0(pop)
}
