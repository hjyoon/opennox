package legacy

/*
#include "network_inventory_fail_51bad0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_server_netInventoryFail_51BAD0
func nox_server_netInventoryFail_51BAD0(
	data *C.uchar,
	unit *C.nox_object_t,
) C.int {
	return C.int(networkInventoryFailCall51BAD0(
		(*[server.NetworkInventoryFailPacketSize51BAD0]byte)(unsafe.Pointer(data)),
		asObjectS((*nox_object_t)(unit)),
	))
}
