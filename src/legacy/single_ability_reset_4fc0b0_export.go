package legacy

/*
#include <stdint.h>
#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func singleAbilityResetExportCall4FC0B0(unit *server.Object, ability int32) {
	C.sub_4FC0B0(asObjectC(unit), C.int32_t(ability))
}
