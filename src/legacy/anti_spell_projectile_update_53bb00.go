package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

var antiSpellProjectileRuntime53BB00 = func() server.AntiSpellProjectileRuntime53BB00 {
	outer := GetServer()
	world := outer.S()
	return server.AntiSpellProjectileRuntime53BB00{
		TargetState:    world.AntiSpellTargetState53BD10(),
		Frame:          world.Frame,
		FPS:            world.TickRate,
		EachMissilePtr: world.Map.EachMissileInCirclePtr518170,
		CanInteract: func(source, target *server.Object) bool {
			return world.CanInteract(source, target, 0)
		},
		DelayedDelete: outer.DelayedDelete,
		Audio: func(missile *server.Object) {
			world.Audio.EventObj(sound.ID(20), missile, 0, 0)
		},
		RandomFloatWide: world.RandomFloat416030,
		RandomInt:       world.Rand.Logic.IntClamp,
		CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, owner *server.Object) {
			Nox_xxx_createSpark_54FD80(pos.X, pos.Y, kind, lifetime, velocity.X, velocity.Y, z, owner)
		},
	}
}

func antiSpellProjectileUpdateCall53BB00(missile *server.Object) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	server.AntiSpellProjectileUpdate53BB00(missile, antiSpellProjectileRuntime53BB00())
}

//export nox_anti_spell_update_native_53BB00
func nox_anti_spell_update_native_53BB00(missile *nox_object_t) {
	server.AntiSpellProjectileRaw53BB00(asObjectS(missile), antiSpellProjectileRuntime53BB00())
}

//export nox_anti_spell_candidate_native_53BD10
func nox_anti_spell_candidate_native_53BD10(candidate, missile *nox_object_t) {
	server.AntiSpellCandidateRaw53BD10(asObjectS(candidate), asObjectS(missile), antiSpellProjectileRuntime53BB00())
}
