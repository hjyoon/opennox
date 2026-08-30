package legacy

/*
#include "player_adjust_stamina_4f7db0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerAdjustStaminaExportCall4F7DB0(unit *server.Object, amount uint8) {
	C.sub_4F7DB0(asObjectC(unit), C.uint8_t(amount))
}

//export sub_4F7DB0
func sub_4F7DB0(unit *C.nox_object_t, amount C.uint8_t) {
	GetServer().S().PlayerAdjustStamina4F7DB0(
		asObjectS((*nox_object_t)(unit)),
		uint8(amount),
	)
}
