package legacy

type unitOwnedTypeHooks4E7BE0[C, O comparable] struct {
	loadCache  func(C) uint32
	storeCache func(C, uint32)
	lookupType func(string) uint32
	loadFirst  func(O) O
	loadType   func(O) uint16
	loadNext   func(O) O
}

// unitHasOwnedType4E7BE0 preserves the shared owned-list contract in
// GAME.EXE 004E7BE0 and 004E7C30. The cache is loaded exactly once. A zero
// cache triggers lookup and store before the owner is dereferenced. Each
// object's zero-extended 16-bit type is compared with the complete 32-bit
// cached ID, and a match returns without loading that object's successor.
func unitHasOwnedType4E7BE0[C, O comparable](
	cache C,
	owner O,
	typeName string,
	hooks unitOwnedTypeHooks4E7BE0[C, O],
) uint32 {
	typeInd := hooks.loadCache(cache)
	if typeInd == 0 {
		typeInd = hooks.lookupType(typeName)
		hooks.storeCache(cache, typeInd)
	}

	obj := hooks.loadFirst(owner)
	var zero O
	for obj != zero {
		if uint32(hooks.loadType(obj)) == typeInd {
			return 1
		}
		obj = hooks.loadNext(obj)
	}
	return 0
}

func unitIsCrown4E7BE0[C, O comparable](cache C, owner O, hooks unitOwnedTypeHooks4E7BE0[C, O]) uint32 {
	return unitHasOwnedType4E7BE0(cache, owner, "Crown", hooks)
}

func unitIsGameBall4E7C30[C, O comparable](cache C, owner O, hooks unitOwnedTypeHooks4E7BE0[C, O]) uint32 {
	return unitHasOwnedType4E7BE0(cache, owner, "GameBall", hooks)
}
