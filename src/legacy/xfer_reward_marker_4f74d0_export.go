package legacy

/*
#include "xfer_reward_marker_4f74d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var rewardMarkerXferCall4F74D0 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerRewardMarkerNative4F74D0(cf, object)
}

func rewardMarkerXferExportCall4F74D0(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerRewardMarker_4F74D0(asObjectC(object)))
}

//export nox_xxx_XFerRewardMarker_4F74D0
func nox_xxx_XFerRewardMarker_4F74D0(object *C.nox_object_t) C.int32_t {
	return C.int32_t(rewardMarkerXferCall4F74D0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
