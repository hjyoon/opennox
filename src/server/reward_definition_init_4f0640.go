package server

// rewardObjectDefinition4F0640 is the native-width replacement for one
// original 20-byte reward-object row. All numeric game values remain exactly
// 32-bit; only the name is represented as a Go string instead of a PE32 VA.
type rewardObjectDefinition4F0640 struct {
	Weight  uint32
	Name    string
	TypeInd uint32
	Kind    uint32
	Slots   uint32
}

// rewardModifierDefinition4F0640 is the native-width replacement for one
// original 24-byte reward-modifier row. Modifier is deliberately a native
// pointer and is never stored in an emulated uint32 cell.
type rewardModifierDefinition4F0640 struct {
	Group         uint32
	Modifier      *ModifierEff
	Name          string
	Slots         uint32
	ExcludeArmor  uint32
	ExcludeWeapon uint32
}

type rewardDefinitionTables4F0640 struct {
	Objects      [58]rewardObjectDefinition4F0640
	WeaponPower  [6]rewardModifierDefinition4F0640
	ArmorQuality [6]rewardModifierDefinition4F0640
	Material     [6]rewardModifierDefinition4F0640
	Enchantments [57]rewardModifierDefinition4F0640
	initialized  bool
}

func (tables *rewardDefinitionTables4F0640) init() {
	tables.Objects = rewardObjectDefinitions4F0640
	tables.WeaponPower = rewardWeaponPowerDefinitions4F0640
	tables.ArmorQuality = rewardArmorQualityDefinitions4F0640
	tables.Material = rewardMaterialDefinitions4F0640
	tables.Enchantments = rewardEnchantmentDefinitions4F0640
	tables.initialized = true
}

func resolveRewardModifierGroup4F0640(
	definitions []rewardModifierDefinition4F0640,
	modifierID func(string) int32,
	modifierDesc func(int32) *ModifierEff,
) {
	for index := range definitions {
		name := definitions[index].Name
		if name == "" {
			break
		}
		id := modifierID(name)
		definitions[index].Modifier = modifierDesc(id)
	}
}

// rewardDefinitionsInit4F0640 preserves the exact externally observable order
// of GAME.EXE 004F0640: 57 object lookups, then the weapon-power,
// armor-quality, material, and enchantment modifier groups. Each next name is
// read only after the current lookup(s) and store have completed.
func rewardDefinitionsInit4F0640(
	tables *rewardDefinitionTables4F0640,
	objectTypeID func(string) int32,
	modifierID func(string) int32,
	modifierDesc func(int32) *ModifierEff,
) {
	for index := range tables.Objects {
		name := tables.Objects[index].Name
		if name == "" {
			break
		}
		if name[0] == '#' {
			name = name[1:]
		}
		tables.Objects[index].TypeInd = uint32(objectTypeID(name))
	}
	resolveRewardModifierGroup4F0640(tables.WeaponPower[:], modifierID, modifierDesc)
	resolveRewardModifierGroup4F0640(tables.ArmorQuality[:], modifierID, modifierDesc)
	resolveRewardModifierGroup4F0640(tables.Material[:], modifierID, modifierDesc)
	resolveRewardModifierGroup4F0640(tables.Enchantments[:], modifierID, modifierDesc)
}

// RewardDefinitionsInit4F0640 replaces the sole active call to the original
// ABI32 initializer. Static row metadata stays fixed-width while resolved
// ModifierEff references retain the native pointer width of the target.
func (s *Server) RewardDefinitionsInit4F0640() {
	if !s.rewardDefinitions.initialized {
		s.rewardDefinitions.init()
	}
	rewardDefinitionsInit4F0640(
		&s.rewardDefinitions,
		func(name string) int32 {
			return int32(s.Types.IndByID(name))
		},
		func(name string) int32 {
			return int32(s.Modif.Nox_xxx_modifGetIdByName413290(name))
		},
		func(id int32) *ModifierEff {
			return s.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
	)
}
