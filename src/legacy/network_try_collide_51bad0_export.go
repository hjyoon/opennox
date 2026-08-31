package legacy

/*
#include "network_try_collide_51bad0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_server_netTryCollide_51BAD0
func nox_server_netTryCollide_51BAD0(
	data *C.uchar,
	unit *C.nox_object_t,
	update unsafe.Pointer,
) C.int {
	return C.int(networkTryCollideCall51BAD0(
		(*[server.NetworkTryCollidePacketSize51BAD0]byte)(unsafe.Pointer(data)),
		asObjectS((*nox_object_t)(unit)),
		(*server.PlayerUpdateData)(update),
	))
}
