package legacy

/*
#include "boom_collide_4e9770.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideBoom_4E9770
func nox_xxx_collideBoom_4E9770(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().BoomCollide4E9770(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.BoomCollideRuntime4E9770{
			CheckDirection: func(first types.Pointf, direction int16, second types.Pointf) int32 {
				return twoPointsAndDirection4E6E50(first, int32(direction), second)
			},
			ChangeOwner: Nox_xxx_changeOwner_52BE40,
			Scorch: func(pos types.Pointf, kind int32) {
				Nox_xxx_sMakeScorch_537AF0(pos, int(kind))
			},
			TraceHitPoint: func() *ntype.Point32 {
				if Get_dword_5d4594_2488620() == 0 {
					return nil
				}
				return memmap.PtrT[ntype.Point32](0x5D4594, 2488612)
			},
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			MapDamageUnits: func(
				pos types.Pointf,
				radius, inner float32,
				damage int32,
				damageType object.DamageType,
				source, excluded *server.Object,
			) {
				srv.Nox_xxx_mapDamageUnitsAround(
					pos, radius, inner, int(damage), damageType, source, excluded, GetDoDamageWalls(),
				)
			},
			MapPushUnits: func(
				pos types.Pointf,
				first, second, force float32,
				source *server.Object,
				arg6, arg7 int32,
			) {
				Nox_xxx_mapPushUnitsAround_52E040(pos, first, second, force, source, int(arg6), int(arg7))
			},
			DelayedDelete:   srv.DelayedDelete,
			InversionEffect: InversionEffectPointer4E03D0(),
		},
	)
}
