package legacy

/*
#include "unit_buff_update_4ff620.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitBuffUpdateRuntime4FF620(srv Server) server.UnitBuffUpdateRuntime4FF620 {
	return server.UnitBuffUpdateRuntime4FF620{
		DamageClear: unitDamageClearCall4EE5E0,
		IncrementElimDeath: func(unit *server.Object) {
			srv.PlayerIncrementElimDeath4D8D40(unit)
		},
		BuffOff: func(unit *server.Object, buff server.EnchantID) {
			_ = spellBuffOffLegacy4FF5B0(unit, int32(buff))
		},
	}
}

func unitBuffUpdateLegacy4FF620(unit *server.Object) {
	srv := GetServer()
	srv.S().UnitBuffUpdate4FF620(unit, unitBuffUpdateRuntime4FF620(srv))
}

var unitBuffUpdateExportImpl4FF620 = unitBuffUpdateLegacy4FF620

func unitBuffUpdateExportCall4FF620(unit *server.Object) {
	C.nox_xxx_updateUnitBuffs_4FF620(asObjectC(unit))
}

//export nox_xxx_updateUnitBuffs_4FF620
func nox_xxx_updateUnitBuffs_4FF620(unit *C.nox_object_t) {
	unitBuffUpdateExportImpl4FF620(asObjectS((*nox_object_t)(unit)))
}
