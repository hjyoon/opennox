package legacy

/*
#include "summon_mana_cost_500ca0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var summonManaCostCall500CA0 = server.SummonManaCost500CA0

func summonManaCostExportCall500CA0(spellID int32, unit *server.Object) int32 {
	return int32(C.sub_500CA0(C.int32_t(spellID), asObjectC(unit)))
}

//export sub_500CA0
func sub_500CA0(spellID C.int32_t, unit *C.nox_object_t) C.int32_t {
	return C.int32_t(summonManaCostCall500CA0(
		int32(spellID),
		asObjectS((*nox_object_t)(unit)),
	))
}
