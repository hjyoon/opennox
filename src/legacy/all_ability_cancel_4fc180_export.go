package legacy

/*
#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func allAbilityCancelExportCall4FC180(unit *server.Object) {
	C.nox_xxx_playerCancelAbils_4FC180(asObjectC(unit))
}
