package legacy

/*
#include <stdint.h>

#include "server__gamemech__explevel.h"

uint32_t* sub_56F980(int token, unsigned char level);
void sub_57AF30(int unit, int mode);

static inline void nox_experienceLevelProtect_4EF2E0(
		uint32_t token, uint8_t level) {
	(void)sub_56F980((int32_t)token, (unsigned char)level);
}

// PauseFX 0057AF30 remains an ABI32 restoration dependency. Keep the sole
// pointer narrowing for this branch isolated here until PauseFX is ported.
static inline void nox_experienceLevelPauseFX_4EF2E0(
		nox_object_t* unit, int32_t mode) {
	sub_57AF30((int32_t)(uintptr_t)unit, mode);
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
			C.nox_experienceLevelPauseFX_4EF2E0(
				asObjectC(unit),
				C.int32_t(mode),
			)
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
