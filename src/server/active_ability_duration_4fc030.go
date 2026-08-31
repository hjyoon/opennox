package server

type activeAbilityDurationHooks4FC030[R comparable, U comparable] struct {
	loadHead     func() R
	loadUnit     func(R) U
	loadAbility  func(R) Ability
	loadNext     func(R) R
	loadDeadline func(R) uint32
	loadFrame    func() uint32
}

// activeAbilityDuration4FC030 preserves GAME.EXE 004FC030. Records are
// searched in the one global active-ability list. Unit identity is compared
// before the signed ability ID, and the next link is loaded only after a
// mismatch. Active and unit flags are deliberately not inspected.
//
// On a match, the deadline is loaded before the current frame and subtraction
// wraps at 32 bits. The original public result is a signed 32-bit int, so the
// same bit pattern may be negative. A miss returns the all-ones pattern.
func activeAbilityDuration4FC030[R comparable, U comparable](
	unit U,
	ability Ability,
	hooks activeAbilityDurationHooks4FC030[R, U],
) int32 {
	var zeroRecord R
	for record := hooks.loadHead(); record != zeroRecord; record = hooks.loadNext(record) {
		if hooks.loadUnit(record) == unit && hooks.loadAbility(record) == ability {
			deadline := hooks.loadDeadline(record)
			frame := hooks.loadFrame()
			return int32(deadline - frame)
		}
	}
	return -1
}
