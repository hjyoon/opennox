package server

const signCollidePlayerClass4EAB40 = uint8(0x04)

type signCollideHooks4EAB40[O comparable, U any] struct {
	classLow func(O) uint8
	loadUse  func(O) U
	callUse  func(U, O, O) int32
}

// signCollide4EAB40 preserves GAME.EXE 004EAB40. A nil or non-Player
// target returns before touching the source. The Player path loads the live
// source Use callback once and invokes it as (target, source). Its return is
// ignored by the registered collision dispatcher, and collision is unobserved.
func signCollide4EAB40[O comparable, U, C any](
	source, target O,
	_ C,
	hooks signCollideHooks4EAB40[O, U],
) {
	var zero O
	if target == zero {
		return
	}
	if hooks.classLow(target)&signCollidePlayerClass4EAB40 == 0 {
		return
	}
	use := hooks.loadUse(source)
	_ = hooks.callUse(use, target, source)
}
