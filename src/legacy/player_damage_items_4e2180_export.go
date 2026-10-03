package legacy

/*
#include "GAME1.h"
#include "player_damage_items_4e2180.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func playerDamageItemsNative4E2180(target, source, effective *server.Object, damage int32, typ object.DamageType) {
	server.PlayerDamageItems4E2180(target, source, effective, damage, typ, server.PlayerDamageItemsRuntime4E2180{
		ItemArmorValue: func(item *server.Object) float64 {
			return float64(C.nox_xxx_itemApplyDefendEffect_415C00(asObjectC(item)))
		},
		EquipDamage: func(item, owner, source, effective *server.Object, amount float32, typ object.DamageType) {
			equipDamageNative4E16D0(item, owner, source, effective, amount, typ)
		},
	})
}

//export nox_xxx_playerDamageItems_4E2180
func nox_xxx_playerDamageItems_4E2180(target, source, effective *C.nox_object_t, damage, typ C.int32_t) {
	playerDamageItemsNative4E2180(
		asObjectS((*nox_object_t)(target)), asObjectS((*nox_object_t)(source)), asObjectS((*nox_object_t)(effective)),
		int32(damage), object.DamageType(typ),
	)
}

func Nox_xxx_playerDamageItems_4E2180(target, source, effective *server.Object, damage int32, typ object.DamageType) {
	C.nox_xxx_playerDamageItems_4E2180(
		(*C.nox_object_t)(asObjectC(target)), (*C.nox_object_t)(asObjectC(source)), (*C.nox_object_t)(asObjectC(effective)),
		C.int32_t(damage), C.int32_t(typ),
	)
}
