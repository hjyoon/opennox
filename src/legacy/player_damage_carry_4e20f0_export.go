package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_playerDecrementHPMana_4E20F0
func nox_xxx_playerDecrementHPMana_4E20F0(target *nox_object_t, output *C.int32_t, amount C.float) {
	server.PlayerDamageCarry4E20F0(asObjectS(target), (*int32)(output), float32(amount))
}
