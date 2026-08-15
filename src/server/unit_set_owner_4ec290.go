package server

const (
	unitSetOwnerDestroyedFlag4EC290 = uint32(0x20)
	unitSetOwnerMonsterClass4EC290  = uint32(0x02)
	unitSetOwnerUnitClassMask4EC290 = uint32(0x06)
)

type unitSetOwnerHooks4EC290[O comparable] struct {
	clearOwner      func(O)
	loadFlags       func(O) uint32
	loadOwner       func(O) O
	loadFirstOwned  func(O) O
	storeNextOwned  func(O, O)
	storeFirstOwned func(O, O)
	storeOwner      func(O, O)
	loadClass       func(O) uint32
	resetMonster    func(O)
	markUnitUpdate  func(O)
}

// unitSetOwner4EC290 preserves GAME.EXE 004EC290. The first argument is the
// prospective owner and the second is the object being attached. A nil object
// is the only path that skips clearOwner. Destroyed owners are replaced by
// their live owner links before the object is inserted at the head of the
// surviving owner's owned-object list. The second class load deliberately
// happens after resetMonster because that callback may mutate the object.
func unitSetOwner4EC290[O comparable](owner, obj O, hooks unitSetOwnerHooks4EC290[O]) {
	var zero O
	if obj == zero {
		return
	}

	hooks.clearOwner(obj)
	if owner != zero {
		for hooks.loadFlags(owner)&unitSetOwnerDestroyedFlag4EC290 != 0 {
			owner = hooks.loadOwner(owner)
			if owner == zero {
				break
			}
		}
		if owner != zero {
			first := hooks.loadFirstOwned(owner)
			hooks.storeNextOwned(obj, first)
			hooks.storeFirstOwned(owner, obj)
		}
	}

	hooks.storeOwner(obj, owner)
	if hooks.loadClass(obj)&unitSetOwnerMonsterClass4EC290 != 0 {
		hooks.resetMonster(obj)
	}
	if hooks.loadClass(obj)&unitSetOwnerUnitClassMask4EC290 != 0 {
		hooks.markUnitUpdate(obj)
	}
}
