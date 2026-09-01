package legacy

/*
#include <stdint.h>

#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func warpReadUseRuntime53F830() server.WarpReadUseRuntime53F830 {
	return server.WarpReadUseRuntime53F830{
		WarpEnabled: func() int32 {
			return int32(C.sub_4D75E0())
		},
		CurrentQuestStage: func() int32 {
			return int32(C.nox_game_getQuestStage_4E3CC0())
		},
		NextStageThreshold: func(stage int32) int32 {
			return Nox_server_questNextStageThreshold_4D74F0(stage)
		},
	}
}

func useWarpReadLegacy53F830(owner, readable *server.Object) int32 {
	return GetServer().S().WarpReadUse53F830(
		owner,
		readable,
		warpReadUseRuntime53F830(),
	)
}

func warpReadUseExportCall53F830(owner, readable *server.Object) int32 {
	return int32(C.sub_53F830(
		asObjectC(owner),
		asObjectC(readable),
	))
}

//export sub_53F830
func sub_53F830(
	owner, readable *C.nox_object_t,
) C.int32_t {
	return C.int32_t(useWarpReadLegacy53F830(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(readable)),
	))
}
