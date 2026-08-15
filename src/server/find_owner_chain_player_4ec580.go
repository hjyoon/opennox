package server

type findOwnerChainPlayerHooks4EC580[O comparable] struct {
	owner    func(O) O
	classLow func(O) uint8
}

// findOwnerChainPlayer4EC580 preserves GAME.EXE 004EC580. The owner link is
// cached before the current object's live class byte is read. A terminal
// object is returned without reading its class, while a Player that itself
// has an owner is returned before following the already-cached owner link.
// A cycle containing neither a terminal object nor a Player remains cyclic.
func findOwnerChainPlayer4EC580[O comparable](
	obj O,
	hooks findOwnerChainPlayerHooks4EC580[O],
) O {
	var zero O
	if obj == zero {
		return zero
	}
	for current := obj; ; {
		owner := hooks.owner(current)
		if owner == zero {
			return current
		}
		if hooks.classLow(current)&0x4 != 0 {
			return current
		}
		current = owner
	}
}
