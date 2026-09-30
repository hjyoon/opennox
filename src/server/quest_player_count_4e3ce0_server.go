package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// QuestPlayerCount4E3CE0 keeps player-unit, update and player pointers at their
// native width while preserving the original read-only counting contract.
func (s *Server) QuestPlayerCount4E3CE0() int32 {
	return questPlayerCount4E3CE0(questPlayerCountHooks4E3CE0[*Object, *PlayerUpdateData, *Player]{
		firstUnit:  s.Players.FirstUnit,
		nextUnit:   s.questNextPlayerUnit4DA7F0,
		loadUpdate: func(unit *Object) *PlayerUpdateData { return (*PlayerUpdateData)(unit.UpdateData) },
		gameHost:   func() bool { return noxflags.HasGame(noxflags.GameHost) },
		noRendering: func() bool {
			return noxflags.HasEngine(noxflags.EngineNoRendering)
		},
		loadPlayer:      func(update *PlayerUpdateData) *Player { return update.Player },
		loadPlayerIndex: func(player *Player) uint8 { return player.PlayerInd },
		loadQuestState:  func(player *Player) uint32 { return player.Field4792 },
	})
}

func (s *Server) questPlayerCount4E3CE0() int {
	return int(s.QuestPlayerCount4E3CE0())
}
