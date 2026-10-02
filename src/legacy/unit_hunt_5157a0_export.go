package legacy

/*
#include "unit_hunt_5157a0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var unitHuntCall5157A0 = server.UnitHunt5157A0

//export nox_server_unit_hunt_5157a0
func nox_server_unit_hunt_5157a0(unit *C.nox_object_t) {
	unitHuntCall5157A0(asObjectS((*nox_object_t)(unit)))
}
