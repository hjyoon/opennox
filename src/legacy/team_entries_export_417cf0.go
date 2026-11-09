package legacy

/*
#include "GAME1.h"
#include "common__system__team.h"
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

//export sub_417CF0
func sub_417CF0() C.int { return C.int(sub417CF0Native()) }

//export sub_417DC0
func sub_417DC0() C.int { return C.int(sub417DC0Native()) }

//export sub_417DE0
func sub_417DE0() C.char { return C.char(sub417DE0Native()) }

//export sub_418390
func sub_418390() C.int { return C.int(sub418390Native()) }

//export sub_4183C0
func sub_4183C0() C.int { return C.int(sub4183C0Native()) }

//export sub_4184D0
func sub_4184D0(team *C.nox_team_t) { sub4184D0Native((*server.Team)(unsafe.Pointer(team))) }
