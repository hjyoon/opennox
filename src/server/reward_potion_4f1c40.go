package server

const rewardPotionObjectKind4F1C40 = uint8(4)

type rewardPotionHooks4F1C40[O any] struct {
	pickSlots         func(uint32) uint32
	loadObjectName    func(int) string
	loadObjectWeight  func(int) uint8
	loadObjectType    func(int) uint32
	loadObjectKind    func(int) uint8
	loadObjectSlots   func(int) uint32
	objectTypeAllowed func(uint32) bool
	randomInt         func(int32, int32) int32
	createObject      func(uint32) O
}

func rewardPotionAddWeight4F1C40(total int32, weight uint8) int32 {
	return int32(uint32(total) + uint32(weight))
}

func rewardPotionObjectEligible4F1C40[O any](
	index int,
	slots uint32,
	hooks rewardPotionHooks4F1C40[O],
) bool {
	if hooks.loadObjectKind(index)&rewardPotionObjectKind4F1C40 == 0 {
		return false
	}
	if hooks.loadObjectSlots(index)&slots == 0 {
		return false
	}
	return hooks.objectTypeAllowed(hooks.loadObjectType(index))
}

// rewardPotion4F1C40 reconstructs GAME.EXE 004F1C40. The stage-slot
// selector runs before the first table-head check. Object definitions are
// traversed in two independent live passes around the RNG callback. Weight is
// loaded only after the kind, slot, and type-allowed gates; its low byte is
// accumulated with signed int32 wrapping and the draw comparison is signed.
// The selected type ID is loaded again after the second-pass weight crosses
// the draw. There are deliberately no callback, table, or factory guards
// beyond the original zero branches.
func rewardPotion4F1C40[O any](stage uint32, hooks rewardPotionHooks4F1C40[O]) O {
	var zero O
	slots := hooks.pickSlots(stage)
	if hooks.loadObjectName(0) == "" {
		return zero
	}
	var total int32
	for index := 0; ; index++ {
		if rewardPotionObjectEligible4F1C40(index, slots, hooks) {
			total = rewardPotionAddWeight4F1C40(total, hooks.loadObjectWeight(index))
		}
		if hooks.loadObjectName(index+1) == "" {
			break
		}
	}
	if total == 0 {
		return zero
	}
	draw := hooks.randomInt(0, total-1)
	if hooks.loadObjectName(0) == "" {
		return zero
	}
	var cumulative int32
	for index := 0; ; index++ {
		if rewardPotionObjectEligible4F1C40(index, slots, hooks) {
			cumulative = rewardPotionAddWeight4F1C40(cumulative, hooks.loadObjectWeight(index))
			if draw < cumulative {
				typeInd := hooks.loadObjectType(index)
				if typeInd == 0 {
					return zero
				}
				return hooks.createObject(typeInd)
			}
		}
		if hooks.loadObjectName(index+1) == "" {
			return zero
		}
	}
}
