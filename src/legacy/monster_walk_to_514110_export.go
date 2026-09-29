package legacy

/*
#include "monster_walk_to_514110.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var monsterWalkToCall514110 = func(unit *server.Object, x, y float32) {
	GetServer().S().MonsterWalkTo514110(unit, x, y)
}

func monsterWalkToExportCall514110(unit *server.Object, x, y float32) {
	C.nox_xxx_monsterWalkTo_514110(asObjectC(unit), C.float(x), C.float(y))
}

//export nox_xxx_monsterWalkTo_514110
func nox_xxx_monsterWalkTo_514110(unit *C.nox_object_t, x, y C.float) {
	monsterWalkToCall514110(
		asObjectS((*nox_object_t)(unit)),
		float32(x),
		float32(y),
	)
}
