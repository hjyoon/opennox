package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netSendComplexObject_518960
func nox_xxx_netSendComplexObject_518960(player C.int, unit *C.uint32_t, report C.int) C.int {
	return C.int(NetSendComplexObject518960(int32(player), (*server.Object)(unsafe.Pointer(unit)), int32(report)))
}
