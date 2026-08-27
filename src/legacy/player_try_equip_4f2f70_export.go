package legacy

/*
#include "player_try_equip_4f2f70.h"
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerTryEquipCall4F2F70(owner, item *server.Object) int32 {
	return server.PlayerTryEquip4F2F70(
		owner,
		item,
		func(owner, item *server.Object, flag1, flag2 int32) int32 {
			return int32(C.nox_xxx_playerEquipWeapon_53A420(
				asObjectC(owner),
				asObjectC(item),
				C.int(flag1),
				C.int(flag2),
			))
		},
		func(owner, item *server.Object, flag1, flag2 int32) int32 {
			return int32(C.nox_xxx_playerEquipArmor_53E650(
				asObjectC(owner),
				asObjectC(item),
				C.int(flag1),
				C.int(flag2),
			))
		},
	)
}

func Nox_xxx_playerTryEquip_4F2F70(owner, item *server.Object) bool {
	return playerTryEquipCall4F2F70(owner, item) != 0
}

//export nox_xxx_playerTryEquip_4F2F70
func nox_xxx_playerTryEquip_4F2F70(
	owner, item *C.nox_object_t,
) C.int32_t {
	return C.int32_t(playerTryEquipCall4F2F70(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	))
}
