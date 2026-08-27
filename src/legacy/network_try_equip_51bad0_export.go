package legacy

/*
#include "network_try_equip_51bad0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_server_netTryEquip_51BAD0
func nox_server_netTryEquip_51BAD0(
	data *C.uchar,
	unit *C.nox_object_t,
	update unsafe.Pointer,
) C.int {
	return C.int(networkTryEquipCall51BAD0(
		(*[server.NetworkTryEquipPacketSize51BAD0]byte)(unsafe.Pointer(data)),
		asObjectS((*nox_object_t)(unit)),
		(*server.PlayerUpdateData)(update),
	))
}
