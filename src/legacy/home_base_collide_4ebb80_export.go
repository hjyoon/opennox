package legacy

/*
#include "home_base_collide_4ebb80.h"

typedef struct float2 float2;
int nox_xxx_netSendPointFx_522FF0(char code, float2* position);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func homeBaseCollideRuntime4EBB80() server.HomeBaseCollideRuntime4EBB80 {
	return server.HomeBaseCollideRuntime4EBB80{
		ObserverMode: func() uint32 {
			return uint32(Get_dword_5d4594_2650652())
		},
		ObserverUpdate: flagPickupObserverUpdate425CA0,
		MoveTo:         Nox_xxx_unitMove_4E7010,
		PointFX: func(code uint32, pos types.Pointf) uint32 {
			return uint32(C.nox_xxx_netSendPointFx_522FF0(
				C.char(code),
				(*C.float2)(unsafe.Pointer(&pos)),
			))
		},
	}
}

//export nox_xxx_collideHomeBase_4EBB80
func nox_xxx_collideHomeBase_4EBB80(
	homeBase, other *C.nox_object_t,
	collision *C.float,
) C.uint32_t {
	s := GetServer().S()
	return C.uint32_t(s.HomeBaseCollide4EBB80(
		asObjectS((*nox_object_t)(homeBase)),
		asObjectS((*nox_object_t)(other)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		homeBaseCollideRuntime4EBB80(),
	))
}
