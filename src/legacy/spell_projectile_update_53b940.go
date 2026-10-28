package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

var spellProjectileUpdateRuntime53B940 = func() server.SpellProjectileUpdateRuntime53B940 {
	world := GetServer().S()
	return server.SpellProjectileUpdateRuntime53B940{
		Frame:    world.Frame,
		TickRate: world.TickRate,
		Lifetime: world.Balance.Float,
		SearchTarget: func(missile, owner *server.Object, spellID uint32) *server.Object {
			flags := world.Spells.Flags(spell.ID(spellID))
			return world.Nox_xxx_spellFlySearchTarget(nil, missile, flags, 600, 0, owner)
		},
		SendPointFX: func(pos types.Pointf) {
			world.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(150), pos)
		},
		Expire: Sub_4E71F0,
	}
}

func spellProjectileUpdateCall53B940(missile *server.Object) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	server.SpellProjectileUpdate53B940(missile, spellProjectileUpdateRuntime53B940())
}

//export nox_spell_fly_update_native_53B940
func nox_spell_fly_update_native_53B940(missile *nox_object_t) {
	server.SpellProjectileUpdateRaw53B940(asObjectS(missile), spellProjectileUpdateRuntime53B940())
}
