package server

const (
	inventoryDetachMonsterClass4ED0C0     = uint32(0x00000002)
	inventoryDetachPlayerClass4ED0C0      = uint32(0x00000004)
	inventoryDetachWeaponClass4ED0C0      = uint32(0x10000000)
	inventoryDetachReportMask4ED0C0       = uint32(0x13001000)
	inventoryDetachDeadFlag4ED0C0         = uint32(0x00008000)
	inventoryDetachNPCEquipSubclass4ED0C0 = uint8(0x10)
	inventoryDetachQuestFlag4ED0C0        = uint32(0x00001000)
	inventoryDetachOnlineFlag4ED0C0       = uint32(0x00000020)
)

// inventoryDetachHooks4ED0C0 separates the original ABI32 object, update-data,
// and Player pointer domains from the exact GAME.EXE 004ED0C0 access order.
// Argument loads are hooks because a nil owner returns before the item argument
// is read. Repeated field loads remain separate so callback mutation stays
// observable.
type inventoryDetachHooks4ED0C0[O comparable, U, P any] struct {
	loadOwnerArg func() O
	loadItemArg  func() O

	loadObjectClass    func(O) uint32
	loadObjectFlags    func(O) uint32
	loadObjectSubclass func(O) uint8
	loadObjectUpdate   func(O) U
	gameFlag           func(uint32) uint32

	loadUpdatePlayer  func(U) P
	loadPlayerField4  func(P) uint32
	storePlayerField4 func(P, uint32)
	netReportDequip   func(uint8, O)
	dequipArmor       func(O, O, int32, int32)
	dequipWeapon      func(O, O, int32, int32)
	loadPlayerIndex   func(P) uint8
	netReportDrop     func(uint8, O)
	loadPlayerProtect func(P) uint32
	protectItem       func(uint32, O)
	npcSetItemEquip   func(O, O, int32)

	loadInventoryPrev    func(O) O
	loadInventoryNext    func(O) O
	storeInventoryNext   func(O, O)
	storeInventoryPrev   func(O, O)
	loadInventoryFirst   func(O) O
	storeInventoryFirst  func(O, O)
	storeInventoryHolder func(O, O)
	clearOwner           func(O)

	loadItemWeight        func(O) uint8
	loadCarryCapacity     func(O) uint16
	storePlayerOverweight func(P, uint32)
}

func addInventoryWeight4ED0C0(sum int32, weight uint8) int32 {
	return sum + int32(weight)
}

func inventoryOverweight4ED0C0(sum int32, capacity uint16) uint32 {
	if sum > int32(capacity) {
		return 1
	}
	return 0
}

// detachInventory4ED0C0 preserves GAME.EXE 004ED0C0.
//
// The entry owner class is cached for the mutually exclusive Player/Monster
// dequip paths. The Player path caches UpdateData, but reloads its Player at
// each network/protection boundary. Inventory unlinking deliberately reloads
// the item's next and previous pointers, clears only InvHolder, then calls the
// owner-clear callback. A live Player-class check after that callback controls
// weight recomputation. The recomputation caches the post-callback UpdateData,
// reads each unsigned byte weight before its live next link, accumulates with
// int32 wrapping, and performs a signed comparison with the owner's unsigned
// 16-bit capacity.
func detachInventory4ED0C0[O comparable, U, P any](
	hooks inventoryDetachHooks4ED0C0[O, U, P],
) {
	var zero O
	owner := hooks.loadOwnerArg()
	if owner == zero {
		return
	}
	item := hooks.loadItemArg()
	if item == zero {
		return
	}

	report := int32(1)
	entryClass := hooks.loadObjectClass(owner)
	if uint8(entryClass)&uint8(inventoryDetachPlayerClass4ED0C0) != 0 {
		update := hooks.loadObjectUpdate(owner)
		if hooks.gameFlag(inventoryDetachQuestFlag4ED0C0) == 0 &&
			hooks.loadObjectFlags(owner)&inventoryDetachDeadFlag4ED0C0 == inventoryDetachDeadFlag4ED0C0 &&
			hooks.loadObjectClass(item)&inventoryDetachReportMask4ED0C0 != 0 {
			report = 0
		}
		if hooks.loadObjectClass(item)&inventoryDetachWeaponClass4ED0C0 != 0 &&
			hooks.gameFlag(inventoryDetachOnlineFlag4ED0C0) != 0 {
			player := hooks.loadUpdatePlayer(update)
			field4 := hooks.loadPlayerField4(player) &^ uint32(1)
			reportDequip := report == 1
			hooks.storePlayerField4(player, field4)
			if reportDequip {
				hooks.netReportDequip(255, item)
			}
		}
		hooks.dequipArmor(owner, item, 0, report)
		hooks.dequipWeapon(owner, item, 0, report)
		player := hooks.loadUpdatePlayer(update)
		hooks.netReportDrop(hooks.loadPlayerIndex(player), item)
		player = hooks.loadUpdatePlayer(update)
		hooks.protectItem(hooks.loadPlayerProtect(player), item)
	} else if uint8(entryClass)&uint8(inventoryDetachMonsterClass4ED0C0) != 0 {
		if hooks.loadObjectSubclass(owner)&inventoryDetachNPCEquipSubclass4ED0C0 != 0 &&
			hooks.loadObjectClass(item)&inventoryDetachWeaponClass4ED0C0 != 0 &&
			hooks.gameFlag(inventoryDetachOnlineFlag4ED0C0) != 0 {
			hooks.npcSetItemEquip(owner, item, 0)
		}
		hooks.dequipArmor(owner, item, 1, 1)
		hooks.dequipWeapon(owner, item, 1, 1)
	}

	previous := hooks.loadInventoryPrev(item)
	if previous != zero {
		hooks.storeInventoryNext(previous, hooks.loadInventoryNext(item))
	} else {
		hooks.storeInventoryFirst(owner, hooks.loadInventoryNext(item))
	}
	next := hooks.loadInventoryNext(item)
	if next != zero {
		hooks.storeInventoryPrev(next, hooks.loadInventoryPrev(item))
	}
	hooks.storeInventoryHolder(item, zero)
	hooks.clearOwner(item)

	if uint8(hooks.loadObjectClass(owner))&uint8(inventoryDetachPlayerClass4ED0C0) == 0 {
		return
	}
	current := hooks.loadInventoryFirst(owner)
	update := hooks.loadObjectUpdate(owner)
	var weight int32
	for current != zero {
		itemWeight := hooks.loadItemWeight(current)
		current = hooks.loadInventoryNext(current)
		weight = addInventoryWeight4ED0C0(weight, itemWeight)
	}
	overweight := inventoryOverweight4ED0C0(weight, hooks.loadCarryCapacity(owner))
	player := hooks.loadUpdatePlayer(update)
	hooks.storePlayerOverweight(player, overweight)
}
