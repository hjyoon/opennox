package legacy

/*
#include "player_mana_recharge_4fd030.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerManaRechargeCall4FD030(unit *server.Object, amount int16) uint16 {
	return GetServer().S().PlayerManaRecharge4FD030(unit, amount, func(unit *server.Object, amount int16) uint16 {
		return playerManaAddCall4EEB80(unit, int32(amount))
	})
}

func playerManaRechargeExportCall4FD030(unit *server.Object, amount int16) uint16 {
	return uint16(C.sub_4FD030(asObjectC(unit), C.int16_t(amount)))
}

//export sub_4FD030
func sub_4FD030(unit *C.nox_object_t, amount C.int16_t) C.uint16_t {
	return C.uint16_t(playerManaRechargeCall4FD030(
		asObjectS((*nox_object_t)(unit)),
		int16(amount),
	))
}
