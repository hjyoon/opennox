package legacy

import (
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/server"
)

func magicMissileUpdateCall53BDA0(missile *server.Object) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	world := GetServer().S()
	server.MagicMissileUpdate53BDA0(missile, server.MagicMissileUpdateRuntime53BDA0{
		Frame: world.Frame,
		FPS:   world.TickRate,
		SearchTarget: func(missile, owner *server.Object) *server.Object {
			return world.Nox_xxx_spellFlySearchTarget(nil, missile, things.SpellOffensive, 600, 0, owner)
		},
		Collide: func(missile *server.Object) { missile.CallCollide(0, 0) },
	})
}
