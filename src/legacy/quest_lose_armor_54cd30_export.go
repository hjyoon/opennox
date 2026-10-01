package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func questLoseArmorCEntry54CD30(unit *server.Object) {
	C.sub_54CD30((*C.nox_object_t)(unsafe.Pointer(unit)))
}

//export nox_server_questLoseArmor_native_54CD30
func nox_server_questLoseArmor_native_54CD30(unit *nox_object_t) {
	outer := GetServer()
	outer.S().QuestLoseArmor54CD30(asObjectS(unit), outer.DelayedDelete)
}
