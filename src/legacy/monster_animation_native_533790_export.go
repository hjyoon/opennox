package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(_Generic(&nox_xxx_mobActionToAnimation_533790, int (*)(nox_object_t*):1,default:0), "native animation selector ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_mobActionToAnimation_533790
func nox_xxx_mobActionToAnimation_533790(unit *C.nox_object_t) C.int {
	return C.int(server.MonsterAnimationNative533790(asObjectS((*nox_object_t)(unit)), monsterAnimationRuntimeNative533790()))
}
