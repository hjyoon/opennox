package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func dispatchClientPlayerRespawn417680(
	player *server.Player,
	equipmentMask uint8,
	respawn func(*server.Player, uint8),
) {
	respawn(player, equipmentMask)
}

//export nox_xxx_cliPlayerRespawn_native_417680
func nox_xxx_cliPlayerRespawn_native_417680(player unsafe.Pointer, equipmentMask C.uint8_t) {
	dispatchClientPlayerRespawn417680(
		(*server.Player)(player),
		uint8(equipmentMask),
		GetServer().S().ClientPlayerRespawn417680,
	)
}
