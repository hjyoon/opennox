package server

const (
	pickupAmmoPlayerClassLow4F3B00 = uint8(0x04)
	pickupAmmoEquipMask4F3B00      = uint32(0x82)
	pickupAmmoWeaponClass4F3B00    = uint32(0x01000000)
	pickupAmmoModifierCount4F3B00  = 4
	pickupAmmoCapacity4F3B00       = uint16(250)
)

// pickupAmmoHooks4F3B00 exposes every observable load, store, and call in
// GAME.EXE 004F3B00. InitData, UseData, update, Player, and modifier entries
// have distinct types so tests can prove which native pointer is cached and
// which field is reloaded after a callback.
type pickupAmmoHooks4F3B00[
	O comparable,
	I any,
	U any,
	M comparable,
	D any,
	P any,
] struct {
	weaponEquipFlags  func(O) uint32
	loadOwnerClassLow func(O) uint8
	loadOwnerUpdate   func(O) D
	loadInventoryHead func(O) O
	loadTypeInd       func(O) uint16
	loadObjectClass   func(O) uint32
	loadInitData      func(O) I
	loadUseData       func(O) U
	loadModifier      func(I, int) M
	loadUseByte       func(U, int) uint8
	storeUseByte      func(U, int, uint8)
	loadInventoryNext func(O) O
	loadUpdatePlayer  func(D) P
	loadPlayerInd     func(P) uint8
	reportCharges     func(uint8, O, uint8, uint8)
	delayedDelete     func(O)
	pickupAudio       func(O, O)
	loadArg4          func() int32
	loadArg3          func() int32
	defaultPickup     func(O, O, int32, int32) int32
}

// pickupAmmo4F3B00 preserves GAME.EXE 004F3B00.
//
// The item equip flags are queried before the owner is read. Non-Players and
// non-ammunition items fall through to WeaponPickup with arg4 loaded before
// arg3. A Player caches UpdateData, the incoming item's ModifierInitData and
// AmmoUseData, then scans the live inventory. A candidate must have the same
// uint16 type, Weapon class, intersecting equip flags, four identical native
// modifier pointers, a clear third use byte, and a combined primary charge no
// greater than 250. Both charge bytes are merged with uint8 wrapping, reported
// through the Player reloaded from cached UpdateData, and the incoming item is
// deleted before pickup audio. The original has no nil-data guards.
func pickupAmmo4F3B00[
	O comparable,
	I any,
	U any,
	M comparable,
	D any,
	P any,
](
	owner, item O,
	hooks pickupAmmoHooks4F3B00[O, I, U, M, D, P],
) int32 {
	itemEquipFlags := hooks.weaponEquipFlags(item)
	if hooks.loadOwnerClassLow(owner)&pickupAmmoPlayerClassLow4F3B00 == 0 {
		return pickupAmmoDefault4F3B00(owner, item, hooks)
	}

	update := hooks.loadOwnerUpdate(owner)
	if itemEquipFlags&pickupAmmoEquipMask4F3B00 == 0 {
		return pickupAmmoDefault4F3B00(owner, item, hooks)
	}

	candidate := hooks.loadInventoryHead(owner)
	itemInit := hooks.loadInitData(item)
	itemUse := hooks.loadUseData(item)
	var zero O
	for candidate != zero {
		if hooks.loadTypeInd(candidate) == hooks.loadTypeInd(item) &&
			hooks.loadObjectClass(candidate)&pickupAmmoWeaponClass4F3B00 != 0 &&
			hooks.weaponEquipFlags(candidate)&itemEquipFlags != 0 {
			candidateInit := hooks.loadInitData(candidate)
			candidateUse := hooks.loadUseData(candidate)
			modifiersEqual := true
			for index := 0; index < pickupAmmoModifierCount4F3B00; index++ {
				candidateModifier := hooks.loadModifier(candidateInit, index)
				itemModifier := hooks.loadModifier(itemInit, index)
				if candidateModifier != itemModifier {
					modifiersEqual = false
				}
			}
			if modifiersEqual && hooks.loadUseByte(candidateUse, 2) == 0 {
				candidateCharge0 := hooks.loadUseByte(candidateUse, 0)
				itemCharge0 := hooks.loadUseByte(itemUse, 0)
				if uint16(candidateCharge0)+uint16(itemCharge0) <= pickupAmmoCapacity4F3B00 {
					itemCharge1 := hooks.loadUseByte(itemUse, 1)
					candidateCharge1 := hooks.loadUseByte(candidateUse, 1)
					mergedCharge1 := itemCharge1 + candidateCharge1
					hooks.storeUseByte(candidateUse, 1, mergedCharge1)

					candidateCharge0 = hooks.loadUseByte(candidateUse, 0)
					itemCharge0 = hooks.loadUseByte(itemUse, 0)
					mergedCharge0 := candidateCharge0 + itemCharge0
					hooks.storeUseByte(candidateUse, 0, mergedCharge0)

					player := hooks.loadUpdatePlayer(update)
					playerInd := hooks.loadPlayerInd(player)
					hooks.reportCharges(playerInd, candidate, mergedCharge1, mergedCharge0)
					hooks.delayedDelete(item)
					hooks.pickupAudio(owner, item)
					return 1
				}
			}
		}
		candidate = hooks.loadInventoryNext(candidate)
	}
	return pickupAmmoDefault4F3B00(owner, item, hooks)
}

func pickupAmmoDefault4F3B00[
	O comparable,
	I any,
	U any,
	M comparable,
	D any,
	P any,
](
	owner, item O,
	hooks pickupAmmoHooks4F3B00[O, I, U, M, D, P],
) int32 {
	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	return hooks.defaultPickup(owner, item, arg3, arg4)
}
