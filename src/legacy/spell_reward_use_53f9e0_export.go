package legacy

/*
#include <stdint.h>

#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func useSpellRewardLegacy53F9E0(owner, item *server.Object) int32 {
	return GetServer().UseSpellReward53F9E0(owner, item)
}

func spellRewardUseExportCall53F9E0(owner, item *server.Object) int32 {
	return int32(C.nox_xxx_useSpellReward_53F9E0(
		asObjectC(owner),
		asObjectC(item),
	))
}

//export nox_xxx_useSpellReward_53F9E0
func nox_xxx_useSpellReward_53F9E0(
	owner, item *C.nox_object_t,
) C.int32_t {
	return C.int32_t(useSpellRewardLegacy53F9E0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	))
}
