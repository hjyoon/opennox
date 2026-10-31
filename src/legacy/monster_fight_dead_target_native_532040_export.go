package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(sizeof(uintptr_t) == sizeof(void*), "dead-target result retains the complete native pointer");
_Static_assert(_Generic(&sub_532040, uintptr_t (*)(nox_object_t*, void*): 1, default: 0), "dead-target native candidate/context ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export sub_532040
func sub_532040(candidate *C.nox_object_t, context unsafe.Pointer) C.uintptr_t {
	unit := asObjectS((*nox_object_t)(candidate))
	if server.MonsterFightDeadTarget532040(unit, (*server.MonsterUpdateData)(context)) {
		*memmap.PtrUint32(0x5D4594, 2487944) = 1
	}
	return C.uintptr_t(uintptr(unsafe.Pointer(candidate)))
}
