package server

const (
	rewardWeaponObjectKind4F14E0  = uint32(1)
	rewardWeaponWandClass4F14E0   = uint32(0x1000)
	rewardWeaponWandSubMask4F14E0 = uint32(0x047f0000)
)

type rewardWeaponModifierTable4F14E0 uint8

const (
	rewardWeaponPowerTable4F14E0 rewardWeaponModifierTable4F14E0 = iota
	rewardWeaponMaterialTable4F14E0
	rewardWeaponEnchantmentTable4F14E0
)

const (
	rewardWeaponPowerBit4F14E0        = uint16(1)
	rewardWeaponMaterialBit4F14E0     = uint16(2)
	rewardWeaponEnchantment1Bit4F14E0 = uint16(4)
	rewardWeaponEnchantment2Bit4F14E0 = uint16(8)
)

type rewardWeaponHooks4F14E0[O, M any] struct {
	pickSlots func(uint32) uint32
	randomInt func(int32, int32) int32

	loadObjectName     func(int) string
	loadObjectWeight   func(int) uint8
	loadObjectTypeInd  func(int) uint32
	loadObjectKind     func(int) uint32
	loadObjectSlots    func(int) uint32
	objectTypeAllowed  func(uint32) bool
	weaponTypeMask     func(uint32) uint32
	createObject       func(uint32) O
	isNilObject        func(O) bool
	loadObjectClass    func(O) uint32
	loadObjectSubClass func(O) uint32

	loadModifierName          func(rewardWeaponModifierTable4F14E0, int) string
	loadModifierValue         func(rewardWeaponModifierTable4F14E0, int) M
	loadModifierGroup         func(rewardWeaponModifierTable4F14E0, int) uint8
	loadModifierSlots         func(rewardWeaponModifierTable4F14E0, int) uint32
	loadModifierExcludeWeapon func(rewardWeaponModifierTable4F14E0, int) uint32
	modifierAllowWeapons      func(M) uint32
	modifierAllowPos          func(M) uint32
	loadReplenishment         func() M
	applyModifiers            func(O, [4]M)
}

func rewardWeaponAddWeight4F14E0(total int32, weight uint8) int32 {
	return int32(uint32(total) + uint32(weight))
}

func rewardWeaponObjectEligible4F14E0[O, M any](
	index int,
	slots uint32,
	hooks rewardWeaponHooks4F14E0[O, M],
) bool {
	if hooks.loadObjectKind(index)&rewardWeaponObjectKind4F14E0 == 0 {
		return false
	}
	if hooks.loadObjectSlots(index)&slots == 0 {
		return false
	}
	return hooks.objectTypeAllowed(hooks.loadObjectTypeInd(index))
}

// rewardWeaponSelectObject4F14E0 preserves the two independent table passes
// around the object RNG. If the live second pass reaches its sentinel without
// selecting a row, GAME.EXE uses the stage dword as a fallback type ID and
// leaves the row index pointing at that sentinel.
func rewardWeaponSelectObject4F14E0[O, M any](
	stage uint32,
	slots uint32,
	hooks rewardWeaponHooks4F14E0[O, M],
) (index int, typeInd uint32, ok bool) {
	if hooks.loadObjectName(0) == "" {
		return 0, 0, false
	}
	var total int32
	for index := 0; ; index++ {
		if rewardWeaponObjectEligible4F14E0(index, slots, hooks) {
			total = rewardWeaponAddWeight4F14E0(total, hooks.loadObjectWeight(index))
		}
		if hooks.loadObjectName(index+1) == "" {
			break
		}
	}
	if total == 0 {
		return 0, 0, false
	}
	draw := hooks.randomInt(0, total-1)
	if hooks.loadObjectName(0) == "" {
		return 0, stage, stage != 0
	}
	var cumulative int32
	for index := 0; ; index++ {
		if rewardWeaponObjectEligible4F14E0(index, slots, hooks) {
			cumulative = rewardWeaponAddWeight4F14E0(cumulative, hooks.loadObjectWeight(index))
			if draw < cumulative {
				typeInd := hooks.loadObjectTypeInd(index)
				return index, typeInd, typeInd != 0
			}
		}
		if hooks.loadObjectName(index+1) == "" {
			return index + 1, stage, stage != 0
		}
	}
}

