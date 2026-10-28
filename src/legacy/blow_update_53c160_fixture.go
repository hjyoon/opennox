package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static void nox_blow_root_probe_53C160(nox_object_t* source) {
 ((void(*)(nox_object_t*))nox_xxx_updateBlow_53C160)(source);
}
static void nox_blow_candidate_probe_53C240(nox_object_t* candidate,nox_object_t* source) {
 ((void(*)(nox_object_t*,nox_object_t*))sub_53C240)(candidate,source);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func blowCProbe53C160(candidate, source *server.Object, direct bool) {
	if direct {
		C.nox_blow_candidate_probe_53C240(asObjectC(candidate), asObjectC(source))
	} else {
		C.nox_blow_root_probe_53C160(asObjectC(source))
	}
}
