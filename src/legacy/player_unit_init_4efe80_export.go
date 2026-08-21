package legacy

/*
#include "player_unit_init_4efe80.h"
*/
import "C"

//export nox_xxx_unitInitPlayer_4EFE80
func nox_xxx_unitInitPlayer_4EFE80(unit *C.nox_object_t) C.uint8_t {
	return C.uint8_t(playerUnitInitCall4EFE80(
		asObjectS((*nox_object_t)(unit)),
	))
}
