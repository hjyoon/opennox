package legacy

/*
#include "defs.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func blowUpdateNative53C160(source *server.Object) {
	GetServer().S().BlowUpdate53C160(source)
}

// Keep this indirection dynamic so tests can prove direct native-width Go
// dispatch while retaining the legacy callback identity stored in thing.bin.
var blowUpdateCall53C160 = blowUpdateNative53C160

//export nox_blow_update_native_53C160
func nox_blow_update_native_53C160(source *nox_object_t) {
	server.BlowUpdateRaw53C160(asObjectS(source), blowUpdateRuntime53C160())
}

//export nox_blow_candidate_native_53C240
func nox_blow_candidate_native_53C240(candidate, source *nox_object_t) {
	server.BlowCandidateRaw53C240(asObjectS(candidate), asObjectS(source), blowUpdateRuntime53C160())
}
