package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_updateBlackPowderBurn_53CCB0_go
func nox_xxx_updateBlackPowderBurn_53CCB0_go(source *nox_object_t) {
	blackPowderBurnUpdateCall53CCB0(asObjectS(source))
}

func blackPowderBurnC53CCB0(source *server.Object) {
	C.nox_xxx_updateBlackPowderBurn_53CCB0(asObjectC(source))
}

func blackPowderBurnUpdateNative53CCB0(source *server.Object) {
	outer := GetServer()
	outer.S().BlackPowderBurnUpdate53CCB0(source, server.BlackPowderBurnUpdateRuntime53CCB0{
		DamageUnitsAround: func(pos types.Pointf, outerRadius, innerRadius float32, damage int, damageType object.DamageType, got *server.Object, excluded server.Obj) {
			outer.Nox_xxx_mapDamageUnitsAround(
				pos, outerRadius, innerRadius, damage, damageType, got, excluded, GetDoDamageWalls(),
			)
		},
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep this indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var blackPowderBurnUpdateCall53CCB0 = blackPowderBurnUpdateNative53CCB0
