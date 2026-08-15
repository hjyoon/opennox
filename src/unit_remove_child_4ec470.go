package opennox

type unitRemoveChildHooks4EC470[O comparable] struct {
	loadFirstOwned  func(O) O
	loadNextOwned   func(O) O
	storeOwner      func(O, O)
	storeNextOwned  func(O, O)
	storeFirstOwned func(O, O)
}

// unitRemoveChildContract4EC470 preserves GAME.EXE 004EC470. Each child's
// next-owned link is cached before its owner and next-owned fields are cleared.
// The parent's first-owned field is cleared last, including for an empty list.
func unitRemoveChildContract4EC470[O comparable](parent O, hooks unitRemoveChildHooks4EC470[O]) {
	var zero O
	if parent == zero {
		return
	}
	child := hooks.loadFirstOwned(parent)
	for child != zero {
		next := hooks.loadNextOwned(child)
		hooks.storeOwner(child, zero)
		hooks.storeNextOwned(child, zero)
		child = next
	}
	hooks.storeFirstOwned(parent, zero)
}
