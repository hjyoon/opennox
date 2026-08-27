package server

const (
	rewardGemRandomPath4F1D30 = `C:\NoxPost\src\server\GameMech\Reward.c`

	rewardGemGoldGateLine4F1D30 = int32(2181)
	rewardGemTypeLine4F1D30     = int32(2185)
	rewardGemGoldTypeLine4F1D30 = int32(2197)

	rewardGemRubyType4F1D30      = "RubyGem"
	rewardGemEmeraldType4F1D30   = "EmeraldGem"
	rewardGemDiamondType4F1D30   = "DiamondGem"
	rewardGemGoldChestType4F1D30 = "QuestGoldChest"
	rewardGemGoldPileType4F1D30  = "QuestGoldPile"
)

type rewardGemGoldRange4F1D30 uint8

const (
	rewardGemGoldDefault4F1D30 rewardGemGoldRange4F1D30 = iota
	rewardGemGoldSlot2_4F1D30
	rewardGemGoldSlot4_4F1D30
	rewardGemGoldSlot8_4F1D30
	rewardGemGoldSlot16_4F1D30
)

type rewardGemGoldBounds4F1D30 struct {
	Minimum int32
	Maximum int32
	Line    int32
}

var rewardGemGoldAmountBounds4F1D30 = [...]rewardGemGoldBounds4F1D30{
	{Minimum: 100, Maximum: 250, Line: 2226},
	{Minimum: 200, Maximum: 500, Line: 2222},
	{Minimum: 400, Maximum: 1000, Line: 2219},
	{Minimum: 800, Maximum: 2000, Line: 2216},
	{Minimum: 1600, Maximum: 4000, Line: 2213},
}

type rewardGemHooks4F1D30[O, D any] struct {
	pickSlots          func(uint32) uint32
	randomInt          func(int32, int32, string, int32) int32
	createObjectByType func(string) O
	isNilObject        func(O) bool
	loadInitData       func(O) D
	loadGoldMaximum    func(rewardGemGoldRange4F1D30) int32
	loadGoldMinimum    func(rewardGemGoldRange4F1D30) int32
	storeGoldAmount    func(D, int32)
}

func rewardGemGoldRangeForSlots4F1D30(slots uint32) rewardGemGoldRange4F1D30 {
	switch slots {
	case 2:
		return rewardGemGoldSlot2_4F1D30
	case 4:
		return rewardGemGoldSlot4_4F1D30
	case 8:
		return rewardGemGoldSlot8_4F1D30
	case 16:
		return rewardGemGoldSlot16_4F1D30
	default:
		return rewardGemGoldDefault4F1D30
	}
}

// rewardGem4F1D30 reconstructs GAME.EXE 004F1D30. The marker argument is
// absent because the original never reads it. Unsigned slots below four skip
// the 90-percent gold gate. The gem branch returns the factory result without
// a nil check. The gold branch checks the factory result, caches its InitData,
// loads the selected maximum before its minimum, performs the inclusive draw,
// and stores the raw signed result bits. No callback, InitData, or range guards
// are added beyond the original branches.
func rewardGem4F1D30[O, D any](stage uint32, hooks rewardGemHooks4F1D30[O, D]) O {
	var zero O
	slots := hooks.pickSlots(stage)
	if slots >= 4 && hooks.randomInt(
		1, 100, rewardGemRandomPath4F1D30, rewardGemGoldGateLine4F1D30,
	) > 90 {
		draw := hooks.randomInt(
			1, 100, rewardGemRandomPath4F1D30, rewardGemTypeLine4F1D30,
		)
		typeName := rewardGemRubyType4F1D30
		if draw >= 50 {
			typeName = rewardGemEmeraldType4F1D30
			if draw >= 90 {
				typeName = rewardGemDiamondType4F1D30
			}
		}
		return hooks.createObjectByType(typeName)
	}

	goldType := rewardGemGoldPileType4F1D30
	if hooks.randomInt(
		1, 2, rewardGemRandomPath4F1D30, rewardGemGoldTypeLine4F1D30,
	) == 1 {
		goldType = rewardGemGoldChestType4F1D30
	}
	result := hooks.createObjectByType(goldType)
	if hooks.isNilObject(result) {
		return zero
	}
	data := hooks.loadInitData(result)
	rangeIndex := rewardGemGoldRangeForSlots4F1D30(slots)
	maximum := hooks.loadGoldMaximum(rangeIndex)
	minimum := hooks.loadGoldMinimum(rangeIndex)
	amount := hooks.randomInt(
		minimum, maximum, rewardGemRandomPath4F1D30,
		rewardGemGoldAmountBounds4F1D30[rangeIndex].Line,
	)
	hooks.storeGoldAmount(data, amount)
	return result
}

// rewardGem2_4F1F00 reconstructs the exact thin wrapper at 004F1F00.
func rewardGem2_4F1F00[O, D any](stage uint32, hooks rewardGemHooks4F1D30[O, D]) O {
	return rewardGem4F1D30(stage, hooks)
}
