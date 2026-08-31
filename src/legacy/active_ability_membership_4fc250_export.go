package legacy

/*
#include <stdint.h>

#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func activeAbilityMembershipExportCall4FC250(unit *server.Object, ability int32) int32 {
	return int32(C.nox_common_playerIsAbilityActive_4FC250(asObjectC(unit), C.int32_t(ability)))
}
