package legacy

/*
#include "player_mana_sub_4eebf0.h"
*/
import "C"

//export nox_xxx_playerManaSub_4EEBF0
func nox_xxx_playerManaSub_4EEBF0(unit *C.nox_object_t, amount C.int32_t) C.uintptr_t {
	return C.uintptr_t(playerManaSubCall4EEBF0(
		asObjectS((*nox_object_t)(unit)),
		int32(amount),
	))
}
