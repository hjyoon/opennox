package legacy

/*
#include <stdint.h>

uint32_t* sub_56F820(int token, unsigned char level);

static inline void nox_playerSyncLevel_protect_4EF140(
		uint32_t token, uint8_t level) {
	(void)sub_56F820((int32_t)token, (unsigned char)level);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerSyncLevelRuntime4EF140() server.PlayerSyncLevelRuntime4EF140 {
	return server.PlayerSyncLevelRuntime4EF140{
		ProtectLevel: func(token uint32, level uint8) {
			C.nox_playerSyncLevel_protect_4EF140(
				C.uint32_t(token),
				C.uint8_t(level),
			)
		},
		ReadValues: playerReadValuesCall4EEDC0,
	}
}

func playerSyncLevelCall4EF140(unit *server.Object) int32 {
	return GetServer().S().PlayerSyncLevel4EF140(
		unit,
		playerSyncLevelRuntime4EF140(),
	)
}
