package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func antiSpellProjectileUpdateCall53BB00(missile *server.Object) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	outer := GetServer()
	world := outer.S()
	server.AntiSpellProjectileUpdate53BB00(missile, server.AntiSpellProjectileRuntime53BB00{
		Frame: world.Frame,
		FPS:   world.TickRate,
		EachMissile: func(pos types.Pointf, radius float32, fnc func(*server.Object) bool) {
			world.Map.EachMissileInCircle(pos, radius, fnc)
		},
		CanInteract: func(source, target *server.Object) bool {
			return world.CanInteract(source, target, 0)
		},
		DelayedDelete: outer.DelayedDelete,
		Audio: func(missile *server.Object) {
			world.Audio.EventObj(sound.ID(20), missile, 0, 0)
		},
		RandomFloat: func(min, max float32) float32 {
			return float32(world.Rand.Logic.FloatClamp(float64(min), float64(max)))
		},
		RandomInt: world.Rand.Logic.IntClamp,
		CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, owner *server.Object) {
			Nox_xxx_createSpark_54FD80(pos.X, pos.Y, kind, lifetime, velocity.X, velocity.Y, z, owner)
		},
	})
}
