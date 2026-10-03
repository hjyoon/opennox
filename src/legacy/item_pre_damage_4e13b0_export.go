package legacy

/*
#include "item_pre_damage_4e13b0.h"
*/
import "C"

import (
	"log/slog"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_itemApplyPreDamageEffect_4E13B0
func nox_xxx_itemApplyPreDamageEffect_4E13B0(targetp, sourcep, weaponp *C.nox_object_t, damagep *C.int32_t) C.int {
	return C.int(server.ItemPreDamage4E13B0(
		asObjectS((*nox_object_t)(targetp)), asObjectS((*nox_object_t)(sourcep)), asObjectS((*nox_object_t)(weaponp)),
		(*int32)(unsafe.Pointer(damagep)), server.ItemPreDamageRuntime4E13B0{
			ApplyPreDamage: func(effect *server.ModifierEff, weapon, source, target *server.Object, damage *int32) {
				if !itemPreDamageCanApplyNative4E13B0(effect) {
					// Never jump to an unported PE32 callback with LP64
					// records. Report it at use and continue the live slots.
					s := GetServer().S()
					if s.Log != nil {
						fields := []any{slog.String("procedure", "004E13B0"), slog.Bool("damage_present", damage != nil)}
						if damage != nil {
							fields = append(fields, slog.Int64("damage", int64(*damage)))
						}
						s.Log.Error("weapon pre-damage native callback is not ported", fields...)
					}
					return
				}
				// The helper passes even a nil damage address. Poison does
				// not use this argument; each stock callback owns its reads.
				itemPreDamageApplyNative4E13B0(effect, weapon, source, target, damage)
			},
		},
	))
}

func Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon *server.Object, damage *int32) int32 {
	return int32(C.nox_xxx_itemApplyPreDamageEffect_4E13B0(
		(*C.nox_object_t)(asObjectC(target)), (*C.nox_object_t)(asObjectC(source)), (*C.nox_object_t)(asObjectC(weapon)),
		(*C.int32_t)(unsafe.Pointer(damage)),
	))
}
