package legacy

/*
#include "player_set_max_mana_4eecd0.h"
*/
import "C"

//export nox_xxx_playerSetMaxMana_4EECD0
func nox_xxx_playerSetMaxMana_4EECD0(unit *C.nox_object_t, maximum C.short) C.uintptr_t {
	return C.uintptr_t(Nox_xxx_playerSetMaxMana_4EECD0(
		asObjectS((*nox_object_t)(unit)),
		uint16(maximum),
	))
}
