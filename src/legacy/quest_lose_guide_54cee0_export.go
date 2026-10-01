package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func questLoseGuideCEntry54CEE0(unit *server.Object) {
	C.sub_54CEE0((*C.nox_object_t)(unsafe.Pointer(unit)))
}

//export nox_server_questLoseGuide_native_54CEE0
func nox_server_questLoseGuide_native_54CEE0(unit *nox_object_t) {
	GetServer().S().QuestLoseGuide54CEE0(asObjectS(unit))
}
