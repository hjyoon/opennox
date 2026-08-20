package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// GodModeControllerRuntime4EF500 supplies the three player callbacks owned by
// the root game package. Each callback deliberately reads EngineAdmin itself,
// matching the live flag observation made by 004EFD80, 004EFC80, and 004EFE10.
type GodModeControllerRuntime4EF500 struct {
	AwardScrolls   func(*Player)
	AwardSpells    func(*Player)
	AwardAbilities func(*Player)
}

type godModeControllerNativeDeps4EF500 struct {
	gameFlag         func(uint32) int32
	loadEngineFlags  func() uint32
	storeEngineFlags func(uint32)
	firstPlayer      func() *Player
	awardScrolls     func(*Player)
	awardSpells      func(*Player)
	awardAbilities   func(*Player)
	nextPlayer       func(*Player) *Player
}

func godModeControllerNative4EF500(
	value uint32,
	deps godModeControllerNativeDeps4EF500,
) {
	godModeController4EF500(godModeControllerHooks4EF500[*Player]{
		gameFlag: deps.gameFlag,
		loadValue: func() uint32 {
			return value
		},
		loadEngineFlags:  deps.loadEngineFlags,
		storeEngineFlags: deps.storeEngineFlags,
		firstPlayer:      deps.firstPlayer,
		awardScrolls:     deps.awardScrolls,
		awardSpells:      deps.awardSpells,
		awardAbilities:   deps.awardAbilities,
		nextPlayer:       deps.nextPlayer,
	})
}

func storeEngineFlags4EF500(target uint32) {
	current := noxflags.GetEngine()
	currentLow := noxflags.EngineFlag(uint32(current))
	targetLow := noxflags.EngineFlag(target)
	if clear := currentLow &^ targetLow; clear != 0 {
		noxflags.UnsetEngine(clear)
	}
	if set := targetLow &^ currentLow; set != 0 {
		noxflags.SetEngine(set)
	}
}

func godModeControllerServerDeps4EF500(
	s *Server,
	runtime GodModeControllerRuntime4EF500,
) godModeControllerNativeDeps4EF500 {
	return godModeControllerNativeDeps4EF500{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadEngineFlags: func() uint32 {
			return uint32(noxflags.GetEngine())
		},
		storeEngineFlags: storeEngineFlags4EF500,
		firstPlayer:      s.Players.First,
		awardScrolls:     runtime.AwardScrolls,
		awardSpells:      runtime.AwardSpells,
		awardAbilities:   runtime.AwardAbilities,
		nextPlayer:       s.Players.Next,
	}
}

// GodModeController4EF500 binds GAME.EXE 004EF500 to the native engine/game
// flags and the live active-player traversal. value remains a fixed-width
// dword so 32- and 64-bit hosts make the same exact-one decision.
func (s *Server) GodModeController4EF500(
	value uint32,
	runtime GodModeControllerRuntime4EF500,
) {
	godModeControllerNative4EF500(
		value,
		godModeControllerServerDeps4EF500(s, runtime),
	)
}
