package server

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
)

// PlayerQuestLivesReport4D99E1 consumes the original report routine's
// entry-cached update data; QuestLivesReport4D9D60 loads its own live pointer.
func (s *Server) PlayerQuestLivesReport4D99E1(unit *Object, update *PlayerUpdateData) {
	questLivesSync4D99E1(unit, update, questLivesSyncHooks4D99E1[*Object, *PlayerUpdateData, *Player]{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		playerByIndex: func(index int32) *Player { return s.Players.ByInd(ntype.PlayerInd(index)) },
		playerUnit:    func(player *Player) *Object { return player.PlayerUnit },
		lives:         func(update *PlayerUpdateData) uint32 { return update.ExtraLives },
		marker:        func(update *PlayerUpdateData, index int32) uint8 { return update.RespawnMarkers[index] },
		report:        s.QuestLivesReport4D9D60,
		lifeByte:      func(update *PlayerUpdateData) uint8 { return uint8(update.ExtraLives) },
		storeMarker:   func(update *PlayerUpdateData, index int32, value uint8) { update.RespawnMarkers[index] = value },
	})
}
