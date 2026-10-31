package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
_Static_assert(_Generic(&nox_xxx_updateNPCAnimData_50A850, char (*)(nox_object_t*):1,default:0), "animation update char/native-object ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func animationUpdatePublicC50A850(unit *server.Object) uint8 {
	return uint8(C.nox_xxx_updateNPCAnimData_50A850(asObjectC(unit)))
}