func rewardWeaponModifierEligible4F14E0[O, M any](
	table rewardWeaponModifierTable4F14E0,
	index int,
	slots uint32,
	weaponMask uint32,
	position uint32,
	hooks rewardWeaponHooks4F14E0[O, M],
) bool {
	if hooks.loadModifierSlots(table, index)&slots == 0 {
		return false
	}
	modifier := hooks.loadModifierValue(table, index)
	if hooks.modifierAllowWeapons(modifier)&weaponMask == 0 {
		return false
	}
	if hooks.loadModifierExcludeWeapon(table, index)&weaponMask != 0 {
		return false
	}
	return position == 0 || hooks.modifierAllowPos(modifier)&position != 0
}

func rewardWeaponCountModifiers4F14E0[O, M any](
	table rewardWeaponModifierTable4F14E0,
	slots uint32,
	weaponMask uint32,
	position uint32,
	hooks rewardWeaponHooks4F14E0[O, M],
) int32 {
	if hooks.loadModifierName(table, 0) == "" {
		return 0
	}
	var count int32
	for index := 0; ; index++ {
		if rewardWeaponModifierEligible4F14E0(table, index, slots, weaponMask, position, hooks) {
			count++
		}
		if hooks.loadModifierName(table, index+1) == "" {
			return count
		}
	}
}

