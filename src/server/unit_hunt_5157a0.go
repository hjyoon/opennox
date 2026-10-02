package server

const (
	unitHuntMonsterClassLow5157A0 = uint8(0x02)
	unitHuntBlockedFlag5157A0     = uint32(0x00008000)
	unitHuntAction5157A0          = uint32(5)
)

// unitHuntHooks5157A0 exposes each load and call of GAME.EXE 005157A0
// without constraining object identity to the original PE32 address width.
type unitHuntHooks5157A0[O comparable] struct {
	loadClassLow     func(O) uint8
	loadFlags        func(O) uint32
	clearActionStack func(O)
	pushAction       func(O, uint32)
}

func unitHunt5157A0[O comparable](unit O, hooks unitHuntHooks5157A0[O]) {
	var nilUnit O
	if unit == nilUnit {
		return
	}
	if hooks.loadClassLow(unit)&unitHuntMonsterClassLow5157A0 == 0 {
		return
	}
	if hooks.loadFlags(unit)&unitHuntBlockedFlag5157A0 != 0 {
		return
	}
	hooks.clearActionStack(unit)
	// The original ignores the push result and does not reload either gate
	// after the clear callback. It does not inspect UpdateData itself.
	hooks.pushAction(unit, unitHuntAction5157A0)
}
