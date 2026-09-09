package legacy

/*
#include "monster_wander_512930.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var monsterWanderCall512930 = func(unit *server.Object) {
	GetServer().S().ScriptMonsterRoam512930(unit)
}

func monsterWanderExportCall512930(unit *server.Object) {
	C.nox_xxx_scriptMonsterRoam_512930(asObjectC(unit))
}

//export nox_xxx_scriptMonsterRoam_512930
func nox_xxx_scriptMonsterRoam_512930(unit *C.nox_object_t) {
	monsterWanderCall512930(asObjectS((*nox_object_t)(unit)))
}
