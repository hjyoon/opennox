package server

const (
	monsterWalkToBlockedFlag514110  = uint32(0x00008000)
	monsterWalkToMonsterClass514110 = uint8(0x02)
	monsterWalkToReportAction514110 = uint32(32)
	monsterWalkToMoveAction514110   = uint32(8)
)

// monsterWalkToHooks514110 exposes the observable loads, calls, and stores of
// GAME.EXE 00514110 while keeping object and action identities native-width.
type monsterWalkToHooks514110[O, A comparable] struct {
	loadFlags        func(O) uint32
	loadClassLow     func(O) uint8
	clearActionStack func(O)
	pushAction       func(O, uint32) A
	storeArgBits     func(A, int, uint32)
}

// monsterWalkTo514110 replaces an eligible monster's action stack with a
// REPORT(FAR_MOVE_TO) followed by FAR_MOVE_TO(x, y). Loads and calls retain the
// exact PE32 order, including continuing to the move action if REPORT is full.
func monsterWalkTo514110[O, A comparable](unit O, x, y uint32, hooks monsterWalkToHooks514110[O, A]) {
	if hooks.loadFlags(unit)&monsterWalkToBlockedFlag514110 != 0 {
		return
	}
	if hooks.loadClassLow(unit)&monsterWalkToMonsterClass514110 == 0 {
		return
	}
	hooks.clearActionStack(unit)
	report := hooks.pushAction(unit, monsterWalkToReportAction514110)
	var nilAction A
	if report != nilAction {
		hooks.storeArgBits(report, 0, monsterWalkToMoveAction514110)
	}
	move := hooks.pushAction(unit, monsterWalkToMoveAction514110)
	if move == nilAction {
		return
	}
	hooks.storeArgBits(move, 0, x)
	hooks.storeArgBits(move, 1, y)
	hooks.storeArgBits(move, 2, 0)
}
