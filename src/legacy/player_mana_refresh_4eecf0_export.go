package legacy

/*
#include "player_mana_refresh_4eecf0.h"
*/
import "C"

//export nox_xxx_playerManaRefresh_4EECF0
func nox_xxx_playerManaRefresh_4EECF0(unit *C.nox_object_t) C.uintptr_t {
	return C.uintptr_t(playerManaRefreshCall4EECF0(
		asObjectS((*nox_object_t)(unit)),
	))
}
