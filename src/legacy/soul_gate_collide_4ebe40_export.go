package legacy

/*
#include "soul_gate_collide_4ebe40.h"

typedef struct float2 float2;
int nox_xxx_netSendPointFx_522FF0(char code, float2* position);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func soulGateCollideRuntime4EBE40() server.SoulGateCollideRuntime4EBE40 {
	return server.SoulGateCollideRuntime4EBE40{
		SetQuestMode: func(value int32) {
			Sub_4D7520(int(value))
		},
		SetQuestTimer: func(frame uint32) {
			Sub_4D71E0(int(frame))
		},
		PointFX: func(code uint32, position *types.Pointf) uint32 {
			return uint32(C.nox_xxx_netSendPointFx_522FF0(
				C.char(code),
				(*C.float2)(unsafe.Pointer(position)),
			))
		},
	}
}

//export sub_4EBE40
func sub_4EBE40(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	GetServer().S().SoulGateCollide4EBE40(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		soulGateCollideRuntime4EBE40(),
	)
}
