package legacy

// scriptHostStatusDeps5166A0 exposes every observable lookup, pointer load,
// status load, and script-stack write shared by GAME.EXE 005166A0 and
// 005166E0. The only nil guard in either original function is the one around
// the host Player pointer; the unit and update-data loads deliberately remain
// eager.
type scriptHostStatusDeps5166A0[P, O, U any] struct {
	hostPlayer       func() P
	playerIsNil      func(P) bool
	loadUnit         func(P) O
	loadUpdate       func(O) U
	loadStateNonzero func(U) bool
	push             func(int32)
}

// scriptHostStatus5166A0 preserves the common IsTalking/PlayerIsTrading
// contract. Both builtins push a canonical boolean and return canonical zero.
func scriptHostStatus5166A0[P, O, U any](deps scriptHostStatusDeps5166A0[P, O, U]) int32 {
	player := deps.hostPlayer()
	value := int32(0)
	if !deps.playerIsNil(player) {
		unit := deps.loadUnit(player)
		update := deps.loadUpdate(unit)
		if deps.loadStateNonzero(update) {
			value = 1
		}
	}
	deps.push(value)
	return 0
}
