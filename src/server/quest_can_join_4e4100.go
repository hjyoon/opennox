package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

type questCanJoinNativeDeps4E4100 struct {
	firstUnit   func() *Object
	nextUnit    func(*Object) *Object
	gameHost    func() bool
	noRendering func() bool
}

// questCanJoinNative4E4100 preserves GAME.EXE 004E4100's full traversal,
// short-circuit flag reads and nonzero-state qualification. The original
// returns the final nil successor (zero) when the count is at least six.
func questCanJoinNative4E4100(deps questCanJoinNativeDeps4E4100) uint32 {
	var count uint32
	for unit := deps.firstUnit(); unit != nil; unit = deps.nextUnit(unit) {
		update := (*PlayerUpdateData)(unit.UpdateData)
		if deps.gameHost() && deps.noRendering() && update.Player.PlayerInd == HostPlayerIndex {
			continue
		}
		if update.Player.Field4792 != 0 {
			count++
		}
	}
	if count < 6 {
		return 1
	}
	return 0
}

// QuestCanJoin4E4100 tests the Quest six-participant limit using native
// Object, PlayerUpdateData and Player pointers instead of PE32 offsets.
func (s *Server) QuestCanJoin4E4100() uint32 {
	return questCanJoinNative4E4100(questCanJoinNativeDeps4E4100{
		firstUnit: s.Players.FirstUnit,
		nextUnit:  s.questNextPlayerUnit4DA7F0,
		gameHost: func() bool {
			return noxflags.HasGame(noxflags.GameHost)
		},
		noRendering: func() bool {
			return noxflags.HasEngine(noxflags.EngineNoRendering)
		},
	})
}
