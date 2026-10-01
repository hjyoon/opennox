package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/server"
)

func questLoseWeaponCEntry54CC40(unit *server.Object) {
	C.sub_54CC40(asObjectC(unit))
}

//export nox_server_questLoseWeapon_native_54CC40
func nox_server_questLoseWeapon_native_54CC40(unit *nox_object_t) {
	outer := GetServer()
	outer.S().QuestLoseWeapon54CC40(asObjectS(unit), func(item *server.Object, class uint8) int32 {
		return int32(bool2int(Nox_xxx_playerClassCanUseItem_57B3D0(item, player.Class(class))))
	}, outer.DelayedDelete)
}
