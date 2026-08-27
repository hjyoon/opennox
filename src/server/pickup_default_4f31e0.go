package server

const (
	pickupDefaultQuestFlag4F31E0      = uint32(0x1000)
	pickupDefaultQuestCoopFlags4F31E0 = uint32(0x1800)
	pickupDefaultPlayerClass4F31E0    = uint8(0x04)
	pickupDefaultFoodClass4F31E0      = uint32(0x10)
	pickupDefaultTeamInform4F31E0     = uint8(16)
	pickupDefaultNormalLimit4F31E0    = int32(3)
	pickupDefaultQuestLimit4F31E0     = int32(9)
	pickupDefaultTooHeavy4F31E0       = "pickup.c:CarryingTooMuch"
	pickupDefaultTooMany4F31E0        = "pickup.c:MaxSameItem"
)

// pickupDefaultHooks4F31E0 exposes every observable call, load, and effect in
// GAME.EXE 004F31E0. The fourth callback argument is deliberately absent from
// the hooks because the original callback receives it but never reads it.
type pickupDefaultHooks4F31E0[O, T comparable, U, P any] struct {
	gameFlagsCheck func(uint32) int32
	itemHasTeam    func(O) bool
	teamsSame      func(O, O) bool
	loadTeamID     func(O) uint8
	findTeam       func(uint8) T
	loadClassLow   func(O) uint8
	loadUpdate     func(O) U
	loadTeamColor  func(T) uint8
	loadPlayer     func(U) P
	loadPlayerInd  func(P) uint8
	informTeam     func(uint8, uint8, uint32)

	loadInventoryHolder func(O) O
	loadCarryCapacity   func(O) uint16
	loadInventoryFirst  func(O) O
	loadWeight          func(O) uint8
	loadInventoryNext   func(O) O
	primaryMessage      func(O, string, uint8)

	loadItemClass     func(O) uint32
	loadItemType      func(O) uint16
	countInventory    func(O, int32) int32
	deleteWorldObject func(O)
	inventoryPut      func(O, O, int32)
}

func pickupDefaultWeightAdd4F31E0(sum uint32, weight uint8) uint32 {
	return sum + uint32(weight)
}

func pickupDefaultWeightBudget4F31E0(capacity uint16, inventoryWeight uint32) int32 {
	return int32((uint32(capacity) << 1) - inventoryWeight)
}

func pickupDefaultFoodBelowLimit4F31E0(count, limit int32) bool {
	return count-limit < 0
}

// pickupDefault4F31E0 preserves GAME.EXE 004F31E0.
//
// The game-mode query occurs before either object is read. Outside Quest, a
// mismatched live item team rejects only when that team still resolves; a
// Player receives team color information before the rejection. Inventory
// holder and capacity gates follow. Capacity is cached before the inventory
// walk, whose byte weights accumulate with 32-bit wrapping. The signed budget
// comparison uses twice the zero-extended capacity. Food count is evaluated
// only after that weight gate, with a second live game-mode query choosing the
// wrapping signed count limit. Success removes the item from the world before
// inserting it into inventory. No argument is nil-guarded.
func pickupDefault4F31E0[O, T comparable, U, P any](
	owner, item O,
	report, ignored int32,
	hooks pickupDefaultHooks4F31E0[O, T, U, P],
) int32 {
	_ = ignored
	var zeroObject O
	var zeroTeam T

	if hooks.gameFlagsCheck(pickupDefaultQuestFlag4F31E0) == 0 &&
		hooks.itemHasTeam(item) &&
		!hooks.teamsSame(owner, item) {
		teamID := hooks.loadTeamID(item)
		team := hooks.findTeam(teamID)
		if team != zeroTeam {
			if hooks.loadClassLow(owner)&pickupDefaultPlayerClass4F31E0 != 0 {
				update := hooks.loadUpdate(owner)
				color := hooks.loadTeamColor(team)
				player := hooks.loadPlayer(update)
				index := hooks.loadPlayerInd(player)
				hooks.informTeam(index, pickupDefaultTeamInform4F31E0, uint32(color))
			}
			return 0
		}
	}

	if hooks.loadInventoryHolder(item) != zeroObject {
		return 0
	}
	capacity := hooks.loadCarryCapacity(owner)
	if capacity == 0 {
		return 0
	}

	var inventoryWeight uint32
	for current := hooks.loadInventoryFirst(owner); current != zeroObject; {
		weight := hooks.loadWeight(current)
		current = hooks.loadInventoryNext(current)
		inventoryWeight = pickupDefaultWeightAdd4F31E0(inventoryWeight, weight)
	}
	itemWeight := hooks.loadWeight(item)
	if pickupDefaultWeightBudget4F31E0(capacity, inventoryWeight) < int32(itemWeight) {
		hooks.primaryMessage(owner, pickupDefaultTooHeavy4F31E0, 0)
		return 0
	}

	if hooks.loadItemClass(item)&pickupDefaultFoodClass4F31E0 == pickupDefaultFoodClass4F31E0 {
		typeInd := hooks.loadItemType(item)
		count := hooks.countInventory(owner, int32(typeInd))
		limit := pickupDefaultNormalLimit4F31E0
		if hooks.gameFlagsCheck(pickupDefaultQuestCoopFlags4F31E0) != 0 {
			limit = pickupDefaultQuestLimit4F31E0
		}
		if !pickupDefaultFoodBelowLimit4F31E0(count, limit) {
			hooks.primaryMessage(owner, pickupDefaultTooMany4F31E0, 0)
			return 0
		}
	}

	hooks.deleteWorldObject(item)
	hooks.inventoryPut(owner, item, report)
	return 1
}
