package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static inline void nox_spell_flight_probe_53B940(nox_object_t* missile) {
 nox_xxx_spellFlyUpdate_53B940(missile);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func spellFlightCProbe53B940(missile *server.Object) {
	C.nox_spell_flight_probe_53B940(asObjectC(missile))
}
