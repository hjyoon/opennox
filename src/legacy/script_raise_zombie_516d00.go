package legacy

/*
#include "script_raise_zombie_516d00.h"
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export sub_516D00
func sub_516D00(unit *nox_object_t) C.uintptr_t {
	return C.uintptr_t(server.ScriptRaiseZombieRaw516D00(asObjectS(unit), func(obj *server.Object) uint32 {
		return uint32(C.nox_xxx_mobRaiseZombie_534AB0(asObjectC(obj)))
	}))
}
