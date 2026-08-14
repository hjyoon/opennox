package legacy

/*
#include "units_same_team_4ec520.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_unitsHaveSameTeam_4EC520
func nox_xxx_unitsHaveSameTeam_4EC520(first, second *C.nox_object_t) C.int {
	if server.UnitsHaveSameTeam4EC520(
		asObjectS((*nox_object_t)(first)),
		asObjectS((*nox_object_t)(second)),
	) {
		return 1
	}
	return 0
}
