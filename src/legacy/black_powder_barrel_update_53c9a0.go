package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func blackPowderBarrelUpdateNative53C9A0(source *server.Object) {
	outer := GetServer()
	world := outer.S()
	world.BlackPowderBarrelUpdate53C9A0(source, server.BlackPowderBarrelUpdateRuntime53C9A0{
		DamageUnitsAround: func(position types.Pointf, outerRadius, innerRadius float32, damage int, damageType object.DamageType, got *server.Object, excluded server.Obj) {
			outer.Nox_xxx_mapDamageUnitsAround(
				position, outerRadius, innerRadius, damage, damageType, got, excluded, GetDoDamageWalls(),
			)
		},
		PushUnitsAround: func(position types.Pointf, outerRadius, innerRadius, force float32) {
			world.MapPushUnitsAround52E040(position, outerRadius, innerRadius, force, server.MapPushUnitsAroundRuntime52E040{
				ApplyForce: outer.ApplyForce,
			})
		},
		RandomInt: world.Rand.Logic.IntClamp,
		RandomFloat: func(minimum, maximum float32) float32 {
			return float32(world.Rand.Logic.FloatClamp(float64(minimum), float64(maximum)))
		},
		TraceRay: func(from, to types.Pointf) bool {
			return world.MapTraceRay(from, to, server.MapTraceFlag1)
		},
		NewObjectByTypeID: world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, owner, position)
		},
		SetDecayTime: func(obj *server.Object, frames uint32) {
			world.DecaySetTime511660(obj, frames)
		},
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep this indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var blackPowderBarrelUpdateCall53C9A0 = blackPowderBarrelUpdateNative53C9A0
