package legacy

/*
#include "GAME4_3.h"
*/
import "C"
import "unsafe"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_mobCalcDir_533CC0(unit *server.Object, position *types.Pointf) {
	server.MonsterCastDirection533CC0(unit, position)
}

//export nox_xxx_mobCalcDir_533CC0
func nox_xxx_mobCalcDir_533CC0(unit *C.nox_object_t, position *C.float2) {
	Nox_xxx_mobCalcDir_533CC0(asObjectS((*nox_object_t)(unit)), (*types.Pointf)(unsafe.Pointer(position)))
}
