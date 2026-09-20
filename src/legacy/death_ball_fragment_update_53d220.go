package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var deathBallFragmentUpdateCall53D220 = func(source *server.Object) {
	outer := GetServer()
	outer.S().DeathBallFragmentUpdate53D220(source, server.DeathBallFragmentUpdateRuntime53D220{
		DamageUnitsAround: func(pos types.Pointf, outerRadius, innerRadius float32, damage int, damageType object.DamageType, got *server.Object, excluded server.Obj) {
			outer.Nox_xxx_mapDamageUnitsAround(
				pos, outerRadius, innerRadius, damage, damageType, got, excluded, GetDoDamageWalls(),
			)
		},
		DelayedDelete: outer.DelayedDelete,
	})
}
