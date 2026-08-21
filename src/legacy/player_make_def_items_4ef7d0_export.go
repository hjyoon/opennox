package legacy

/*
#include "player_make_def_items_4ef7d0.h"
*/
import "C"

//export nox_xxx_playerMakeDefItems_4EF7D0
func nox_xxx_playerMakeDefItems_4EF7D0(
	unit *C.nox_object_t,
	restoreStats C.int32_t,
	keepItems C.int32_t,
) C.uint8_t {
	return C.uint8_t(playerMakeDefItemsCall4EF7D0(
		asObjectS((*nox_object_t)(unit)),
		int32(restoreStats),
		int32(keepItems),
	))
}
