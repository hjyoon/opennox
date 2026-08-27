package server

const (
	questItemClassRejected4F2590    = uint32(0x40)
	questItemClassSubclass4F2590    = uint32(0x10)
	questItemClassBook4F2590        = uint32(0x100)
	questItemClassWeapon4F2590      = uint32(0x01000000)
	questItemClassArmor4F2590       = uint32(0x02000000)
	questItemSubclassDirect4F2590   = uint32(0x1ff78)
	questItemDefaultWandMask4F2590  = uint32(0x10000)
	questItemDefaultArmorMask4F2590 = uint32(0x405)
)

var questItemTypeNames4F2590 = [...]string{
	"Diamond",
	"Emerald",
	"Ruby",
	"SulphorousFlareWand",
	"StreetSneakers",
	"StreetShirt",
	"StreetPants",
}

var questItemReplenishmentNames4F2590 = [...]string{
	"Replenishment1",
	"Replenishment2",
	"Replenishment3",
	"Replenishment4",
}

type questItemModifierTable4F2590 uint8

const (
	questItemModifierWeaponPower4F2590 questItemModifierTable4F2590 = iota
	questItemModifierArmorQuality4F2590
	questItemModifierMaterial4F2590
	questItemModifierEnchantments4F2590
)

// questItemEligibilityCache4F2590 is the native-width equivalent of the
// original lazy globals at 005D4594. The Replenishment1 descriptor used by
// default equipment is deliberately separate from the four-descriptor cache
// used by enchantment validation, matching GAME.EXE.
type questItemEligibilityCache4F2590[M comparable] struct {
	typeIDs              [len(questItemTypeNames4F2590)]uint32
	replenishments       [len(questItemReplenishmentNames4F2590)]M
	defaultReplenishment M
}

type questItemEligibilityHooks4F2590[O, D any, M comparable] struct {
	objectTypeID  func(string) uint32
	modifierID    func(string) int32
	modifierDesc  func(int32) M
	isNilModifier func(M) bool

	loadClass        func(O) uint32
	loadSubclass     func(O) uint32
	loadTypeInd      func(O) uint16
	loadUseSpell     func(O) uint8
	loadUseGuideName func(O) string
	loadUseAbility   func(O) uint8
	guideID          func(string) uint32

	loadModifierData func(O) D
	loadModifier     func(D, int) M
	loadModifierName func(M) string
	loadAllowWeapons func(M) uint32
	loadAllowArmor   func(M) uint32
	loadAllowPos     func(M) uint32

	loadWeaponEquipFlags func(O) uint32
	loadArmorEquipFlags  func(O) uint32

	loadRewardObjectName    func(int) string
	loadRewardObjectTypeInd func(int) uint32
	loadRewardSpellID       func(int) uint32
	loadRewardSpellSlots    func(int) uint32
	loadRewardGuideID       func(int) uint32
	loadRewardGuideSlots    func(int) uint32

	loadModifierRowName          func(questItemModifierTable4F2590, int) string
	loadModifierRowModifier      func(questItemModifierTable4F2590, int) M
	loadModifierRowExcludeArmor  func(questItemModifierTable4F2590, int) uint32
	loadModifierRowExcludeWeapon func(questItemModifierTable4F2590, int) uint32
}

