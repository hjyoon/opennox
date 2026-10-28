package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static void nox_meteor_native_entry_probe(int kind, nox_object_t* source) {
 if (kind == 0) ((void (*)(nox_object_t*))nox_xxx_updateMeteorShower_53D5A0)(source);
 else ((void (*)(nox_object_t*))nox_xxx_meteorExplode_53D6E0)(source);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func meteorNativeEntryProbe53D5A0(kind int, source *server.Object) {
	C.nox_meteor_native_entry_probe(C.int(kind), asObjectC(source))
}
