package legacy

/*
#include "unit_banish_5017f0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var unitBanishCall5017F0 = func(unit *server.Object) {
	srv := GetServer()
	srv.S().BanishUnit5017F0(unit, server.UnitBanishRuntime5017F0{
		DelayedDelete: srv.DelayedDelete,
		SendPointFX:   srv.S().Nox_xxx_netSendPointFx_522FF0,
	})
}

func unitBanishExportCall5017F0(unit *server.Object) {
	C.nox_xxx_banishUnit_5017F0(asObjectC(unit))
}

//export nox_xxx_banishUnit_5017F0
func nox_xxx_banishUnit_5017F0(unit *C.nox_object_t) {
	unitBanishCall5017F0(asObjectS((*nox_object_t)(unit)))
}

func Nox_xxx_banishUnit_5017F0(unit *server.Object) {
	unitBanishCall5017F0(unit)
}
