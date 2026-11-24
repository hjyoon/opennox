package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func equipmentDropServerSubFlag53AB10(flag uint32) int32 {
	if Nox_xxx_getServerSubFlags_409E60()&flag != 0 {
		return 1
	}
	return 0
}

var weaponDropCall53AB10 = func(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().WeaponDrop53AB10(owner, item, point, server.WeaponDropRuntime53AB10{
		DefaultDrop:   defaultDropCall4ED290,
		ServerSubFlag: equipmentDropServerSubFlag53AB10,
	})
}

var armorDropCall53EB70 = func(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().ArmorDrop53EB70(owner, item, point, server.ArmorDropRuntime53EB70{
		DefaultDrop:   defaultDropCall4ED290,
		ServerSubFlag: equipmentDropServerSubFlag53AB10,
	})
}

//export nox_xxx_dropWeapon_53AB10
func nox_xxx_dropWeapon_53AB10(owner, item *nox_object_t, point *C.float2) C.int {
	return C.int(weaponDropCall53AB10(asObjectS(owner), asObjectS(item), (*types.Pointf)(unsafe.Pointer(point))))
}

//export nox_xxx_dropArmor_53EB70
func nox_xxx_dropArmor_53EB70(owner, item *nox_object_t, point *C.float2) C.int {
	return C.int(armorDropCall53EB70(asObjectS(owner), asObjectS(item), (*types.Pointf)(unsafe.Pointer(point))))
}

//export sub_53AAB0
func sub_53AAB0(item *nox_object_t) {
	if item != nil {
		GetServer().S().WeaponDropSound53AAB0(asObjectS(item))
	}
}

//export sub_53EAE0
func sub_53EAE0(item *nox_object_t) {
	if item != nil {
		GetServer().S().ArmorDropSound53EAE0(asObjectS(item))
	}
}
