package legacy

/*
#include "player_try_dequip_4f2fb0.h"
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerTryDequipCall4F2FB0(owner, item *server.Object) int32 {
	return server.PlayerTryDequip4F2FB0(
		owner,
		item,
		func(owner, item *server.Object, flag1, flag2 int32) int32 {
			return int32(C.nox_xxx_playerDequipWeapon_53A140(
				asObjectC(owner),
				asObjectC(item),
				C.int(flag1),
				C.int(flag2),
			))
		},
		func(owner, item *server.Object, flag1, flag2 int32) int32 {
			return int32(C.sub_53E430(
				asObjectC(owner),
				asObjectC(item),
				C.int(flag1),
				C.int(flag2),
			))
		},
	)
}

func Nox_xxx_playerTryDequip_4F2FB0(owner, item *server.Object) bool {
	return playerTryDequipCall4F2FB0(owner, item) != 0
}

//export nox_xxx_playerTryDequip_4F2FB0_go
func nox_xxx_playerTryDequip_4F2FB0_go(
	owner, item *C.nox_object_t,
) C.int32_t {
	return C.int32_t(playerTryDequipCall4F2FB0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	))
}
