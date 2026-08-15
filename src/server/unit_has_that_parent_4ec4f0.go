package server

type unitHasThatParentHooks4EC4F0[O comparable] struct {
	loadOwner func(O) O
}

// unitHasThatParent4EC4F0 preserves GAME.EXE 004EC4F0. Both arguments are
// rejected before any owner link is read. The starting object participates in
// the identity test, and a matching object is returned without reading its
// owner link.
func unitHasThatParent4EC4F0[O comparable](
	obj, owner O,
	hooks unitHasThatParentHooks4EC4F0[O],
) bool {
	var zero O
	if obj == zero || owner == zero {
		return false
	}
	for current := obj; current != zero; current = hooks.loadOwner(current) {
		if current == owner {
			return true
		}
	}
	return false
}
