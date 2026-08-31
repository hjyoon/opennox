package server

type activeAbilityDeadlineHooks4FC070[R comparable, U comparable] struct {
	loadHead       func() R
	loadAbilityArg func() Ability
	loadUnitArg    func() U
	loadUnit       func(R) U
	loadAbility    func(R) Ability
	loadNext       func(R) R
	loadDeltaArg   func() int32
	loadFrame      func() uint32
	storeDeadline  func(R, uint32)
}

// activeAbilityDeadline4FC070 preserves GAME.EXE 004FC070. The global list
// head is loaded first; an empty list returns without reading any argument.
// Otherwise the signed ability argument is read before the unit argument.
// Records compare native unit identity before signed ability ID, and the next
// link is loaded only after a mismatch. Active and unit flags are deliberately
// not inspected.
//
// On a match, the signed delta is loaded before the current frame. Addition
// wraps at 32 bits and the resulting deadline is stored exactly once.
func activeAbilityDeadline4FC070[R comparable, U comparable](
	hooks activeAbilityDeadlineHooks4FC070[R, U],
) {
	var zeroRecord R
	record := hooks.loadHead()
	if record == zeroRecord {
		return
	}

	ability := hooks.loadAbilityArg()
	unit := hooks.loadUnitArg()
	for {
		if hooks.loadUnit(record) == unit && hooks.loadAbility(record) == ability {
			delta := hooks.loadDeltaArg()
			frame := hooks.loadFrame()
			hooks.storeDeadline(record, frame+uint32(delta))
			return
		}
		record = hooks.loadNext(record)
		if record == zeroRecord {
			return
		}
	}
}
