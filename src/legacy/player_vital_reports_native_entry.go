package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_netSendTeam_4D8670
func nox_xxx_netSendTeam_4D8670(recipient C.int, unit *nox_object_t) C.int {
	return C.int(GetServer().S().NetSendTeamHealth4D8670(int(recipient), asObjectS(unit)))
}

//export nox_xxx_netSendPlrHealthToTeam_4D86E0
func nox_xxx_netSendPlrHealthToTeam_4D86E0(recipient C.int) C.uintptr_t {
	return C.uintptr_t(GetServer().S().ReportPlayerHealthToTeam4D86E0(int(recipient)))
}

func ReportPlayerVitalsThrottled528030(s *server.Server, unit *server.Object) int16 {
	return s.ReportPlayerVitalsThrottled528030(unit, func(recipient byte) {
		s.ReportPlayerHealthToTeam4D86E0(int(recipient))
	}, func(recipient byte, unit *server.Object) {
		NetReportManaNative4D8930(s, recipient, unit)
	})
}

//export sub_528030
func sub_528030(unit *nox_object_t) C.short {
	return C.short(ReportPlayerVitalsThrottled528030(GetServer().S(), asObjectS(unit)))
}
