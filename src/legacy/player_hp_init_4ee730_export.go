package legacy

/*
#include "player_hp_init_4ee730.h"
*/
import "C"

//export nox_xxx_playerHP_4EE730
func nox_xxx_playerHP_4EE730(unit *C.nox_object_t) {
	playerHPInitCall4EE730(asObjectS((*nox_object_t)(unit)))
}
