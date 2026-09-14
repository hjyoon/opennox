package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func projectileTrailUpdateCall53AEC0(projectile *server.Object) {
	if projectile == nil {
		return
	}
	world := GetServer().S()
	server.ProjectileTrailUpdate53AEC0(projectile, server.ProjectileTrailRuntime53AEC0{
		RandomFloat: func(min, max float32) float32 {
			return float32(world.Rand.Logic.FloatClamp(float64(min), float64(max)))
		},
		CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, owner *server.Object) {
			Nox_xxx_createSpark_54FD80(pos.X, pos.Y, kind, lifetime, velocity.X, velocity.Y, z, owner)
		},
	})
}
