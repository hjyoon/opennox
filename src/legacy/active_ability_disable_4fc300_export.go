package legacy

/*
#include <stdint.h>
#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func activeAbilityDisableExportCall4FC300(unit *server.Object, ability int32) {
	C.sub_4FC300(asObjectC(unit), C.int32_t(ability))
}
