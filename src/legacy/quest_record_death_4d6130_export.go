package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func questRecordDeathCall4D6130(unit *server.Object) unsafe.Pointer {
	return server.QuestRecordDeath4D6130(unit)
}

// questRecordDeathBridge4D6130 exercises the retained native C entry point;
// its mixed Object/Player result is an address-sized scalar, never a PE32 int.
func questRecordDeathBridge4D6130(unit *server.Object) uintptr {
	return uintptr(C.sub_4D6130((*C.nox_object_t)(unsafe.Pointer(unit))))
}

//export nox_server_questRecordDeath_native_4D6130
func nox_server_questRecordDeath_native_4D6130(unit *nox_object_t) C.uintptr_t {
	return C.uintptr_t(uintptr(questRecordDeathCall4D6130(asObjectS(unit))))
}
