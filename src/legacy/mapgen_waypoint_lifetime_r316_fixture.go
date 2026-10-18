package legacy

/*
#include "GAME5_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func mapgenWaypointRaw316() *server.Waypoint {
	return asWaypointS(C.sub_579E70())
}
