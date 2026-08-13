package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_server_questMaybeWarp_4E8F60
func nox_server_questMaybeWarp_4E8F60() C.int {
	result := GetServer().S().QuestMaybeWarp4E8F60(server.QuestMaybeWarpRuntime4E8F60{
		CurrentQuestStage: func() uint32 {
			return uint32(C.nox_game_getQuestStage_4E3CC0())
		},
		NextStageThreshold: func(stage uint32) uint32 {
			return uint32(Nox_server_questNextStageThreshold_4D74F0(int32(stage)))
		},
	})
	return C.int(result)
}
