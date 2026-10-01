package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func questLoseAbilityCEntry54CFB0(unit *server.Object) int8 {
	return int8(C.sub_54CFB0((*C.nox_object_t)(unsafe.Pointer(unit))))
}

//export nox_server_questLoseAbility_native_54CFB0
func nox_server_questLoseAbility_native_54CFB0(unit *nox_object_t) C.char {
	return C.char(GetServer().S().QuestLoseAbility54CFB0(asObjectS(unit)))
}
