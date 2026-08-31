package legacy

/*
#include <stdint.h>

#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerAbilityCooldownGetExportCall4FBE60(unit *server.Object, ability int32) int32 {
	return int32(C.sub_4FBE60(asObjectC(unit), C.int32_t(ability)))
}
