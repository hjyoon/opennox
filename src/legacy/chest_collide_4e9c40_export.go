package legacy

/*
#include <stdint.h>
#include "chest_collide_4e9c40.h"

extern uint64_t qword_5d4594_1567940;
void nox_xxx_chest_4EDF00(int source, int target);
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
			// These effects remain ABI32 internally and are tracked as the
			// separate 004EDF00/004EDA40 restoration scope.
			ChestOpen: func(chest, unit *server.Object) {
				C.nox_xxx_chest_4EDF00(
					C.int(uintptr(chest.CObj())),
					C.int(uintptr(unit.CObj())),
				)
			},
			DropAllItems: Nox_xxx_dropAllItems_4EDA40,
		},
	)
}
