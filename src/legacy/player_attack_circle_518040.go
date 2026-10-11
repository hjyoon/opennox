package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerAttackScanCircleNative_518040
func nox_xxx_playerAttackScanCircleNative_518040(center *C.float2, radius C.float, callback, data unsafe.Pointer) {
	if center == nil || callback == nil {
		return
	}
	s := GetServer().S()
	server.PlayerAttackScanCircle518040((*types.Pointf)(unsafe.Pointer(center)), float32(radius),
		s.Map.EachObjInRect, circleBoxCollisionNative54A990,
		func(candidate *server.Object) {
			ccall.CallVoidPtr2(callback, unsafe.Pointer(candidate), data)
		})
}
