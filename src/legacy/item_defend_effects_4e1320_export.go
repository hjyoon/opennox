package legacy

/*
#include "item_defend_effects_4e1320.h"
*/
import "C"

import (
	"log/slog"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_itemApplyDefendEffect2_4E1320
func nox_xxx_itemApplyDefendEffect2_4E1320(targetp, sourcep, weaponp *C.nox_object_t, damagep *C.int32_t, typ C.int32_t) C.int {
	return C.int(server.ItemDefendEffects4E1320(
		asObjectS((*nox_object_t)(targetp)), asObjectS((*nox_object_t)(sourcep)), asObjectS((*nox_object_t)(weaponp)),
		(*int32)(unsafe.Pointer(damagep)), int32(typ), server.ItemDefendEffectsRuntime4E1320{
			ApplyDefend: func(effect *server.ModifierEff, item, target, weapon, source *server.Object, context *[2]int32) {
				if !playerDamageCanApplyLateDefendNative4E1320(effect) {
					// An arbitrary PE32 callback cannot receive these LP64
					// records. Keep it visible without a C fallback.
					s := GetServer().S()
					if s.Log != nil {
						s.Log.Error("equipped Defend native callback is not ported",
							slog.String("procedure", "004E1320"), slog.Int64("damage", int64(context[0])), slog.Int64("damage_type", int64(context[1])))
					}
					return
				}
				context[0] = playerDamageApplyLateDefendNative4E1320(effect, item, target, weapon, source, context[0], object.DamageType(context[1]))
			},
		},
	))
}

func Nox_xxx_itemApplyDefendEffect2_4E1320(target, source, weapon *server.Object, damage *int32, typ object.DamageType) int32 {
	return int32(C.nox_xxx_itemApplyDefendEffect2_4E1320(
		(*C.nox_object_t)(asObjectC(target)), (*C.nox_object_t)(asObjectC(source)), (*C.nox_object_t)(asObjectC(weapon)),
		(*C.int32_t)(unsafe.Pointer(damage)), C.int32_t(typ),
	))
}
