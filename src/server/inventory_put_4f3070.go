package server

const (
	inventoryPutDestroyedFlag4F3070 = uint8(0x20)
	inventoryPutPlayerClass4F3070   = uint8(0x04)
	inventoryPutAudioClass4F3070    = uint8(0x40)
	inventoryPutSound4F3070         = int32(820)
)

// inventoryPutHooks4F3070 keeps the original object, update-data, and Player
// pointer domains separate while exposing every observable load, store, and
// callback in GAME.EXE 004F3070. Repeated loads are deliberately not folded:
// callbacks may mutate inventory links, player fields, class bits, or capacity.
type inventoryPutHooks4F3070[O comparable, U, P any] struct {
	loadFlagsLow func(O) uint8
	loadClassLow func(O) uint8
	loadUpdate   func(O) U
	loadPlayer   func(U) P

	storeInventoryPrev   func(O, O)
	loadInventoryFirst   func(O) O
	storeInventoryNext   func(O, O)
	storeInventoryFirst  func(O, O)
	storeInventoryHolder func(O, O)
	setOwner             func(O, O)

	loadPlayerIndex   func(P) uint8
	reportPickup      func(uint8, O)
	loadPlayerProtect func(P) uint32
	protectItem       func(uint32, O)

	loadItemWeight        func(O) uint8
	loadInventoryNext     func(O) O
	loadCarryCapacity     func(O) uint16
	storePlayerOverweight func(P, uint32)
	audioEvent            func(int32, O, int32, uint32)
}

func addInventoryPutWeight4F3070(sum int32, weight uint8) int32 {
	return sum + int32(weight)
}

func inventoryPutOverweight4F3070(sum int32, capacity uint16) uint32 {
	if sum > int32(capacity) {
		return 1
	}
	return 0
}

// inventoryPut4F3070 preserves GAME.EXE 004F3070.
//
// Nil and destroyed arguments return before mutation. The item is prepended to
// the owner's doubly linked inventory using two independent head loads, then
// holder and object ownership are published before any player work. A live
// owner-class byte controls the Player path. That path loads UpdateData.Player
// without guards even when reporting is disabled, optionally reports pickup,
// always protects the item, and then recomputes encumbrance from the live head.
// Weight is read as an unsigned byte, accumulated with signed 32-bit wrapping,
// and compared signed against the zero-extended unsigned 16-bit capacity. The
// final item-class byte is live and controls the pickup sound after all earlier
// callbacks and stores.
func inventoryPut4F3070[O comparable, U, P any](
	owner, item O,
	report int32,
	hooks inventoryPutHooks4F3070[O, U, P],
) {
	var zero O
	if owner == zero || item == zero {
		return
	}
	if hooks.loadFlagsLow(owner)&inventoryPutDestroyedFlag4F3070 != 0 {
		return
	}
	if hooks.loadFlagsLow(item)&inventoryPutDestroyedFlag4F3070 != 0 {
		return
	}

	hooks.storeInventoryPrev(item, zero)
	first := hooks.loadInventoryFirst(owner)
	hooks.storeInventoryNext(item, first)
	first = hooks.loadInventoryFirst(owner)
	if first != zero {
		hooks.storeInventoryPrev(first, item)
	}
	hooks.storeInventoryFirst(owner, item)
	hooks.storeInventoryHolder(item, owner)
	hooks.setOwner(owner, item)

	if hooks.loadClassLow(owner)&inventoryPutPlayerClass4F3070 != 0 {
		update := hooks.loadUpdate(owner)
		player := hooks.loadPlayer(update)
		if report != 0 {
			hooks.reportPickup(hooks.loadPlayerIndex(player), item)
		}
		hooks.protectItem(hooks.loadPlayerProtect(player), item)

		var weight int32
		current := hooks.loadInventoryFirst(owner)
		for current != zero {
			itemWeight := hooks.loadItemWeight(current)
			current = hooks.loadInventoryNext(current)
			weight = addInventoryPutWeight4F3070(weight, itemWeight)
		}
		overweight := inventoryPutOverweight4F3070(weight, hooks.loadCarryCapacity(owner))
		hooks.storePlayerOverweight(player, overweight)
	}

	if hooks.loadClassLow(item)&inventoryPutAudioClass4F3070 != 0 {
		hooks.audioEvent(inventoryPutSound4F3070, owner, 0, 0)
	}
}
