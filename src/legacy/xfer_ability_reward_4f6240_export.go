package legacy

/*
#include "xfer_ability_reward_4f6240.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var abilityRewardXferCall4F6240 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerAbilityRewardNative4F6240(cf, object)
}

func abilityRewardXferExportCall4F6240(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerAbilityReward_4F6240(asObjectC(object)))
}

//export nox_xxx_XFerAbilityReward_4F6240
func nox_xxx_XFerAbilityReward_4F6240(object *C.nox_object_t) C.int32_t {
	return C.int32_t(abilityRewardXferCall4F6240(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
