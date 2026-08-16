package server

// dropOwnedCrownsHooks4ED050 separates the original 32-bit Crown type cache
// from native-width object, update-data, and position identities. Argument
// loads are hooks because GAME.EXE 004ED050 initializes the type cache before
// reading the owner, and does not read the pickup target for an empty owned
// list.
type dropOwnedCrownsHooks4ED050[O comparable, D, P any] struct {
	loadCrownTypeCache  func() uint32
	lookupCrownType     func() uint32
	storeCrownTypeCache func(uint32)
	loadOwnerArg        func() O
	firstOwned          func(O) O
	loadTargetArg       func() O
	loadTypeIndex       func(O) uint16
	loadUpdate          func(O) D
	ownerPosition       func(O) P
	dropCrown           func(O, O, P) uint32
	storePickupTarget   func(D, O)
	nextOwned           func(O) O
}

// dropOwnedCrowns4ED050 preserves GAME.EXE 004ED050.
//
// The shared Crown type cache is read before either argument. A zero cache is
// resolved and stored even when resolution returns zero. The owner's first
// owned object is then loaded; an empty list returns without reading the
// pickup-target argument. For every candidate the cache is reloaded before
// the zero-extended 16-bit type index. Matching candidates cache UpdateData
// before CrownDrop, store the target through that cached record after the
// callback, and only then read the candidate's live next-owned link. All
// matches are processed; CrownDrop's 32-bit result is deliberately ignored.
func dropOwnedCrowns4ED050[O comparable, D, P any](
	hooks dropOwnedCrownsHooks4ED050[O, D, P],
) {
	if hooks.loadCrownTypeCache() == 0 {
		crownType := hooks.lookupCrownType()
		hooks.storeCrownTypeCache(crownType)
	}

	owner := hooks.loadOwnerArg()
	item := hooks.firstOwned(owner)
	var zero O
	if item == zero {
		return
	}
	target := hooks.loadTargetArg()
	for item != zero {
		crownType := hooks.loadCrownTypeCache()
		if uint32(hooks.loadTypeIndex(item)) == crownType {
			update := hooks.loadUpdate(item)
			position := hooks.ownerPosition(owner)
			_ = hooks.dropCrown(owner, item, position)
			hooks.storePickupTarget(update, target)
		}
		item = hooks.nextOwned(item)
	}
}
