package legacy

/*
#include "unit_buff_clear_4ff580.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitBuffClearRuntime4FF580() server.UnitBuffClearRuntime4FF580 {
	return server.UnitBuffClearRuntime4FF580{
		ResetPlayerProtection: func(player *server.Player, flags uint32) {
			Nox_xxx_playerResetProtectionCRC_56F7D0(player.ProtUnitBuffs, int(flags))
		},
	}
}

func unitBuffClearLegacy4FF580(unit *server.Object) {
	unit.UnitBuffClear4FF580(unitBuffClearRuntime4FF580())
}

var unitBuffClearExportImpl4FF580 = unitBuffClearLegacy4FF580

func unitBuffClearExportCall4FF580(unit *server.Object) {
	C.nox_xxx_unitClearBuffs_4FF580(asObjectC(unit))
}

//export nox_xxx_unitClearBuffs_4FF580
func nox_xxx_unitClearBuffs_4FF580(unit *C.nox_object_t) {
	unitBuffClearExportImpl4FF580(asObjectS((*nox_object_t)(unit)))
}
