package legacy

/*
#include "player_confused_direction_4f7a40.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerConfusedDirectionExportCall4F7A40(unit *server.Object) server.Dir16 {
	return server.Dir16(C.nox_xxx_playerConfusedGetDirection_4F7A40(asObjectC(unit)))
}

//export nox_xxx_playerConfusedGetDirection_4F7A40
func nox_xxx_playerConfusedGetDirection_4F7A40(unit *C.nox_object_t) C.int {
	return C.int(GetServer().S().PlayerConfusedDirection4F7A40(
		asObjectS((*nox_object_t)(unit)),
	))
}
