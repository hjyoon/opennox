package server

func questItemGuideID4F2590(name string) uint32 {
	for guideID, guideName := range rewardFieldGuideNames4F0D20 {
		if guideName == name {
			return uint32(guideID)
		}
	}
	return 0
}

func (s *Server) questItemModifierRows4F2590(table questItemModifierTable4F2590) []rewardModifierDefinition4F0640 {
	switch table {
	case questItemModifierWeaponPower4F2590:
		return s.rewardDefinitions.WeaponPower[:]
	case questItemModifierArmorQuality4F2590:
		return s.rewardDefinitions.ArmorQuality[:]
	case questItemModifierMaterial4F2590:
		return s.rewardDefinitions.Material[:]
	case questItemModifierEnchantments4F2590:
		return s.rewardDefinitions.Enchantments[:]
	default:
		panic("invalid Quest modifier table")
	}
}

// QuestItemEligible4F2590 binds GAME.EXE 004F2590..004F2B60 to native
// Server, Object, ModifierInitData, and ModifierEff values. The original
// reward tables remain live, and no object or modifier pointer is narrowed to
// a PE32 integer.
//
//go:noinline
func (s *Server) QuestItemEligible4F2590(item *Object) int32 {
	modifierRows := func(table questItemModifierTable4F2590) []rewardModifierDefinition4F0640 {
		return s.questItemModifierRows4F2590(table)
	}
	return questItemEligible4F2590(item, &s.questItemEligibility, questItemEligibilityHooks4F2590[
		*Object,
		*ModifierInitData,
		*ModifierEff,
	]{
		objectTypeID: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		modifierID: func(name string) int32 {
			return int32(s.Modif.Nox_xxx_modifGetIdByName413290(name))
		},
		modifierDesc: func(id int32) *ModifierEff {
			return s.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
		isNilModifier: func(modifier *ModifierEff) bool {
			return modifier == nil
		},
		loadClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		loadSubclass: func(item *Object) uint32 {
			return uint32(item.ObjSubClass)
		},
		loadTypeInd: func(item *Object) uint16 {
			return item.TypeInd
		},
		loadUseSpell: func(item *Object) uint8 {
			return item.UseDataSpellReward().Spell
		},
		loadUseGuideName: func(item *Object) string {
			return item.UseDataFieldGuide().Creature()
		},
		loadUseAbility: func(item *Object) uint8 {
			return item.UseDataAbilityReward().Ability
		},
		guideID: questItemGuideID4F2590,
		loadModifierData: func(item *Object) *ModifierInitData {
			return item.InitDataModifier()
		},
		loadModifier: func(data *ModifierInitData, slot int) *ModifierEff {
			return data.Modifiers[slot]
		},
		loadModifierName: func(modifier *ModifierEff) string {
			return modifier.Name()
		},
		loadAllowWeapons: func(modifier *ModifierEff) uint32 {
			return modifier.AllowWeapons28
		},
		loadAllowArmor: func(modifier *ModifierEff) uint32 {
			return modifier.AllowArmor32
		},
		loadAllowPos: func(modifier *ModifierEff) uint32 {
			return modifier.AllowPos36
		},
		loadWeaponEquipFlags: s.Weapons.Nox_xxx_weaponInventoryEquipFlags_415820,
		loadArmorEquipFlags:  s.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70,
		loadRewardObjectName: func(index int) string {
			return s.rewardDefinitions.Objects[index].Name
		},
		loadRewardObjectTypeInd: func(index int) uint32 {
			return s.rewardDefinitions.Objects[index].TypeInd
		},
		loadRewardSpellID: func(index int) uint32 {
			return rewardSpellDefinitions4F09F0[index].SpellID
		},
		loadRewardSpellSlots: func(index int) uint32 {
			return rewardSpellDefinitions4F09F0[index].Slots
		},
		loadRewardGuideID: func(index int) uint32 {
			return rewardFieldGuideDefinitions4F0D20[index].GuideID
		},
		loadRewardGuideSlots: func(index int) uint32 {
			return rewardFieldGuideDefinitions4F0D20[index].Slots
		},
		loadModifierRowName: func(table questItemModifierTable4F2590, index int) string {
			return modifierRows(table)[index].Name
		},
		loadModifierRowModifier: func(table questItemModifierTable4F2590, index int) *ModifierEff {
			return modifierRows(table)[index].Modifier
		},
		loadModifierRowExcludeArmor: func(table questItemModifierTable4F2590, index int) uint32 {
			return modifierRows(table)[index].ExcludeArmor
		},
		loadModifierRowExcludeWeapon: func(table questItemModifierTable4F2590, index int) uint32 {
			return modifierRows(table)[index].ExcludeWeapon
		},
	})
}
