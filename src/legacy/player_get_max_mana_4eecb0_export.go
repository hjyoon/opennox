package legacy

/*
#include "player_get_max_mana_4eecb0.h"
*/
import "C"

//export nox_xxx_playerGetMaxMana_4EECB0
func nox_xxx_playerGetMaxMana_4EECB0(unit *C.nox_object_t) C.short {
	return C.short(int16(Nox_xxx_playerGetMaxMana_4EECB0(asObjectS((*nox_object_t)(unit)))))
}
