package legacy

/*
#include <stdint.h>
#include "unit_damage_clear_4ee5e0.h"

int nox_xxx_monsterCallDieFn_50A3D0(uint32_t* unit);
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitDamageClearMonsterDie4EE5E0(unit *server.Object) {
	C.nox_xxx_monsterCallDieFn_50A3D0((*C.uint32_t)(unit.CObj()))
}

//export nox_xxx_unitDamageClear_4EE5E0
func nox_xxx_unitDamageClear_4EE5E0(unit *C.nox_object_t, damage C.int) {
	unitDamageClearCall4EE5E0(
		asObjectS((*nox_object_t)(unit)),
		int32(damage),
	)
}
