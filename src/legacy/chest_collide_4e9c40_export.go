package legacy

/*
#include <stdint.h>
#include "chest_collide_4e9c40.h"

extern uint64_t qword_5d4594_1567940;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideChest_4E9C40
func nox_xxx_collideChest_4E9C40(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().ChestCollide4E9C40(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
		server.ChestCollideRuntime4E9C40{
			Ticks: PlatformTicks,
			LoadFeedbackTicks: func() uint64 {
				return uint64(C.qword_5d4594_1567940)
			},
			StoreFeedbackTicks: func(value uint64) {
				C.qword_5d4594_1567940 = C.uint64_t(value)
			},
			DelayedDelete: srv.DelayedDelete,
			ChestOpen:     chestOpenCall4EDF00,
			// This effect remains a separately tracked restored dependency.
			DropAllItems: Nox_xxx_dropAllItems_4EDA40,
		},
	)
}
