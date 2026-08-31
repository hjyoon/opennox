package legacy

/*
#include <stdint.h>

#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerAbilityCooldownSetExportCall4FBEA0(unit *server.Object, ability, cooldown int32) int32 {
	return int32(C.sub_4FBEA0(asObjectC(unit), C.int32_t(ability), C.int32_t(cooldown)))
}
