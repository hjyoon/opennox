package legacy

/*
#include "defs.h"
#include "GAME4_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_checkIsKillable_528190
func nox_xxx_checkIsKillable_528190(unit *C.nox_object_t) C.int {
	return C.int(server.ObjectKillableNative528190(asObjectS(unit)))
}
