package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

const questPlayerClassByte4DA7F0 = uint8(0x04)

type questMaybeWarpNativeDeps4E8F60 struct {
	currentQuestStage  func() uint32
	nextStageThreshold func(uint32) uint32
	firstUnit          func() *Object
	nextUnit           func(*Object) *Object
	gameHost           func() int32
	noRendering        func() int32
}

// QuestMaybeWarpRuntime4E8F60 supplies the Quest-stage operations whose
// storage remains in the legacy runtime.
type QuestMaybeWarpRuntime4E8F60 struct {
	CurrentQuestStage  func() uint32
	NextStageThreshold func(uint32) uint32
}

func questMaybeWarpNative4E8F60(deps questMaybeWarpNativeDeps4E8F60) int32 {
	return questMaybeWarp4E8F60(questMaybeWarpHooks4E8F60[*Object, *PlayerUpdateData, *Player]{
		currentQuestStage:  deps.currentQuestStage,
		nextStageThreshold: deps.nextStageThreshold,
		firstUnit:          deps.firstUnit,
		nextUnit:           deps.nextUnit,
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
		loadQuestWarpGate: func(update *PlayerUpdateData) *Object {
			return update.QuestWarpGate
		},
		loadQuestStage: func(player *Player) uint32 {
			return player.QuestStage
		},
	})
}

// questNextPlayerUnit4DA7F0 mirrors GAME.EXE 004DA7F0 without the extra class
// and allocation-liveness reads performed by the general UpdateDataPlayer
// helper.
func (s *Server) questNextPlayerUnit4DA7F0(unit *Object) *Object {
	if unit == nil || uint8(unit.ObjClass)&questPlayerClassByte4DA7F0 == 0 {
		return nil
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	for player := s.Players.Next(update.Player); player != nil; player = s.Players.Next(player) {
		if player.PlayerUnit != nil {
			return player.PlayerUnit
		}
	}
	return nil
}

// QuestMaybeWarp4E8F60 binds the restored warp eligibility contract to native
// Object, PlayerUpdateData and Player pointers.
func (s *Server) QuestMaybeWarp4E8F60(runtime QuestMaybeWarpRuntime4E8F60) int32 {
	return questMaybeWarpNative4E8F60(questMaybeWarpNativeDeps4E8F60{
		currentQuestStage:  runtime.CurrentQuestStage,
		nextStageThreshold: runtime.NextStageThreshold,
		firstUnit:          s.Players.FirstUnit,
		nextUnit:           s.questNextPlayerUnit4DA7F0,
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