func rewardWeaponSelectModifier4F14E0[O, M any](
	table rewardWeaponModifierTable4F14E0,
	slots uint32,
	weaponMask uint32,
	position uint32,
	hooks rewardWeaponHooks4F14E0[O, M],
) (int, bool) {
	count := rewardWeaponCountModifiers4F14E0(table, slots, weaponMask, position, hooks)
	if count == 0 {
		return 0, false
	}
	draw := hooks.randomInt(0, count-1)
	if hooks.loadModifierName(table, 0) == "" {
		return 0, false
	}
	var ordinal int32
	for index := 0; ; index++ {
		if rewardWeaponModifierEligible4F14E0(table, index, slots, weaponMask, position, hooks) {
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

func rewardWeaponModifierCount4F14E0(
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

func rewardWeaponCategoryMask4F14E0(
	count int32,
	randomInt func(int32, int32) int32,
) uint16 {
	switch count {
	case 1:
		roll := randomInt(1, 100)
		if roll <= 20 {
			return rewardWeaponEnchantment1Bit4F14E0
		}
		if roll <= 50 {
			return rewardWeaponPowerBit4F14E0
		}
		return rewardWeaponMaterialBit4F14E0
	case 2:
		roll := randomInt(1, 100)
		if roll <= 12 {
			return rewardWeaponPowerBit4F14E0 | rewardWeaponEnchantment1Bit4F14E0
		}
		if roll <= 25 {
			return rewardWeaponMaterialBit4F14E0 | rewardWeaponEnchantment1Bit4F14E0
		}
		return rewardWeaponPowerBit4F14E0 | rewardWeaponMaterialBit4F14E0
	case 3:
		return rewardWeaponPowerBit4F14E0 | rewardWeaponMaterialBit4F14E0 |
			rewardWeaponEnchantment1Bit4F14E0
	case 4:
		return rewardWeaponPowerBit4F14E0 | rewardWeaponMaterialBit4F14E0 |
			rewardWeaponEnchantment1Bit4F14E0 | rewardWeaponEnchantment2Bit4F14E0
	default:
		return 0
	}
}

func rewardWeaponPromoteMask4F14E0[O, M any](
	mask uint16,
	slots uint32,
	weaponMask uint32,
	hooks rewardWeaponHooks4F14E0[O, M],
) uint16 {
	if mask&rewardWeaponPowerBit4F14E0 != 0 &&
		rewardWeaponCountModifiers4F14E0(
			rewardWeaponPowerTable4F14E0, slots, weaponMask, 0, hooks,
		) == 0 {
		if mask&rewardWeaponMaterialBit4F14E0 != 0 {
			if mask&rewardWeaponEnchantment1Bit4F14E0 != 0 {
				if mask&rewardWeaponEnchantment2Bit4F14E0 == 0 {
					mask |= rewardWeaponEnchantment2Bit4F14E0
				}
			} else {
				mask |= rewardWeaponEnchantment1Bit4F14E0
			}
		} else {
			mask |= rewardWeaponMaterialBit4F14E0
		}
		mask &^= rewardWeaponPowerBit4F14E0
	}
	if mask&rewardWeaponMaterialBit4F14E0 != 0 &&
		rewardWeaponCountModifiers4F14E0(
			rewardWeaponMaterialTable4F14E0, slots, weaponMask, 0, hooks,
		) == 0 {
		if mask&rewardWeaponEnchantment1Bit4F14E0 != 0 {
			if mask&rewardWeaponEnchantment2Bit4F14E0 == 0 {
				mask |= rewardWeaponEnchantment2Bit4F14E0
			}
		} else {
			mask |= rewardWeaponEnchantment1Bit4F14E0
		}
		mask &^= rewardWeaponMaterialBit4F14E0
	}
	if mask&rewardWeaponEnchantment1Bit4F14E0 != 0 &&
		rewardWeaponCountModifiers4F14E0(
			rewardWeaponEnchantmentTable4F14E0, slots, weaponMask, 1, hooks,
		) == 0 {
		mask &^= rewardWeaponEnchantment1Bit4F14E0 | rewardWeaponEnchantment2Bit4F14E0
	}
	return mask
}

func rewardWeaponEnchantmentSlots4F14E0(
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

// rewardWeapon4F14E0 reconstructs GAME.EXE 004F14E0. The marker argument is
// absent because the original body never reads it. Object and modifier tables
// remain live across every RNG call, modifier descriptors retain their native
// pointer identity, and scalar masks keep the original fixed widths.
func rewardWeapon4F14E0[O, M any](stage uint32, hooks rewardWeaponHooks4F14E0[O, M]) O {
	var zero O
	slots := hooks.pickSlots(stage)
	objectIndex, typeInd, selected := rewardWeaponSelectObject4F14E0(stage, slots, hooks)
	if !selected {
		return zero
	}
	weaponMask := hooks.weaponTypeMask(typeInd)
	object := hooks.createObject(typeInd)
	if hooks.isNilObject(object) {
		return zero
	}
	if hooks.loadObjectClass(object)&rewardWeaponWandClass4F14E0 != 0 &&
		hooks.loadObjectSubClass(object)&rewardWeaponWandSubMask4F14E0 != 0 {
		if hooks.loadObjectName(objectIndex)[0] == '#' {
			var modifiers [4]M
			modifiers[2] = hooks.loadReplenishment()
			hooks.applyModifiers(object, modifiers)
		}
		return object
	}

	modifierCount, recognizedSlots := rewardWeaponModifierCount4F14E0(slots, hooks.randomInt)
	if !recognizedSlots || modifierCount == 0 {
		return object
	}
	mask := rewardWeaponCategoryMask4F14E0(modifierCount, hooks.randomInt)
	mask = rewardWeaponPromoteMask4F14E0(mask, slots, weaponMask, hooks)
	if mask == 0 {
		return object
	}

	var modifiers [4]M
	firstEnchantmentIndex := 0
	if mask&rewardWeaponPowerBit4F14E0 != 0 {
		if index, ok := rewardWeaponSelectModifier4F14E0(
			rewardWeaponPowerTable4F14E0, slots, weaponMask, 0, hooks,
		); ok {
			modifiers[0] = hooks.loadModifierValue(rewardWeaponPowerTable4F14E0, index)
		}
	}
	if mask&rewardWeaponMaterialBit4F14E0 != 0 {
		if index, ok := rewardWeaponSelectModifier4F14E0(
			rewardWeaponMaterialTable4F14E0, slots, weaponMask, 0, hooks,
		); ok {
			modifiers[1] = hooks.loadModifierValue(rewardWeaponMaterialTable4F14E0, index)
		}
	}
	if mask&rewardWeaponEnchantment1Bit4F14E0 != 0 {
		firstSlots := rewardWeaponEnchantmentSlots4F14E0(stage, hooks.randomInt)
		if index, ok := rewardWeaponSelectModifier4F14E0(
			rewardWeaponEnchantmentTable4F14E0, firstSlots, weaponMask, 1, hooks,
		); ok {
			firstEnchantmentIndex = index
			modifiers[2] = hooks.loadModifierValue(rewardWeaponEnchantmentTable4F14E0, index)
		}
	}
	if mask&rewardWeaponEnchantment2Bit4F14E0 != 0 {
		secondSlots := rewardWeaponEnchantmentSlots4F14E0(stage, hooks.randomInt)
		if index, ok := rewardWeaponSelectModifier4F14E0(
			rewardWeaponEnchantmentTable4F14E0, secondSlots, weaponMask, 2, hooks,
		); ok && hooks.loadModifierGroup(rewardWeaponEnchantmentTable4F14E0, firstEnchantmentIndex) !=
			hooks.loadModifierGroup(rewardWeaponEnchantmentTable4F14E0, index) {
			modifiers[3] = hooks.loadModifierValue(rewardWeaponEnchantmentTable4F14E0, index)
		}
	}
	hooks.applyModifiers(object, modifiers)
	return object
}
