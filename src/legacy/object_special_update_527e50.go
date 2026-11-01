package legacy

/*
#include "defs.h"
#include "GAME4_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var objectSpecialRuntimeNative527E50 = func() server.ObjectSpecialRuntime527E50 {
	return GetServer().ObjectSpecialRuntimeNative527E50()
}

func NetUpdateObjectSpecial527E50(recipient, unit *server.Object) int32 {
	return server.ObjectSpecialUpdateNative527E50(recipient, unit, objectSpecialRuntimeNative527E50())
}

//export nox_xxx_netUpdateObjectSpecial_527E50
func nox_xxx_netUpdateObjectSpecial_527E50(recipient, unit *C.nox_object_t) C.int {
	return C.int(NetUpdateObjectSpecial527E50(asObjectS(recipient), asObjectS(unit)))
}
