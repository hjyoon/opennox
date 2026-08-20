package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// PlayerLevelSetRuntime4EF410 supplies the legacy-owned protection, derived-
// value, GUI-book, and PauseFX services. Object-bearing callbacks retain the
// host pointer width; protection tokens, levels, and book arguments remain
// fixed-width GAME.EXE values.
type PlayerLevelSetRuntime4EF410 struct {
	ProtectExperience func(uint32, float32)
	ProtectLevel      func(uint32, uint8)
	ReadValues        func(*Object, int32) int32
	BookAbility       func(int32, int32, int32)
	PauseFX           func(*Object, int32)
}

type playerLevelSetNativeDeps4EF410 struct {
	loadXPTable       func(string, int32) float64
	protectExperience func(uint32, float32)
	reportExperience  func(*Object)
	protectLevel      func(uint32, uint8)
	readValues        func(*Object, int32) int32
	gameFlag          func(uint32) int32
	bookAbility       func(int32, int32, int32)
	pauseFX           func(*Object, int32)
}

func playerLevelSetNative4EF410(
	unit *Object,
	level uint8,
	deps playerLevelSetNativeDeps4EF410,
) {
	playerLevelSet4EF410(playerLevelSetHooks4EF410[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadLevelArg: func() uint8 {
			return level
		},
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			// Do not use UpdateDataPlayer: 004EF410 has no class or nil gate.
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadXPTable: deps.loadXPTable,
		storeExperience: func(unit *Object, experience float32) {
			unit.Experience = experience
		},
		loadExperienceToken: func(player *Player) uint32 {
			return player.ProtUnitExperience
		},
		protectExperience: deps.protectExperience,
		reportExperience:  deps.reportExperience,
		loadLevelToken: func(player *Player) uint32 {
			return player.ProtPlayerLevel
		},
		storeLevel: func(player *Player, level uint8) {
			player.Level = level
		},
		protectLevel: deps.protectLevel,
		readValues: func(unit *Object, value int32) {
			_ = deps.readValues(unit, value)
		},
		gameFlag: deps.gameFlag,
		loadPlayerClass: func(player *Player) uint8 {
			return uint8(player.Info().PlayerClass())
		},
		loadAbilityLevel: func(player *Player, ability int32) uint32 {
			return player.SpellLvl[ability]
		},
		bookAbility: deps.bookAbility,
		pauseFX:     deps.pauseFX,
	})
}

func playerLevelSetServerDeps4EF410(
	s *Server,
	runtime PlayerLevelSetRuntime4EF410,
) playerLevelSetNativeDeps4EF410 {
	return playerLevelSetNativeDeps4EF410{
		loadXPTable: func(key string, index int32) float64 {
			return s.Balance.FloatInd(key, int(index))
		},
		protectExperience: runtime.ProtectExperience,
		reportExperience:  s.NetReportExperience,
		protectLevel:      runtime.ProtectLevel,
		readValues:        runtime.ReadValues,
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		bookAbility: runtime.BookAbility,
		pauseFX:     runtime.PauseFX,
	}
}

// PlayerLevelSet4EF410 binds GAME.EXE 004EF410 to native-width Object,
// PlayerUpdateData, Player, PlayerInfo, and ability-table layouts. The input
// remains one byte because every observable consumer in the original callee
// uses the low byte, with signed interpretation limited to clamp and index.
func (s *Server) PlayerLevelSet4EF410(
	unit *Object,
	level uint8,
	runtime PlayerLevelSetRuntime4EF410,
) {
	playerLevelSetNative4EF410(
		unit,
		level,
		playerLevelSetServerDeps4EF410(s, runtime),
	)
}
