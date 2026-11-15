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

//export nox_xxx_mobCast_541300
func nox_xxx_mobCast_541300(id C.int, unit *C.nox_object_t, arg *C.nox_spell_accept_arg_t) C.char {
	return C.char(monsterCastCall541300(int32(id), asObjectS((*nox_object_t)(unit)), (*server.SpellAcceptArg)(unsafe.Pointer(arg))))
}

//export nox_xxx_mobActionCast_5413B0
func nox_xxx_mobActionCast_5413B0(unit *C.nox_object_t, mode C.int) C.char {
	return C.char(monsterActionCastCall5413B0(asObjectS((*nox_object_t)(unit)), int32(mode)))
}

//export nox_xxx_mobCastRandomRecoil_541490
func nox_xxx_mobCastRandomRecoil_541490(unit, target *C.nox_object_t, out *C.float2) {
	monsterCastRandomRecoilCall541490(asObjectS((*nox_object_t)(unit)), asObjectS((*nox_object_t)(target)), (*types.Pointf)(unsafe.Pointer(out)))
}
