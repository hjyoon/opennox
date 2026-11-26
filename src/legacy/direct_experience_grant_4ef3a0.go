package legacy

/*
#include <stdint.h>

#include "server__gamemech__explevel.h"

int32_t sub_56FA40(int token, float award);

static inline void nox_directExperienceProtect_4EF3A0(
		uint32_t token, float award) {
	(void)sub_56FA40((int32_t)token, award);
}

*/
import "C"

import "github.com/opennox/opennox/v1/server"

func directExperienceGrantRuntime4EF3A0() server.DirectExperienceGrantRuntime4EF3A0 {
	return server.DirectExperienceGrantRuntime4EF3A0{
		ProtectExperience: func(token uint32, award float32) {
			C.nox_directExperienceProtect_4EF3A0(
				C.uint32_t(token),
				C.float(award),
			)
		},
		SendLineMessage: func(unit *server.Object, message string, points uint32) {
			text, free := CWString(message)
			defer free()
			sendNumberedLineMessage(unit, text, points)
		},
		SyncLevel: experienceLevelUpdateCall4EF2E0,
	}
}

func directExperienceGrantCall4EF3A0(unit *server.Object, award float32) {
	GetServer().S().DirectExperienceGrant4EF3A0(
		unit,
		award,
		directExperienceGrantRuntime4EF3A0(),
	)
}

// Nox_xxx_plyrGiveExp_4EF3A0_exp_level grants experience without routing a
// native-width Object pointer through the legacy PE32 script stack.
func Nox_xxx_plyrGiveExp_4EF3A0_exp_level(unit *server.Object, award float32) {
	directExperienceGrantCall4EF3A0(unit, award)
}
