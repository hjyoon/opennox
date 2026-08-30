package server

// inventoryContainsHooks4F78E0 keeps the object identity domain native-width
// while exposing every live load made by GAME.EXE 004F78E0.
type inventoryContainsHooks4F78E0[O comparable] struct {
	loadItemHolder  func(O) O
	loadHolderFirst func(O) O
	loadItemNext    func(O) O
}

// inventoryContains4F78E0 preserves GAME.EXE 004F78E0..004F7910. The item's
// holder link is read before the holder's inventory head. Membership requires
// both that relationship and exact object identity in the live next chain.
func inventoryContains4F78E0[O comparable](holder, item O, hooks inventoryContainsHooks4F78E0[O]) int32 {
	if hooks.loadItemHolder(item) != holder {
		return 0
	}
	current := hooks.loadHolderFirst(holder)
	var zero O
	for current != zero {
		if current == item {
			return 1
		}
		current = hooks.loadItemNext(current)
	}
	return 0
}

// equippedItemByCodeHooks4F7920 exposes the live inventory loads made by
// GAME.EXE 004F7920 without storing an object pointer in a PE32 integer slot.
type equippedItemByCodeHooks4F7920[O comparable] struct {
	loadHolderFirst func(O) O
	loadItemNetCode func(O) uint32
	loadItemNext    func(O) O
}

// equippedItemByCode4F7920 returns the exact first object identity whose live
// four-byte network code matches code, preserving GAME.EXE 004F7920..004F7943.
func equippedItemByCode4F7920[O comparable](holder O, code uint32, hooks equippedItemByCodeHooks4F7920[O]) O {
	current := hooks.loadHolderFirst(holder)
	var zero O
	for current != zero {
		if hooks.loadItemNetCode(current) == code {
			return current
		}
		current = hooks.loadItemNext(current)
	}
	return zero
}
