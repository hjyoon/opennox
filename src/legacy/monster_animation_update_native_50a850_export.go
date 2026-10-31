package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
_Static_assert(_Generic(&nox_xxx_updateNPCAnimData_50A850, char (*)(nox_object_t*):1,default:0), "animation update char/native-object ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_updateNPCAnimData_50A850
func nox_xxx_updateNPCAnimData_50A850(unit *C.nox_object_t) C.char {
	return C.char(server.MonsterAnimationUpdateNative50A850(asObjectS((*nox_object_t)(unit)), monsterAnimationUpdateRuntimeNative50A850()))
}
