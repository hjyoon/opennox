package legacy

/*
#include <stdint.h>
#include "GAME3_2.h"
#include "GAME3_3.h"

extern uint64_t qword_5d4594_1567940;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideDoor_4E8AC0
func nox_xxx_collideDoor_4E8AC0(
	door, unit *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().DoorCollide4E8AC0(
		asObjectS((*nox_object_t)(door)),
		asObjectS((*nox_object_t)(unit)),
		unsafe.Pointer(collision),
		server.DoorCollideRuntime4E8AC0{
			Ticks: PlatformTicks,
			LoadFeedbackTicks: func() uint64 {
				return uint64(C.qword_5d4594_1567940)
			},
			StoreFeedbackTicks: func(value uint64) {
				C.qword_5d4594_1567940 = C.uint64_t(value)
			},
			StoreQuestFrame: func(frame uint32) {
				C.sub_4D71E0(C.int(frame))
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}
