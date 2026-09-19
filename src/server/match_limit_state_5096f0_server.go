package server

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// MatchLimitStateRuntime5096F0 supplies state still owned by the outer game
// and the retained timer subsystem. Winner callbacks stay native-width.
type MatchLimitStateRuntime5096F0 struct {
	LimitExpired     func() int32
	SwitchToNextMap  func()
	PrintAutoExit    func()
	MatchStateActive func() int32
	ResolveTeamMode  func() int32
	ResolveHighScore func() int32
	ResolveLowScore  func() int32
	ClearTimer       func() int32
}

type matchLimitStateNativeDeps5096F0 struct {
	firstPlayer  func() *Object
	nextPlayer   func(*Object) *Object
	sendPosition func(uint8, [5]byte)
	audioEvent   func(uint32, *Object, int32, uint32)
}

func (s *Server) matchLimitStateNative5096F0(
	runtime MatchLimitStateRuntime5096F0,
	deps matchLimitStateNativeDeps5096F0,
) int32 {
	return matchLimitState5096F0(matchLimitStateHooks5096F0[*Object, *PlayerUpdateData, *Player]{
		limitExpired: runtime.LimitExpired,
		hasGameFlags: func(flags uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flags))
		},
		switchToNextMap: runtime.SwitchToNextMap,
		printAutoExit:   runtime.PrintAutoExit,
		firstPlayer:     deps.firstPlayer,
		nextPlayer:      deps.nextPlayer,
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadQuestExit: func(update *PlayerUpdateData) *Object {
			return update.QuestExit
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadQuestState: func(player *Player) uint32 {
			return player.Field4792
		},
		recordQuestProgress: s.exitCollideRecordProgress4D60E0,
		matchStateActive:    runtime.MatchStateActive,
		resolveTeamMode:     runtime.ResolveTeamMode,
		resolveHighScore:    runtime.ResolveHighScore,
		resolveLowScore:     runtime.ResolveLowScore,
		setGameFlags: func(flags uint32) {
			noxflags.SetGame(noxflags.GameFlag(flags))
		},
		loadPositionX: func(player *Player) float32 {
			return player.Pos3632Vec.X
		},
		loadPositionY: func(player *Player) float32 {
			return player.Pos3632Vec.Y
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		sendPosition: deps.sendPosition,
		loadNetCode: func(unit *Object) uint32 {
			return unit.NetCode
		},
		audioEvent: deps.audioEvent,
		clearTimer: runtime.ClearTimer,
	})
}

// MatchLimitState5096F0 binds GAME.EXE 005096F0 to native Object,
// PlayerUpdateData, and Player pointers.
func (s *Server) MatchLimitState5096F0(runtime MatchLimitStateRuntime5096F0) int32 {
	return s.matchLimitStateNative5096F0(runtime, matchLimitStateNativeDeps5096F0{
		firstPlayer: s.Players.FirstUnit,
		nextPlayer:  s.questNextPlayerUnit4DA7F0,
		sendPosition: func(index uint8, packet [5]byte) {
			s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind1, packet[:])
		},
		audioEvent: func(id uint32, unit *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), code)
		},
	})
}
