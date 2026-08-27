package legacy

import "github.com/opennox/opennox/v1/server"

func rewardReplenishRuntime4F2210(outer Server) server.RewardReplenishRuntime4F2210 {
	return server.RewardReplenishRuntime4F2210{
		QuestStage: func() uint32 {
			return uint32(Nox_game_getQuestStage_4E3CC0())
		},
		PlayerCount: func() int32 {
			return int32(Nox_xxx_player_4E3CE0())
		},
		DelayedDelete: outer.DelayedDelete,
	}
}
