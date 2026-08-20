package legacy

/*
#include <stdint.h>

#include "server__object__health.h"

uint32_t* sub_56FA40(int token, float award);
void sub_4EF2E0_exp_level(int unit);

static inline void nox_unitGiveXPProtect_4EF270(
		uint32_t token, float award) {
	(void)sub_56FA40((int32_t)token, award);
}

// 004EF2E0 is the next sequential restoration unit. Keep its one remaining
// ABI32 conversion isolated here until that body receives a typed export.
static inline void nox_unitGiveXPSyncLevel_4EF270(nox_object_t* unit) {
	sub_4EF2E0_exp_level((int32_t)(uintptr_t)unit);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitGiveXPRuntime4EF270() server.UnitGiveXPRuntime4EF270 {
	return server.UnitGiveXPRuntime4EF270{
		ProtectExperience: func(token uint32, award float32) {
			C.nox_unitGiveXPProtect_4EF270(C.uint32_t(token), C.float(award))
		},
		SyncLevel: func(unit *server.Object) {
			C.nox_unitGiveXPSyncLevel_4EF270(asObjectC(unit))
		},
	}
}

func unitGiveXPCall4EF270(unit *server.Object, target float32) float64 {
	return GetServer().S().UnitGiveXP4EF270(unit, target, unitGiveXPRuntime4EF270())
}
