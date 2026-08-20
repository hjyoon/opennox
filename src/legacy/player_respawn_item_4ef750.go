package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func playerRespawnItemRuntime4EF750() server.PlayerRespawnItemRuntime4EF750 {
	return server.PlayerRespawnItemRuntime4EF750{
		ApplyModifierAttrs: func(item *server.Object, attrs *server.ModifierInitData) {
			C.nox_xxx_modifSetItemAttrs_4E4990(
				asObjectC(item),
				(*C.nox_modifier_attrs_t)(unsafe.Pointer(attrs)),
			)
		},
		PlaceInventory: func(player, item *server.Object, a4, a5 int32) bool {
			return Nox_xxx_inventoryServPlace_4F36F0(
				player,
				item,
				int(a4),
				int(a5),
			)
		},
	}
}

func playerRespawnItemCall4EF750(
	player *server.Object,
	typeID string,
	attrs *server.ModifierInitData,
	a4 int32,
	a5 int32,
) *server.Object {
	return GetServer().S().PlayerRespawnItem4EF750(
		player,
		typeID,
		attrs,
		a4,
		a5,
		playerRespawnItemRuntime4EF750(),
	)
}
