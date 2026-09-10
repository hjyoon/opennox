package legacy

/*
#include "unit_order_533900.h"
*/
import "C"

import (
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/server"
)

func unitOrderBanish5017F0(unit *server.Object) {
	srv := GetServer()
	s := srv.S()
	glyphType := s.Types.GlyphID()
	for item := unit.InvFirstItem; item != nil; {
		next := item.InvNextItem
		if int(item.TypeInd) == glyphType {
			srv.DelayedDelete(item)
		}
		item = next
	}
	s.Nox_xxx_netSendPointFx_522FF0(netmsg.MSG_FX_BLUE_SPARKS, unit.PosVec)
	update := (*server.MonsterUpdateData)(unit.UpdateData)
	srv.NoxScriptC().ScriptCallback(&update.ScriptDeath, nil, unit, server.NoxEventMonsterDead)
	srv.DelayedDelete(unit)
}

func unitOrderRuntime533900() server.UnitOrderRuntime533900 {
	return server.UnitOrderRuntime533900{
		MonsterDefByType: Nox_xxx_monsterDefByTT_517560,
		Banish:           unitOrderBanish5017F0,
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
