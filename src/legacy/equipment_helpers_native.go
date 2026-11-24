package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func equipFirstInventoryWeapon(owner *server.Object, matches func(uint32) bool) C.int {
	if owner == nil {
		return 0
	}
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		flags := uint32(nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(item)))
		if matches(flags) {
			return C.int(playerEquipWeaponNativeCall53A420(owner, item))
		}
	}
	return 0
}

//export sub_539FB0
func sub_539FB0(owner *nox_object_t) C.int {
	return equipFirstInventoryWeapon(asObjectS(owner), func(flags uint32) bool { return flags == 128 })
}

//export nox_xxx_playerTryReloadQuiver_539FF0
func nox_xxx_playerTryReloadQuiver_539FF0(owner *nox_object_t) C.int {
	return equipFirstInventoryWeapon(asObjectS(owner), func(flags uint32) bool { return flags == 2 })
}

//export sub_53A680
func sub_53A680(owner *nox_object_t) C.int {
	return equipFirstInventoryWeapon(asObjectS(owner), func(flags uint32) bool { return flags&0xC != 0 })
}

//export sub_53A6C0
func sub_53A6C0(owner, item *nox_object_t) {
	Nox_xxx_pickupWeaponAudio_53A6C0(asObjectS(owner), asObjectS(item))
}

//export sub_53E2D0
func sub_53E2D0(item *nox_object_t) C.int {
	o := asObjectS(item)
	if !o.ObjClass.Has(object.ClassArmor) {
		return 1
	}
	return C.int(bool2int(nox_xxx_unitArmorInventoryEquipFlags_415C70(item)&0xC0D == 0))
}
