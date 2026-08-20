package legacy

/*
#include "player_sync_level_4ef140.h"
*/
import "C"

//export sub_4EF140
func sub_4EF140(unit *C.nox_object_t) C.int32_t {
	return C.int32_t(playerSyncLevelCall4EF140(
		asObjectS((*nox_object_t)(unit)),
	))
}
