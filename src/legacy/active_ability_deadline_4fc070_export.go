package legacy

/*
#include <stdint.h>

#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func activeAbilityDeadlineExportCall4FC070(unit *server.Object, ability, delta int32) {
	C.sub_4FC070(asObjectC(unit), C.int32_t(ability), C.int32_t(delta))
}
