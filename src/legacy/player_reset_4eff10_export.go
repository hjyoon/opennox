package legacy

/*
#include "player_reset_4eff10.h"
*/
import "C"

//export sub_4EFF10
func sub_4EFF10(unit *C.nox_object_t) C.int32_t {
	return C.int32_t(playerResetCall4EFF10(
		asObjectS((*nox_object_t)(unit)),
	))
}
