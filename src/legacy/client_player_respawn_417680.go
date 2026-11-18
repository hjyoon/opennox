package legacy

/*
#include "GAME1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func dispatchClientPlayerRespawn417680(
	player *server.Player,
	equipmentMask uint8,
	respawn func(*server.Player, uint8),
) {
	respawn(player, equipmentMask)
}

//export nox_xxx_cliPlayerRespawn_417680
func nox_xxx_cliPlayerRespawn_417680(player *nox_playerInfo, equipmentMask C.uint8_t) {
	dispatchClientPlayerRespawn417680(
		asPlayerS(player),
		uint8(equipmentMask),
		GetServer().S().ClientPlayerRespawn417680,
	)
}
