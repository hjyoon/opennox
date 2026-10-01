package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func questLoseSpellCEntry54CE00(unit *server.Object) {
	C.sub_54CE00((*C.nox_object_t)(unsafe.Pointer(unit)))
}

//export nox_server_questLoseSpell_native_54CE00
func nox_server_questLoseSpell_native_54CE00(unit *nox_object_t) {
	GetServer().S().QuestLoseSpell54CE00(asObjectS(unit))
}
