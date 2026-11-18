package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var monsterWeaponSearchState544AE0 server.MonsterWeaponSearchState544AE0

//export nox_xxx_mobActionHunt_5449D0
func nox_xxx_mobActionHunt_5449D0(unit *nox_object_t) *C.int {
	item := GetServer().S().MonsterActionHuntNative5449D0(asObjectS(unit))
	return (*C.int)(unsafe.Pointer(item))
}

//export sub_544AE0
func sub_544AE0(unit *nox_object_t, radius C.float) *nox_object_t {
	result := GetServer().S().MonsterSearchWeaponNative544AE0(asObjectS(unit), float32(radius),
		&monsterWeaponSearchState544AE0, Nox_xxx_playerClassCanUseItem_57B3D0)
	return asObjectC(result)
}

//export sub_544B20
func sub_544B20(candidate, unit *nox_object_t) {
	GetServer().S().MonsterSearchWeaponCandidateNative544B20(asObjectS(candidate), asObjectS(unit),
		&monsterWeaponSearchState544AE0, Nox_xxx_playerClassCanUseItem_57B3D0)
}
