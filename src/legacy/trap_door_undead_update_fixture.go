package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static void nox_trapdoor_undead_update_probe(int kind, nox_object_t* source) {
 if (kind == 0) ((void(*)(nox_object_t*))nox_xxx_updateTrapDoor_53DE80)(source);
 else ((void(*)(nox_object_t*))nox_xxx_updateUndeadKiller_53E190)(source);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func trapDoorUndeadNativeProbe(kind int, source *server.Object) {
	C.nox_trapdoor_undead_update_probe(C.int(kind), asObjectC(source))
}
