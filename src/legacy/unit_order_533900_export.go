package legacy

/*
#include "unit_order_533900.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitOrderRuntime533900() server.UnitOrderRuntime533900 {
	return server.UnitOrderRuntime533900{
		MonsterDefByType: Nox_xxx_monsterDefByTT_517560,
		Banish:           Nox_xxx_banishUnit_5017F0,
		Observe:          Nox_xxx_playerObserveMonster_4DDE80,
	}
}

var unitOrderCall533900 = func(owner, creature *server.Object, order uint32) {
	GetServer().S().OrderUnit533900(owner, creature, order, unitOrderRuntime533900())
}

func unitOrderExportCall533900(owner, creature *server.Object, order int32) {
	C.nox_xxx_orderUnit_533900(asObjectC(owner), asObjectC(creature), C.int32_t(order))
}

//export nox_xxx_orderUnit_533900
func nox_xxx_orderUnit_533900(owner, creature *C.nox_object_t, order C.int32_t) {
	unitOrderCall533900(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(creature)),
		uint32(order),
	)
}

func Nox_xxx_orderUnit_533900(owner, creature *server.Object, order uint32) {
	unitOrderCall533900(owner, creature, order)
}
