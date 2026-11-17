package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func teamPredicateNativeValue(value *C.nox_object_team_t) *server.ObjectTeam {
	return (*server.ObjectTeam)(unsafe.Pointer(value))
}

//export nox_xxx_servObjectHasTeam_419130
func nox_xxx_servObjectHasTeam_419130(value *C.nox_object_team_t) C.int {
	if teamPredicateNativeValue(value).Has() {
		return 1
	}
	return 0
}

//export nox_xxx_servCompareTeams_419150
func nox_xxx_servCompareTeams_419150(first, second *C.nox_object_team_t) C.int {
	if teamPredicateNativeValue(first).SameAs(teamPredicateNativeValue(second)) {
		return 1
	}
	return 0
}
