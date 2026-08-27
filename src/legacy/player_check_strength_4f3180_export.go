package legacy

/*
#include "player_check_strength_4f3180.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerCheckStrengthCall4F3180(player, item *server.Object) int32 {
	return GetServer().S().PlayerCheckStrength4F3180(player, item)
}

func playerCheckStrengthExportCall4F3180(player, item *server.Object) int32 {
	return int32(C.nox_xxx_playerCheckStrength_4F3180(
		asObjectC(player),
		asObjectC(item),
	))
}

//export nox_xxx_playerCheckStrength_4F3180
func nox_xxx_playerCheckStrength_4F3180(
	player, item *C.nox_object_t,
) C.int32_t {
	return C.int32_t(playerCheckStrengthCall4F3180(
		asObjectS((*nox_object_t)(player)),
		asObjectS((*nox_object_t)(item)),
	))
}
