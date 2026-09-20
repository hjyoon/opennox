package legacy

/*
#include <stdint.h>
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var pickupCollideCall4E8DF0 = func(item, unit *server.Object, collision unsafe.Pointer) uintptr {
	return GetServer().S().PickupCollide4E8DF0(
		item,
		unit,
		collision,
		server.PickupCollideRuntime4E8DF0{
			PlaceInventory: Nox_xxx_inventoryServPlace_4F36F0,
		},
	)
}

//export nox_xxx_collidePickup_4E8DF0
func nox_xxx_collidePickup_4E8DF0(
	item, unit *C.nox_object_t,
	collision *C.float,
) C.uintptr_t {
	return C.uintptr_t(pickupCollideCall4E8DF0(
		asObjectS((*nox_object_t)(item)),
		asObjectS((*nox_object_t)(unit)),
		unsafe.Pointer(collision),
	))
}
