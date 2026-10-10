package legacy

/*
#include "player_handle_elim_death_54d7a0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func Nox_xxx_playerHandleElimDeathNative54D7A0(victim, killer *server.Object) {
	C.nox_player_handle_elim_death_native_call_54D7A0(asObjectC(victim), asObjectC(killer))
}

//export nox_server_player_handle_elim_death_native_54D7A0
func nox_server_player_handle_elim_death_native_54D7A0(victim, killer *nox_object_t) {
	server.PlayerHandleElimDeath54D7A0(asObjectS(victim), asObjectS(killer), playerUpdateScoreRuntime54D980(GetServer()))
}
