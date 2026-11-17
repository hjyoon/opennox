package legacy

/*
#include "common__system__team.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Return native integer values across private Cgo frames. The public C entries
// keep their pointer types without introducing Go pointer stores into C memory.

//export nox_server_team_first_member_native
func nox_server_team_first_member_native(team *C.nox_team_t) C.uintptr_t {
	if team == nil {
		return 0
	}
	value := GetServer().S().Teams.FirstObjectTeam(asTeam(team))
	return C.uintptr_t(uintptr(unsafe.Pointer(value)))
}

//export nox_server_team_next_member_native
func nox_server_team_next_member_native(value *C.nox_object_team_t) C.uintptr_t {
	if value == nil {
		return 0
	}
	next := GetServer().S().Teams.NextObjectTeam((*server.ObjectTeam)(unsafe.Pointer(value)))
	return C.uintptr_t(uintptr(unsafe.Pointer(next)))
}
