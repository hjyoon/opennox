package server

const rewardArmorObjectKind4F0E80 = uint32(2)

type rewardArmorModifierTable4F0E80 uint8

const (
	rewardArmorQualityTable4F0E80 rewardArmorModifierTable4F0E80 = iota
	rewardArmorMaterialTable4F0E80
	rewardArmorEnchantmentTable4F0E80
)

const (
	rewardArmorQualityBit4F0E80      = uint16(1)
	rewardArmorMaterialBit4F0E80     = uint16(2)
	rewardArmorEnchantment1Bit4F0E80 = uint16(4)
	rewardArmorEnchantment2Bit4F0E80 = uint16(8)
)

type rewardArmorHooks4F0E80[O, M any] struct {
	pickSlots func(uint32) uint32
	randomInt func(int32, int32) int32

	loadObjectName    func(int) string
	loadObjectWeight  func(int) uint8
	loadObjectTypeInd func(int) uint32
	loadObjectKind    func(int) uint32
	loadObjectSlots   func(int) uint32
	objectTypeAllowed func(uint32) bool
	armorTypeMask     func(uint32) uint32
	createObject      func(uint32) O
	isNilObject       func(O) bool

	loadModifierName         func(rewardArmorModifierTable4F0E80, int) string
	loadModifierValue        func(rewardArmorModifierTable4F0E80, int) M
	loadModifierGroup        func(rewardArmorModifierTable4F0E80, int) uint8
	loadModifierSlots        func(rewardArmorModifierTable4F0E80, int) uint32
	loadModifierExcludeArmor func(rewardArmorModifierTable4F0E80, int) uint32
	loadModifierAllowArmor   func(rewardArmorModifierTable4F0E80, int) uint32
	loadModifierAllowPos     func(rewardArmorModifierTable4F0E80, int) uint32
	applyModifiers           func(O, [4]M)
}

func rewardArmorAddWeight4F0E80(total int32, weight uint8) int32 {
	return int32(uint32(total) + uint32(weight))
}

func rewardArmorObjectEligible4F0E80[O, M any](
	index int,
	slots uint32,
	hooks rewardArmorHooks4F0E80[O, M],
) bool {
	if hooks.loadObjectKind(index)&rewardArmorObjectKind4F0E80 == 0 {
		return false
	}
	if hooks.loadObjectSlots(index)&slots == 0 {
		return false
	}
	return hooks.objectTypeAllowed(hooks.loadObjectTypeInd(index))
}

func rewardArmorSelectObject4F0E80[O, M any](
	slots uint32,
	hooks rewardArmorHooks4F0E80[O, M],
) (int, bool) {
	if hooks.loadObjectName(0) == "" {
		return 0, false
	}
	var total int32
	for index := 0; ; index++ {
		if rewardArmorObjectEligible4F0E80(index, slots, hooks) {
			total = rewardArmorAddWeight4F0E80(total, hooks.loadObjectWeight(index))
		}
		if hooks.loadObjectName(index+1) == "" {
			break
		}
	}
	if total == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, total-1)
	if hooks.loadObjectName(0) == "" {
		return 0, false
	}
	var cumulative int32
	for index := 0; ; index++ {
		if rewardArmorObjectEligible4F0E80(index, slots, hooks) {
			cumulative = rewardArmorAddWeight4F0E80(cumulative, hooks.loadObjectWeight(index))
			if draw < cumulative {
				return index, true
			}
		}
		if hooks.loadObjectName(index+1) == "" {
			return 0, false
		}
	}
}

func rewardArmorModifierEligible4F0E80[O, M any](
	table rewardArmorModifierTable4F0E80,
	index int,
	slots uint32,
	armorMask uint32,
	position uint32,
	hooks rewardArmorHooks4F0E80[O, M],
) bool {
	if hooks.loadModifierSlots(table, index)&slots == 0 {
		return false
	}
	if hooks.loadModifierAllowArmor(table, index)&armorMask == 0 {
		return false
	}
	if hooks.loadModifierExcludeArmor(table, index)&armorMask != 0 {
		return false
	}
	return position == 0 || hooks.loadModifierAllowPos(table, index)&position != 0
}

