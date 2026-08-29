package legacy

/*
#include "xfer_spell_reward_4f5f30.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var spellRewardXferCall4F5F30 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerSpellRewardNative4F5F30(cf, object)
}

func spellRewardXferExportCall4F5F30(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerSpellReward_4F5F30(asObjectC(object)))
}

//export nox_xxx_XFerSpellReward_4F5F30
func nox_xxx_XFerSpellReward_4F5F30(object *C.nox_object_t) C.int32_t {
	return C.int32_t(spellRewardXferCall4F5F30(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
