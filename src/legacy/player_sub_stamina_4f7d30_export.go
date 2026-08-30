package legacy

/*
#include "player_sub_stamina_4f7d30.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerSubStaminaExportCall4F7D30(unit *server.Object, amount int32) int32 {
	return int32(C.nox_xxx_playerSubStamina_4F7D30(asObjectC(unit), C.int32_t(amount)))
}

//export nox_xxx_playerSubStamina_4F7D30
func nox_xxx_playerSubStamina_4F7D30(unit *C.nox_object_t, amount C.int32_t) C.int32_t {
	return C.int32_t(GetServer().S().PlayerSubStamina4F7D30(
		asObjectS((*nox_object_t)(unit)),
		int32(amount),
	))
}
