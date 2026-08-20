package server

type unusedHealthLinksRemoveHooks4EE3E0[O, H any] struct {
	loadHealth    func(O) H
	loadNext      func(H) O
	loadPrevious  func(H) O
	storePrevious func(H, O)
	storeNext     func(H, O)
	storeHead     func(O)
}

// unusedHealthLinksRemove4EE3E0 preserves the exact memory-access order of
// the unreferenced GAME.EXE routine at 004EE3E0. HealthData +8 is an object
// next link and +12 is an object previous link in the original ABI32 list.
//
// The first half uses the HealthData loaded at entry to repair its initial
// successor. The object HealthData pointer is then loaded again, and the
// second half uses that live record to repair its predecessor or the global
// head. The removed record's own links are deliberately not cleared.
func unusedHealthLinksRemove4EE3E0[O, H comparable](
	obj O,
	hooks unusedHealthLinksRemoveHooks4EE3E0[O, H],
) {
	var nilObject O
	if obj == nilObject {
		return
	}

	health := hooks.loadHealth(obj)
	var nilHealth H
	if health == nilHealth {
		return
	}

	next := hooks.loadNext(health)
	if next != nilObject {
		nextHealth := hooks.loadHealth(next)
		previous := hooks.loadPrevious(health)
		hooks.storePrevious(nextHealth, previous)
	}

	health = hooks.loadHealth(obj)
	previous := hooks.loadPrevious(health)
	if previous != nilObject {
		previousHealth := hooks.loadHealth(previous)
		next = hooks.loadNext(health)
		hooks.storeNext(previousHealth, next)
		return
	}

	next = hooks.loadNext(health)
	hooks.storeHead(next)
}
