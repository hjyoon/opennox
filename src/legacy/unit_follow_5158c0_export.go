package legacy

/*
#include "unit_follow_5158c0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var unitFollowCall5158C0 = func(unit, target *server.Object) {
	GetServer().S().UnitSetFollow5158C0(unit, target)
}

func unitFollowExportCall5158C0(unit, target *server.Object) {
	C.nox_xxx_unitSetFollow_5158C0(asObjectC(unit), asObjectC(target))
}

//export nox_xxx_unitSetFollow_5158C0
func nox_xxx_unitSetFollow_5158C0(unit, target *C.nox_object_t) {
	unitFollowCall5158C0(
		asObjectS((*nox_object_t)(unit)),
		asObjectS((*nox_object_t)(target)),
	)
}
