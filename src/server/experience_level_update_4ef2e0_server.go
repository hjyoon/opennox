package server

import (
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// ExperienceLevelUpdateRuntime4EF2E0 supplies services still owned by the
// legacy runtime. Every Object-bearing callback keeps the host pointer width;
// only protection tokens and original integer arguments remain fixed-width.
type ExperienceLevelUpdateRuntime4EF2E0 struct {
	GameGet         func() int32
	GameSubActive   func() bool
	ProtectLevel    func(uint32, uint8)
	ReadValues      func(*Object, int32) int32
	PauseFX         func(*Object, int32)
	SendLineMessage func(*Object, string) bool
}

type experienceLevelUpdateNativeDeps4EF2E0 struct {
	gameGet         func() int32
	gameSubActive   func() bool
	loadXPTable     func(string, int32) float64
	protectLevel    func(uint32, uint8)
	readValues      func(*Object, int32) int32
	gameFlag        func(uint32) int32
	pauseFX         func(*Object, int32)
	audio           func(uint32, *Object, int32, uint32)
	loadString      func(string, string, int) string
	sendLineMessage func(*Object, string) bool
}

func experienceLevelUpdateNative4EF2E0(
	unit *Object,
	deps experienceLevelUpdateNativeDeps4EF2E0,
) {
	experienceLevelUpdate4EF2E0(experienceLevelUpdateHooks4EF2E0[
		*Object,
		*PlayerUpdateData,
		*Player,
		string,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		gameGet:       deps.gameGet,
		gameSubActive: deps.gameSubActive,
		loadLevel: func(player *Player) uint8 {
			return player.Level
		},
		loadXPTable: deps.loadXPTable,
		loadExperience: func(unit *Object) float32 {
			return unit.Experience
		},
		storeLevel: func(player *Player, level uint8) {
			player.Level = level
		},
		loadLevelToken: func(player *Player) uint32 {
			return player.ProtPlayerLevel
		},
		protectLevel: deps.protectLevel,
		readValues: func(unit *Object, reward int32) {
			_ = deps.readValues(unit, reward)
		},
		gameFlag: deps.gameFlag,
		pauseFX:  deps.pauseFX,
		loadNetCode: func(unit *Object) uint32 {
			return unit.NetCode
		},
		audio:      deps.audio,
		loadString: deps.loadString,
		sendLineMessage: func(unit *Object, message string) {
			_ = deps.sendLineMessage(unit, message)
		},
	})
}

func experienceLevelUpdateServerDeps4EF2E0(
	s *Server,
	runtime ExperienceLevelUpdateRuntime4EF2E0,
) experienceLevelUpdateNativeDeps4EF2E0 {
	return experienceLevelUpdateNativeDeps4EF2E0{
		gameGet:       runtime.GameGet,
		gameSubActive: runtime.GameSubActive,
		loadXPTable: func(key string, index int32) float64 {
			return s.Balance.FloatInd(key, int(index))
		},
		protectLevel: runtime.ProtectLevel,
		readValues:   runtime.ReadValues,
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		pauseFX: runtime.PauseFX,
		audio: func(id uint32, unit *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), code)
		},
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: runtime.SendLineMessage,
	}
}

// ExperienceLevelUpdate4EF2E0 binds GAME.EXE 004EF2E0 to native-width
// Object, PlayerUpdateData, and Player layouts while retaining the original
// fixed-width game flags, sound arguments, protection token, and reward.
func (s *Server) ExperienceLevelUpdate4EF2E0(
	unit *Object,
	runtime ExperienceLevelUpdateRuntime4EF2E0,
) {
	experienceLevelUpdateNative4EF2E0(
		unit,
		experienceLevelUpdateServerDeps4EF2E0(s, runtime),
	)
}
