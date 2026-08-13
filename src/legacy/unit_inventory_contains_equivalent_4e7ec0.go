package legacy

type inventoryContainsEquivalentHooks4E7EC0[O comparable] struct {
	loadFirst  func(O) O
	equivalent func(O, O) bool
	loadNext   func(O) O
}

// inventoryContainsEquivalent4E7EC0 preserves GAME.EXE 004E7EC0. Nil owner
// and nil item return in that order before the inventory-head load. Each
// candidate is compared by 004E7DE0 before its live successor is loaded, and
// a match returns one without observing that successor.
func inventoryContainsEquivalent4E7EC0[O comparable](
	owner O,
	item O,
	hooks inventoryContainsEquivalentHooks4E7EC0[O],
) bool {
	var zero O
	if owner == zero || item == zero {
		return false
	}

	for candidate := hooks.loadFirst(owner); candidate != zero; candidate = hooks.loadNext(candidate) {
		if hooks.equivalent(candidate, item) {
			return true
		}
	}
	return false
}
