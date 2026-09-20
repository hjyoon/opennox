package legacy

/*
#include <stdint.h>

#include "server__gamemech__explevel.h"

uint32_t* sub_56F980(int token, unsigned char level);

static inline void nox_experienceLevelProtect_4EF2E0(
		uint32_t token, uint8_t level) {
	(void)sub_56F980((int32_t)token, (unsigned char)level);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func experienceLevelUpdateRuntime4EF2E0() server.ExperienceLevelUpdateRuntime4EF2E0 {
	return server.ExperienceLevelUpdateRuntime4EF2E0{
		GameGet: func() int32 {
			if Nox_xxx_gameGet_4DB1B0() {
				return 1
			}
			return 0
		},
		GameSubActive: func() bool {
			return Sub_4DB1C0() != nil
		},
		ProtectLevel: func(token uint32, level uint8) {
			C.nox_experienceLevelProtect_4EF2E0(
				C.uint32_t(token),
				C.uint8_t(level),
			)
		},
		ReadValues: playerReadValuesCall4EEDC0,
		PauseFX: func(unit *server.Object, mode int32) {
			pauseFXStartCall57AF30(unit, mode)
		},
		SendLineMessage: Nox_xxx_netSendLineMessage_4D9EB0,
	}
}

func experienceLevelUpdateCall4EF2E0(unit *server.Object) {
	GetServer().S().ExperienceLevelUpdate4EF2E0(
		unit,
		experienceLevelUpdateRuntime4EF2E0(),
	)
}
