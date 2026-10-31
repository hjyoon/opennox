package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(_Generic(&sub_532880, uintptr_t (*)(nox_object_t*): 1, default: 0), "native hit-latch ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_532880
func sub_532880(unit *C.nox_object_t) C.uintptr_t {
	return C.uintptr_t(uintptr(unsafe.Pointer(server.MonsterHitLatch532880(asObjectS((*nox_object_t)(unit)), gameFrame))))
}
