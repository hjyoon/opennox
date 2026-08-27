package legacy

/*
#include "GAME3_2.h"
#include "secondary_weapon_53ab90.h"
*/
import "C"

import (
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

func secondaryWeaponExportCall53AB90(owner, item *server.Object) {
	C.sub_53AB90(asObjectC(owner), asObjectC(item))
}

//export sub_53AB90
func sub_53AB90(owner, item *C.nox_object_t) {
	GetServer().S().SecondaryWeaponReport53AB90(
		asObjectS(owner),
		asObjectS(item),
		func(item *server.Object, class player.Class) bool {
			return Nox_xxx_playerClassCanUseItem_57B3D0(item, class)
		},
		Nox_xxx_playerCheckStrength_4F3180,
		func(index byte) {
			C.nox_xxx_netSendSecondaryWeapon_4D9670(C.int(index), (*C.uint)(nil), 1)
		},
	)
}
