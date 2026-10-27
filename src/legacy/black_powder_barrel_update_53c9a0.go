package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_updateBlackPowderBarrel_53C9A0_go
func nox_xxx_updateBlackPowderBarrel_53C9A0_go(source *nox_object_t) {
	blackPowderBarrelUpdateCall53C9A0(asObjectS(source))
}

func blackPowderBarrelC53C9A0(source *server.Object) {
	C.nox_xxx_updateBlackPowderBarrel_53C9A0(asObjectC(source))
}

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
		RandomInt:   world.Rand.Logic.IntClamp,
		RandomFloat: world.RandomFloat416030,
		TraceRay: func(from, to types.Pointf) bool {
			return world.MapTraceRay(from, to, server.MapTraceFlag1)
		},
		LoadFlameType:        func(index int) int { return int(int32(memmap.Uint32(0x5D4594, 2488684+uintptr(index)*4))) },
		StoreFlameType:       func(index, value int) { *memmap.PtrUint32(0x5D4594, 2488684+uintptr(index)*4) = uint32(value) },
		TypeIndexByID:        world.Types.IndByID,
		NewObjectByTypeIndex: world.NewObjectByTypeInd,
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
