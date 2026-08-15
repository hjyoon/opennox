package legacy

/*
#include <stdint.h>

#include "ankh_collide_4ebf40.h"
#include "memmap.h"

typedef struct float2 float2;
int nox_xxx_netSendPointFx_522FF0(char code, float2* position);

extern uint64_t qword_5d4594_1567940;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func ankhCollideRuntime4EBF40() server.AnkhCollideRuntime4EBF40 {
	return server.AnkhCollideRuntime4EBF40{
		Ticks: PlatformTicks,
		LoadFeedbackTicks: func() uint64 {
			return uint64(C.qword_5d4594_1567940)
		},
		StoreFeedbackTicks: func(value uint64) {
			C.qword_5d4594_1567940 = C.uint64_t(value)
		},
		LoadResetName: func() string {
			ptr := C.mem_getPtr(C.uintptr_t(0x5D4594), C.uintptr_t(1568012))
			return alloc.GoString16((*uint16)(ptr))
		},
		LoadResetSerialFirst: func() uint8 {
			ptr := C.mem_getPtr(C.uintptr_t(0x5D4594), C.uintptr_t(1568016))
			return *(*uint8)(ptr)
		},
		PointFX: func(code uint32, position *types.Pointf) uint32 {
			return uint32(C.nox_xxx_netSendPointFx_522FF0(
				C.char(code),
				(*C.float2)(unsafe.Pointer(position)),
			))
		},
	}
}

//export nox_xxx_collideAnkhQuest_4EBF40
func nox_xxx_collideAnkhQuest_4EBF40(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	GetServer().S().AnkhCollide4EBF40(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		ankhCollideRuntime4EBF40(),
	)
}
