package server

const (
	godModeControllerCoopFlag4EF500 = uint32(0x800)
	godModeControllerEnable4EF500   = uint32(1)
	godModeControllerEngineMask     = uint32(0x30)

	godModeSetMessage4EF500   = "godset"
	godModeUnsetMessage4EF500 = "godunset"
)

type godModeControllerHooks4EF500[P comparable] struct {
	gameFlag         func(uint32) int32
	loadValue        func() uint32
	loadEngineFlags  func() uint32
	storeEngineFlags func(uint32)
	firstPlayer      func() P
	awardScrolls     func(P)
	awardSpells      func(P)
	awardAbilities   func(P)
	nextPlayer       func(P) P
}

// godModeController4EF500 preserves GAME.EXE 004EF500. The cooperative-game
// query is the first observation; a zero result returns without reading the
// value or engine flags. A whole input dword equal to one sets Admin and
// GodMode, while every other bit pattern clears both. All unrelated engine
// flag bits survive the full-dword load/store.
//
// The first active player is fetched only after the flag store. For every
// player, scrolls, spells, and abilities are awarded in that order. The live
// successor is queried after all three callbacks, so callback mutations can
// change which player is visited next.
func godModeController4EF500[P comparable](hooks godModeControllerHooks4EF500[P]) {
	if hooks.gameFlag(godModeControllerCoopFlag4EF500) == 0 {
		return
	}

	value := hooks.loadValue()
	flags := hooks.loadEngineFlags()
	if value == godModeControllerEnable4EF500 {
		flags |= godModeControllerEngineMask
	} else {
		flags &^= godModeControllerEngineMask
	}
	hooks.storeEngineFlags(flags)

	var nilPlayer P
	for player := hooks.firstPlayer(); player != nilPlayer; player = hooks.nextPlayer(player) {
		hooks.awardScrolls(player)
		hooks.awardSpells(player)
		hooks.awardAbilities(player)
	}
}

// GodModeCommandRuntime4EF500 supplies the observable services used by the
// original set-god and unset-god command callers at 00442410 and 00442450.
type GodModeCommandRuntime4EF500 struct {
	QuestMode  func() bool
	SetGod     func(uint32)
	LoadString func(string) string
	Print      func(string)
}

// GodModeCommand4EF500 preserves the original command callers. Set checks
// Quest first and is completely quiet in Quest; otherwise it passes exact one
// to 004EF500 before loading and printing godset. Unset never queries Quest and
// passes zero before loading and printing godunset.
func GodModeCommand4EF500(enable bool, runtime GodModeCommandRuntime4EF500) {
	if enable {
		if runtime.QuestMode() {
			return
		}
		runtime.SetGod(godModeControllerEnable4EF500)
		message := runtime.LoadString(godModeSetMessage4EF500)
		runtime.Print(message)
		return
	}

	runtime.SetGod(0)
	message := runtime.LoadString(godModeUnsetMessage4EF500)
	runtime.Print(message)
}
