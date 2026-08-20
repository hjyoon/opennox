package legacy

/*
#include "player_respawn_item_4ef750.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerRespawnItem_4EF750
func nox_xxx_playerRespawnItem_4EF750(
	player *C.nox_object_t,
	typeID *C.char,
	attrs *C.nox_modifier_attrs_t,
	a4 C.int32_t,
	a5 C.int32_t,
) *C.nox_object_t {
	item := playerRespawnItemCall4EF750(
		asObjectS((*nox_object_t)(player)),
		GoString(typeID),
		(*server.ModifierInitData)(unsafe.Pointer(attrs)),
		int32(a4),
		int32(a5),
	)
	return (*C.nox_object_t)(asObjectC(item))
}
