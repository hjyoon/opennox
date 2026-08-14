package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

type questAllPlayersExitedNativeDeps4E9010 struct {
	firstUnit   func() *Object
	nextUnit    func(*Object) *Object
	gameHost    func() int32
	noRendering func() int32
}

func questAllPlayersExitedNative4E9010(deps questAllPlayersExitedNativeDeps4E9010) int32 {
	return questAllPlayersExited4E9010(questAllPlayersExitedHooks4E9010[*Object, *PlayerUpdateData, *Player]{
		firstUnit: deps.firstUnit,
		nextUnit:  deps.nextUnit,
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		gameHost:    deps.gameHost,
		noRendering: deps.noRendering,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		loadQuestState: func(player *Player) uint32 {
			return player.Field4792
		},
		loadQuestExit: func(update *PlayerUpdateData) *Object {
			return update.QuestExit
		},
	})
}

// QuestAllPlayersExited4E9010 binds the restored Quest-exit readiness
// contract to native Object, PlayerUpdateData and Player pointers.
func (s *Server) QuestAllPlayersExited4E9010() int32 {
	return questAllPlayersExitedNative4E9010(questAllPlayersExitedNativeDeps4E9010{
		firstUnit: s.Players.FirstUnit,
		nextUnit:  s.questNextPlayerUnit4DA7F0,
		gameHost: func() int32 {
			if noxflags.HasGame(noxflags.GameHost) {
				return 1
			}
			return 0
		},
		noRendering: func() int32 {
			if noxflags.HasEngine(noxflags.EngineNoRendering) {
				return 1
			}
			return 0
		},
	})
}
