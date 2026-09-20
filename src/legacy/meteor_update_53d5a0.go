package legacy

import (
	"image"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func meteorShowerUpdateNative53D5A0(source *server.Object) {
	outer := GetServer()
	world := outer.S()
	world.MeteorShowerUpdate53D5A0(source, server.MeteorShowerUpdateRuntime53D5A0{
		RandomInt: world.Rand.Logic.IntClamp,
		RandomFloat: func(minimum, maximum float64) float64 {
			return world.Rand.Logic.FloatClamp(minimum, maximum)
		},
		TraceRay:          world.MapTraceRay9,
		NewObjectByTypeID: world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, owner, position)
		},
		Raise:         (*server.Object).Raise,
		DelayedDelete: outer.DelayedDelete,
	})
}

func meteorUpdateNative53D6E0(source *server.Object) {
	outer := GetServer()
	world := outer.S()
	world.MeteorUpdate53D6E0(source, server.MeteorUpdateRuntime53D6E0{
		AudioEvent: func(id uint32, object *server.Object) {
			world.Audio.EventObj(sound.ID(id), object, 0, 0)
		},
		MakeScorch:        Nox_xxx_sMakeScorch_537AF0,
		Earthquake:        world.Nox_xxx_earthquakeSend_4D9110,
		NewObjectByTypeID: world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, owner, position)
		},
		FindParentChainPlayer: (*server.Object).FindOwnerChainPlayer,
		DamageUnitsAround: func(position types.Pointf, outerRadius, innerRadius float32, damage int, damageType object.DamageType, source *server.Object, excluded server.Obj, damageWalls bool) {
			outer.Nox_xxx_mapDamageUnitsAround(
				position, outerRadius, innerRadius, damage, damageType, source, excluded, damageWalls,
			)
		},
		DamageWalls: func(rect image.Rectangle, position types.Pointf, radius float32, damage int, damageType object.DamageType, source *server.Object) {
			outer.Nox_xxx_mapDamageToWalls_534FC0(rect, position, radius, damage, damageType, source)
		},
		DamageSource:  GetDoDamageWalls,
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep these indirections dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback addresses stored in thing.bin objects.
var (
	meteorShowerUpdateCall53D5A0 = meteorShowerUpdateNative53D5A0
	meteorUpdateCall53D6E0       = meteorUpdateNative53D6E0
)
