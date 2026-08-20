package legacy

/*
#include <stdint.h>

#include "server__object__health.h"

typedef uint16_t wchar2_t;

intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* player, wchar2_t* format, ...);

static inline void nox_soloMonsterKillRewardSendLine_4EE500(
		nox_object_t* player, wchar2_t* format, uint32_t points) {
	nox_xxx_netSendLineMessage_4D9EB0(player, format, points);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func soloMonsterKillRewardRuntime4EE500() server.SoloMonsterKillRewardRuntime4EE500 {
	return server.SoloMonsterKillRewardRuntime4EE500{
		GiveXP: func(player *server.Object, experience float32) float64 {
			return unitGiveXPCall4EF270(player, experience)
		},
		SendLineMessage: func(player *server.Object, format string, points uint32) {
			message, free := CWString(format)
			defer free()
			C.nox_soloMonsterKillRewardSendLine_4EE500(
				asObjectC(player),
				message,
				C.uint32_t(points),
			)
		},
	}
}

func soloMonsterKillRewardCall4EE500(killed *server.Object) {
	GetServer().S().SoloMonsterKillReward4EE500(
		killed,
		soloMonsterKillRewardRuntime4EE500(),
	)
}

//export nox_xxx_soloMonsterKillReward_4EE500_obj_health
func nox_xxx_soloMonsterKillReward_4EE500_obj_health(killed *C.nox_object_t) {
	soloMonsterKillRewardCall4EE500(asObjectS((*nox_object_t)(killed)))
}
