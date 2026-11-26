package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Team iteration borrows the attached object's native membership record.

//export nox_xxx_teamCheckSmth_418C60
func nox_xxx_teamCheckSmth_418C60(team *C.nox_team_t) *C.nox_object_team_t {
	if team == nil {
		return nil
	}
	value := GetServer().S().Teams.FirstObjectTeam(asTeam(team))
	return (*C.nox_object_team_t)(unsafe.Pointer(value))
}

//export sub_418C70
func sub_418C70(value *C.nox_object_team_t) *C.nox_object_team_t {
	if value == nil {
		return nil
	}
	next := GetServer().S().Teams.NextObjectTeam((*server.ObjectTeam)(unsafe.Pointer(value)))
	return (*C.nox_object_team_t)(unsafe.Pointer(next))
}
