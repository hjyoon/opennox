package legacy

/*
#include <stdint.h>

#include "server__gamemech__explevel.h"

typedef uint16_t wchar2_t;

uint32_t* sub_56FA40(int token, float award);
intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* unit, wchar2_t* message, ...);

static inline void nox_directExperienceProtect_4EF3A0(
		uint32_t token, float award) {
	(void)sub_56FA40((int32_t)token, award);
}

static inline intptr_t nox_directExperienceLine_4EF3A0(
		nox_object_t* unit, wchar2_t* message, uint32_t points) {
	return nox_xxx_netSendLineMessage_4D9EB0(unit, message, points);
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
			C.nox_directExperienceLine_4EF3A0(
				asObjectC(unit),
				text,
				C.uint32_t(points),
			)
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
