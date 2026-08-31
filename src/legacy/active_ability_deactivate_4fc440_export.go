package legacy

/*
#include <stdint.h>

#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export sub_4FC440
func sub_4FC440(unit *nox_object_t, ability int32) {
	GetServer().S().Abils.Sub4FC440(asObjectS(unit), server.Ability(ability))
}

func activeAbilityDeactivateExportCall4FC440(unit *server.Object, ability int32) {
	C.sub_4FC440(asObjectC(unit), C.int32_t(ability))
}
