package legacy

/*
#include "player_mana_add_4eeb80.h"
*/
import "C"

//export nox_xxx_playerManaAdd_4EEB80
func nox_xxx_playerManaAdd_4EEB80(unit *C.nox_object_t, amount C.int16_t) C.uint16_t {
	return C.uint16_t(playerManaAddCall4EEB80(
		asObjectS((*nox_object_t)(unit)),
		int32(int16(amount)),
	))
}
