package server

const (
	unitFollowMonsterClassLow5158C0 = uint8(0x02)
	unitFollowBlockedFlag5158C0     = uint32(0x00008000)
	unitFollowAction5158C0          = uint32(3)
)

// unitFollowHooks5158C0 exposes every observable load, call, and store made by
// GAME.EXE 005158C0. Generic handles keep object and action identities native
// width while the copied position fields retain their exact PE32 bit patterns.
type unitFollowHooks5158C0[O, A comparable] struct {
	loadClassLow     func(O) uint8
	loadFlags        func(O) uint32
	clearActionStack func(O)
	pushAction       func(O, uint32) A
	loadPosXBits     func(O) uint32
	storeArgBits     func(A, int, uint32)
	loadPosYBits     func(O) uint32
	storeTarget      func(A, int, O)
}

// unitFollow5158C0 replaces an eligible monster's action stack with ESCORT,
// then initializes the new action from the live target. In particular, target
// position loads happen after both action-stack calls, as in the PE32 oracle.
func unitFollow5158C0[O, A comparable](unit, target O, hooks unitFollowHooks5158C0[O, A]) {
	var nilObject O
	if unit == nilObject {
		return
	}
	if target == nilObject {
		return
	}
	if hooks.loadClassLow(unit)&unitFollowMonsterClassLow5158C0 == 0 {
		return
	}
	if unit == target {
		return
	}
	if hooks.loadFlags(unit)&unitFollowBlockedFlag5158C0 != 0 {
		return
	}
	hooks.clearActionStack(unit)
	action := hooks.pushAction(unit, unitFollowAction5158C0)
	var nilAction A
	if action == nilAction {
		return
	}
	x := hooks.loadPosXBits(target)
	hooks.storeArgBits(action, 0, x)
	y := hooks.loadPosYBits(target)
	hooks.storeArgBits(action, 1, y)
	hooks.storeTarget(action, 2, target)
}