func rewardArmorCountModifiers4F0E80[O, M any](
	table rewardArmorModifierTable4F0E80,
	slots uint32,
	armorMask uint32,
	position uint32,
	hooks rewardArmorHooks4F0E80[O, M],
) int32 {
	if hooks.loadModifierName(table, 0) == "" {
		return 0
	}
	var count int32
	for index := 0; ; index++ {
		if rewardArmorModifierEligible4F0E80(table, index, slots, armorMask, position, hooks) {
			count++
		}
		if hooks.loadModifierName(table, index+1) == "" {
			return count
		}
	}
}

func rewardArmorSelectModifier4F0E80[O, M any](
	table rewardArmorModifierTable4F0E80,
	slots uint32,
	armorMask uint32,
	position uint32,
	hooks rewardArmorHooks4F0E80[O, M],
) (int, bool) {
	count := rewardArmorCountModifiers4F0E80(table, slots, armorMask, position, hooks)
	if count == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, count-1)
	if hooks.loadModifierName(table, 0) == "" {
		return 0, false
	}
	var ordinal int32
	for index := 0; ; index++ {
		if rewardArmorModifierEligible4F0E80(table, index, slots, armorMask, position, hooks) {
			if ordinal == draw {
				return index, true
			}
			ordinal++
		}
		if hooks.loadModifierName(table, index+1) == "" {
			return 0, false
		}
	}
}

func rewardArmorModifierCount4F0E80(
	slots uint32,
	randomInt func(int32, int32) int32,
) (int32, bool) {
	switch slots {
	case 2:
		return randomInt(0, 1), true
	case 4:
		return randomInt(0, 2), true
	case 8:
		return randomInt(1, 3), true
	case 16:
		return randomInt(2, 4), true
	default:
		return 0, false
	}
}

func rewardArmorCategoryMask4F0E80(
	count int32,
	stage uint32,
	randomInt func(int32, int32) int32,
) uint16 {
	switch count {
	case 1:
		roll := randomInt(1, 100)
		if roll <= 20 {
			return rewardArmorEnchantment1Bit4F0E80
		}
		if roll <= 50 {
			return rewardArmorQualityBit4F0E80
		}
		return rewardArmorMaterialBit4F0E80
	case 2:
		roll := randomInt(1, 100)
		if roll <= 12 {
			return rewardArmorQualityBit4F0E80 | rewardArmorEnchantment1Bit4F0E80
		}
		if roll <= 25 {
			return rewardArmorMaterialBit4F0E80 | rewardArmorEnchantment1Bit4F0E80
		}
		return rewardArmorQualityBit4F0E80 | rewardArmorMaterialBit4F0E80
	case 3:
		return rewardArmorQualityBit4F0E80 | rewardArmorMaterialBit4F0E80 | rewardArmorEnchantment1Bit4F0E80
	case 4:
		return rewardArmorQualityBit4F0E80 | rewardArmorMaterialBit4F0E80 |
			rewardArmorEnchantment1Bit4F0E80 | rewardArmorEnchantment2Bit4F0E80
	default:
		return uint16(stage)
	}
}

func rewardArmorPromoteMask4F0E80[O, M any](
	mask uint16,
	slots uint32,
	armorMask uint32,
	hooks rewardArmorHooks4F0E80[O, M],
) uint16 {
	if mask&rewardArmorQualityBit4F0E80 != 0 &&
		rewardArmorCountModifiers4F0E80(rewardArmorQualityTable4F0E80, slots, armorMask, 0, hooks) == 0 {
		if mask&rewardArmorMaterialBit4F0E80 != 0 {
			if mask&rewardArmorEnchantment1Bit4F0E80 != 0 {
				if mask&rewardArmorEnchantment2Bit4F0E80 == 0 {
					mask |= rewardArmorEnchantment2Bit4F0E80
				}
			} else {
				mask |= rewardArmorEnchantment1Bit4F0E80
			}
		} else {
			mask |= rewardArmorMaterialBit4F0E80
		}
		mask &^= rewardArmorQualityBit4F0E80
	}
	if mask&rewardArmorMaterialBit4F0E80 != 0 &&
		rewardArmorCountModifiers4F0E80(rewardArmorMaterialTable4F0E80, slots, armorMask, 0, hooks) == 0 {
		if mask&rewardArmorEnchantment1Bit4F0E80 != 0 {
			if mask&rewardArmorEnchantment2Bit4F0E80 == 0 {
				mask |= rewardArmorEnchantment2Bit4F0E80
			}
		} else {
			mask |= rewardArmorEnchantment1Bit4F0E80
		}
		mask &^= rewardArmorMaterialBit4F0E80
	}
	if mask&rewardArmorEnchantment1Bit4F0E80 != 0 &&
		rewardArmorCountModifiers4F0E80(rewardArmorEnchantmentTable4F0E80, slots, armorMask, 1, hooks) == 0 {
		mask &^= rewardArmorEnchantment1Bit4F0E80 | rewardArmorEnchantment2Bit4F0E80
	}
	return mask
}

