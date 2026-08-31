package legacy

/*
#include <stdint.h>

#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_probablyWarcryCheck_4FC3E0
func nox_xxx_probablyWarcryCheck_4FC3E0(unit *nox_object_t, ability int32) int32 {
	return GetServer().S().Abils.ActiveValue4FC3E0(asObjectS(unit), server.Ability(ability))
}

func activeAbilityValueExportCall4FC3E0(unit *server.Object, ability int32) int32 {
	return int32(C.nox_xxx_probablyWarcryCheck_4FC3E0(asObjectC(unit), C.int32_t(ability)))
}
