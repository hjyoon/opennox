package legacy

/*
#include "defs.h"
#include "GAME3_2.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

var unitHeightRuntimeNative4D9020 = func() server.UnitHeightRuntime4D9020 {
	core := GetServer().S()
	return server.UnitHeightRuntime4D9020{
		UnitCode: func(unit *server.Object) uint32 { return uint32(core.GetUnitNetCode(unit)) },
		Submit: func(recipient, kind int32, packet []byte) int32 {
			if core.NetList.AddToMsgListCli(ntype.PlayerInd(recipient), netlist.Kind(kind), packet) {
				return 1
			}
			return 0
		},
	}
}

//export nox_xxx_netReportUnitHeight_4D9020
func nox_xxx_netReportUnitHeight_4D9020(recipient C.int, unit *C.nox_object_t) C.int {
	return C.int(server.ObjectHeightReportNative4D9020(int32(recipient), asObjectS(unit), unitHeightRuntimeNative4D9020()))
}