func questItemBool4F2590(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func questItemLoadTypeCache4F2590[O, D any, M comparable](
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) {
	if cache.typeIDs[0] != 0 {
		return
	}
	for index, name := range questItemTypeNames4F2590 {
		cache.typeIDs[index] = hooks.objectTypeID(name)
	}
}

func questItemLoadReplenishments4F2960[O, D any, M comparable](
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) {
	if !hooks.isNilModifier(cache.replenishments[0]) {
		return
	}
	for index, name := range questItemReplenishmentNames4F2590 {
		cache.replenishments[index] = hooks.modifierDesc(hooks.modifierID(name))
	}
}

func questItemLoadDefaultReplenishment4F2B60[O, D any, M comparable](
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) {
	if !hooks.isNilModifier(cache.defaultReplenishment) {
		return
	}
	cache.defaultReplenishment = hooks.modifierDesc(hooks.modifierID(questItemReplenishmentNames4F2590[0]))
}

// questBookItemEligible4F2700 reconstructs the class-0x100 helper. Subclass
// bits are checked in the original spell, field-guide, ability order.
func questBookItemEligible4F2700[O, D any, M comparable](
	item O,
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	subclass := hooks.loadSubclass(item)
	if subclass&1 != 0 {
		current := hooks.loadRewardSpellID(0)
		spell := uint32(hooks.loadUseSpell(item))
		if current == 0 {
			return 0
		}
		for index := 0; ; index++ {
			if current == spell && hooks.loadRewardSpellSlots(index) != 0 {
				return 1
			}
			current = hooks.loadRewardSpellID(index + 1)
			if current == 0 {
				return 0
			}
		}
	}
	if subclass&2 != 0 {
		guide := hooks.guideID(hooks.loadUseGuideName(item))
		current := hooks.loadRewardGuideID(0)
		if current == 0 {
			return 0
		}
		for index := 0; ; index++ {
			if current == guide && hooks.loadRewardGuideSlots(index) != 0 {
				return 1
			}
			current = hooks.loadRewardGuideID(index + 1)
			if current == 0 {
				return 0
			}
		}
	}
	if subclass&4 == 0 {
		return 0
	}
	ability := hooks.loadUseAbility(item)
	return questItemBool4F2590(ability != 0 && ability < 6)
}

func questItemFindModifierRow4F2590[O, D any, M comparable](
	table questItemModifierTable4F2590,
	modifier M,
	hooks questItemEligibilityHooks4F2590[O, D, M],
) (int, bool) {
	if hooks.loadModifierRowName(table, 0) == "" {
		return 0, false
	}
	for index := 0; ; index++ {
		if hooks.loadModifierRowModifier(table, index) == modifier {
			return index, true
		}
		if hooks.loadModifierRowName(table, index+1) == "" {
			return 0, false
		}
	}
}

func questItemModifierAllowed4F2590[O, D any, M comparable](
	modifier M,
	table questItemModifierTable4F2590,
	row int,
	weapon bool,
	equipFlags uint32,
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	if weapon {
		if equipFlags&hooks.loadAllowWeapons(modifier) == 0 {
			return 0
		}
		return questItemBool4F2590(equipFlags&hooks.loadModifierRowExcludeWeapon(table, row) == 0)
	}
	if equipFlags&hooks.loadAllowArmor(modifier) == 0 {
		return 0
	}
	return questItemBool4F2590(equipFlags&hooks.loadModifierRowExcludeArmor(table, row) == 0)
}

// questFirstModifierEligible4F27E0 validates slot zero against WeaponPower or
// ArmorQuality. GAME.EXE reloads class after selecting the equipment service;
// both reads remain visible here.
func questFirstModifierEligible4F27E0[O, D any, M comparable](
	item O,
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	data := hooks.loadModifierData(item)
	modifier := hooks.loadModifier(data, 0)
	if hooks.isNilModifier(modifier) {
		return 1
	}
	weaponAtEquip := hooks.loadClass(item)&questItemClassWeapon4F2590 != 0
	var equipFlags uint32
	if weaponAtEquip {
		equipFlags = hooks.loadWeaponEquipFlags(item)
	} else {
		equipFlags = hooks.loadArmorEquipFlags(item)
	}
	weaponAtTable := hooks.loadClass(item)&questItemClassWeapon4F2590 != 0
	table := questItemModifierArmorQuality4F2590
	if weaponAtTable {
		table = questItemModifierWeaponPower4F2590
	}
	row, ok := questItemFindModifierRow4F2590(table, modifier, hooks)
	if !ok {
		return 0
	}
	return questItemModifierAllowed4F2590(modifier, table, row, weaponAtTable, equipFlags, hooks)
}

// questMaterialModifierEligible4F28C0 validates slot one against Material.
func questMaterialModifierEligible4F28C0[O, D any, M comparable](
	item O,
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	data := hooks.loadModifierData(item)
	modifier := hooks.loadModifier(data, 1)
	if hooks.isNilModifier(modifier) {
		return 1
	}
	weaponAtEquip := hooks.loadClass(item)&questItemClassWeapon4F2590 != 0
	var equipFlags uint32
	if weaponAtEquip {
		equipFlags = hooks.loadWeaponEquipFlags(item)
	} else {
		equipFlags = hooks.loadArmorEquipFlags(item)
	}
	row, ok := questItemFindModifierRow4F2590(questItemModifierMaterial4F2590, modifier, hooks)
	if !ok {
		return 0
	}
	weaponAtCheck := hooks.loadClass(item)&questItemClassWeapon4F2590 != 0
	return questItemModifierAllowed4F2590(
		modifier,
		questItemModifierMaterial4F2590,
		row,
		weaponAtCheck,
		equipFlags,
		hooks,
	)
}

func questItemIsReplenishment4F2960[M comparable](modifier M, cache *questItemEligibilityCache4F2590[M]) bool {
	for _, replenishment := range cache.replenishments {
		if modifier == replenishment {
			return true
		}
	}
	return false
}

func questItemRewardHasHashType4F2960[O, D any, M comparable](
	item O,
	firstName string,
	hooks questItemEligibilityHooks4F2590[O, D, M],
) bool {
	if firstName == "" {
		return false
	}
	found := false
	for index, name := 0, firstName; name != ""; index, name = index+1, hooks.loadRewardObjectName(index+1) {
		if name[0] == '#' {
			typeInd := uint32(hooks.loadTypeInd(item))
			if hooks.loadRewardObjectTypeInd(index) == typeInd {
				found = true
			}
		}
	}
	return found
}

// questEnchantmentModifiersEligible4F2960 validates slots two and three.
// Replenishment1..4 use the hash-prefixed reward-object exception regardless
// of whether the descriptor also appears in the Enchantments table.
func questEnchantmentModifiersEligible4F2960[O, D any, M comparable](
	item O,
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	data := hooks.loadModifierData(item)
	questItemLoadReplenishments4F2960(cache, hooks)
	weaponAtEquip := hooks.loadClass(item)&questItemClassWeapon4F2590 != 0
	var equipFlags uint32
	if weaponAtEquip {
		equipFlags = hooks.loadWeaponEquipFlags(item)
	} else {
		equipFlags = hooks.loadArmorEquipFlags(item)
	}
	firstRewardName := hooks.loadRewardObjectName(0)

	for slot := 2; slot < 4; slot++ {
		modifier := hooks.loadModifier(data, slot)
		if hooks.isNilModifier(modifier) {
			continue
		}
		position := uint32(1 << (slot - 2))
		if hooks.loadAllowPos(modifier)&position == 0 {
			return 0
		}
		row, inTable := questItemFindModifierRow4F2590(questItemModifierEnchantments4F2590, modifier, hooks)
		if questItemIsReplenishment4F2960(modifier, cache) {
			if !questItemRewardHasHashType4F2960(item, firstRewardName, hooks) {
				return 0
			}
			continue
		}
		if !inTable {
			return 0
		}
		weaponAtCheck := hooks.loadClass(item)&questItemClassWeapon4F2590 != 0
		if questItemModifierAllowed4F2590(
			modifier,
			questItemModifierEnchantments4F2590,
			row,
			weaponAtCheck,
			equipFlags,
			hooks,
		) == 0 {
			return 0
		}
	}
	return 1
}

func questEquipmentModifiersEligible4F27A0[O, D any, M comparable](
	item O,
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	if questFirstModifierEligible4F27E0(item, hooks) == 0 {
		return 0
	}
	if questMaterialModifierEligible4F28C0(item, hooks) == 0 {
		return 0
	}
	return questItemBool4F2590(questEnchantmentModifiersEligible4F2960(item, cache, hooks) != 0)
}

func questItemASCIIEqualFoldPrefix4F2B60(value, prefix string) bool {
	if len(value) < len(prefix) {
		return false
	}
	for index := range prefix {
		left, right := value[index], prefix[index]
		if left >= 'A' && left <= 'Z' {
			left += 'a' - 'A'
		}
		if right >= 'A' && right <= 'Z' {
			right += 'a' - 'A'
		}
		if left != right {
			return false
		}
	}
	return true
}

// questDefaultItemModifiersEligible4F2B60 preserves the two default-item
// exceptions: the flare wand has exactly Replenishment1 in slot two, while
// selected Street clothing accepts only nil or UserColo-prefixed modifiers.
func questDefaultItemModifiersEligible4F2B60[O, D any, M comparable](
	item O,
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	questItemLoadDefaultReplenishment4F2B60(cache, hooks)
	if hooks.loadClass(item)&questItemClassWeapon4F2590 != 0 &&
		hooks.loadWeaponEquipFlags(item)&questItemDefaultWandMask4F2590 != 0 {
		data := hooks.loadModifierData(item)
		if !hooks.isNilModifier(hooks.loadModifier(data, 0)) ||
			!hooks.isNilModifier(hooks.loadModifier(data, 1)) ||
			hooks.loadModifier(data, 2) != cache.defaultReplenishment ||
			!hooks.isNilModifier(hooks.loadModifier(data, 3)) {
			return 0
		}
	}
	if hooks.loadClass(item)&questItemClassArmor4F2590 != 0 {
		equipFlags := hooks.loadArmorEquipFlags(item)
		data := hooks.loadModifierData(item)
		if equipFlags&questItemDefaultArmorMask4F2590 != 0 {
			for slot := 0; slot < 4; slot++ {
				modifier := hooks.loadModifier(data, slot)
				if !hooks.isNilModifier(modifier) &&
					!questItemASCIIEqualFoldPrefix4F2B60(hooks.loadModifierName(modifier), "UserColo") {
					return 0
				}
			}
		}
	}
	return 1
}

// questItemEligible4F2590 reconstructs GAME.EXE 004F2590 together with its
// inseparable helpers through 004F2B60. Every table is read live and every
// pointer-shaped value remains native-width through the generic contract.
func questItemEligible4F2590[O, D any, M comparable](
	item O,
	cache *questItemEligibilityCache4F2590[M],
	hooks questItemEligibilityHooks4F2590[O, D, M],
) int32 {
	questItemLoadTypeCache4F2590(cache, hooks)
	class := hooks.loadClass(item)
	if class&questItemClassRejected4F2590 != 0 {
		return 0
	}
	if class&questItemClassSubclass4F2590 != 0 {
		return questItemBool4F2590(hooks.loadSubclass(item)&questItemSubclassDirect4F2590 != 0)
	}
	if class&questItemClassBook4F2590 != 0 {
		return questBookItemEligible4F2700(item, hooks)
	}

	typeInd := uint32(hooks.loadTypeInd(item))
	if typeInd == cache.typeIDs[0] || typeInd == cache.typeIDs[1] || typeInd == cache.typeIDs[2] {
		return 1
	}

	rewardObject := false
	if hooks.loadRewardObjectName(0) != "" {
		for index := 0; ; index++ {
			if hooks.loadRewardObjectTypeInd(index) == typeInd {
				rewardObject = true
				break
			}
			if hooks.loadRewardObjectName(index+1) == "" {
				break
			}
		}
	}
	if typeInd == cache.typeIDs[3] || typeInd == cache.typeIDs[4] ||
		typeInd == cache.typeIDs[5] || typeInd == cache.typeIDs[6] {
		return questDefaultItemModifiersEligible4F2B60(item, cache, hooks)
	}
	if !rewardObject {
		return 0
	}
	if class&questItemClassWeapon4F2590 != 0 {
		return questEquipmentModifiersEligible4F27A0(item, cache, hooks)
	}
	if class&questItemClassArmor4F2590 != 0 {
		return questEquipmentModifiersEligible4F27A0(item, cache, hooks)
	}
	return 1
}