func rewardArmorSecondEnchantmentSlots4F0E80(
	stage uint32,
	randomInt func(int32, int32) int32,
) uint32 {
	base := stage >> 1
	if base < 1 {
		base = 1
	} else if base >= 5 {
		base = 4
	}
	minimum := base - 1
	if minimum < 1 {
		minimum = 1
	}
	maximum := base + 1
	if maximum >= 5 {
		maximum = 4
	}
	return uint32(randomInt(int32(minimum), int32(maximum)))
}

// rewardArmor4F0E80 reconstructs GAME.EXE 004F0E80. Object and modifier
// tables are deliberately read in separate count and selection passes around
// RNG calls. Object weights use only their low byte and wrap as signed int32.
// Modifier descriptor fields are loaded without nil guards, matching the
// original initialized-table contract. The marker argument at the legacy ABI
// boundary is unused by the original function and therefore is not represented
// here.
func rewardArmor4F0E80[O, M any](stage uint32, hooks rewardArmorHooks4F0E80[O, M]) O {
	var zero O
	slots := hooks.pickSlots(stage)
	objectIndex, selected := rewardArmorSelectObject4F0E80(slots, hooks)
	if !selected {
		return zero
	}
	typeInd := hooks.loadObjectTypeInd(objectIndex)
	if typeInd == 0 {
		return zero
	}
	armorMask := hooks.armorTypeMask(typeInd)
	object := hooks.createObject(typeInd)
	if hooks.isNilObject(object) {
		return zero
	}
	modifierCount, recognizedSlots := rewardArmorModifierCount4F0E80(slots, hooks.randomInt)
	if !recognizedSlots || modifierCount == 0 {
		return object
	}

	mask := rewardArmorCategoryMask4F0E80(modifierCount, stage, hooks.randomInt)
	mask = rewardArmorPromoteMask4F0E80(mask, slots, armorMask, hooks)
	if mask == 0 {
		return object
	}

	var modifiers [4]M
	firstEnchantmentIndex := 0
	if mask&rewardArmorQualityBit4F0E80 != 0 {
		if index, ok := rewardArmorSelectModifier4F0E80(
			rewardArmorQualityTable4F0E80, slots, armorMask, 0, hooks,
		); ok {
			modifiers[0] = hooks.loadModifierValue(rewardArmorQualityTable4F0E80, index)
		}
	}
	if mask&rewardArmorMaterialBit4F0E80 != 0 {
		if index, ok := rewardArmorSelectModifier4F0E80(
			rewardArmorMaterialTable4F0E80, slots, armorMask, 0, hooks,
		); ok {
			modifiers[1] = hooks.loadModifierValue(rewardArmorMaterialTable4F0E80, index)
		}
	}
	if mask&rewardArmorEnchantment1Bit4F0E80 != 0 {
		if index, ok := rewardArmorSelectModifier4F0E80(
			rewardArmorEnchantmentTable4F0E80, slots, armorMask, 1, hooks,
		); ok {
			firstEnchantmentIndex = index
			modifiers[2] = hooks.loadModifierValue(rewardArmorEnchantmentTable4F0E80, index)
		}
	}
	if mask&rewardArmorEnchantment2Bit4F0E80 != 0 {
		secondSlots := rewardArmorSecondEnchantmentSlots4F0E80(stage, hooks.randomInt)
		if index, ok := rewardArmorSelectModifier4F0E80(
			rewardArmorEnchantmentTable4F0E80, secondSlots, armorMask, 2, hooks,
		); ok && hooks.loadModifierGroup(rewardArmorEnchantmentTable4F0E80, firstEnchantmentIndex) !=
			hooks.loadModifierGroup(rewardArmorEnchantmentTable4F0E80, index) {
			modifiers[3] = hooks.loadModifierValue(rewardArmorEnchantmentTable4F0E80, index)
		}
	}
	hooks.applyModifiers(object, modifiers)
	return object
}
