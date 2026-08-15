package server

type unitTransferSlavesHooks4EC4B0[O comparable] struct {
	loadFirstOwned func(O) O
	loadNextOwned  func(O) O
	loadOwner      func(O) O
	setOwner       func(O, O)
}

// unitTransferSlaves4EC4B0 preserves GAME.EXE 004EC4B0. Each child's
// next-owned link is cached before the source's live owner is loaded and the
// child is reassigned. There is no explicit store to the source's first-owned
// field; the owner setter repairs that list as a consequence of reassignment.
func unitTransferSlaves4EC4B0[O comparable](source O, hooks unitTransferSlavesHooks4EC4B0[O]) {
	var zero O
	if source == zero {
		return
	}
	child := hooks.loadFirstOwned(source)
	for child != zero {
		next := hooks.loadNextOwned(child)
		owner := hooks.loadOwner(source)
		hooks.setOwner(owner, child)
		child = next
	}
}
