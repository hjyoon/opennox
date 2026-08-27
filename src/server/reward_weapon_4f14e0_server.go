package server

type rewardWeaponNativeTables4F14E0 struct {
	objects      []rewardObjectDefinition4F0640
	weaponPower  []rewardModifierDefinition4F0640
	material     []rewardModifierDefinition4F0640
	enchantments []rewardModifierDefinition4F0640
}

type rewardWeaponNativeDeps4F14E0 struct {
	tables            rewardWeaponNativeTables4F14E0
	pickSlots         func(uint32) uint32
	randomInt         func(int32, int32) int32
	objectTypeAllowed func(uint32) bool
	weaponTypeMask    func(uint32) uint32
	createObject      func(uint32) *Object
	loadReplenishment func() *ModifierEff
	applyModifiers    func(*Object, *ModifierInitData)
}

func (deps rewardWeaponNativeDeps4F14E0) modifierTable(
	table rewardWeaponModifierTable4F14E0,
) []rewardModifierDefinition4F0640 {
	switch table {
	case rewardWeaponPowerTable4F14E0:
		return deps.tables.weaponPower
	case rewardWeaponMaterialTable4F14E0:
		return deps.tables.material
	case rewardWeaponEnchantmentTable4F14E0:
		return deps.tables.enchantments
	default:
		panic("invalid reward weapon modifier table")
	}
}

func rewardWeaponNative4F14E0(stage uint32, deps rewardWeaponNativeDeps4F14E0) *Object {
	return rewardWeapon4F14E0(stage, rewardWeaponHooks4F14E0[*Object, *ModifierEff]{
		pickSlots: deps.pickSlots,
		randomInt: deps.randomInt,
		loadObjectName: func(index int) string {
			return deps.tables.objects[index].Name
		},
		loadObjectWeight: func(index int) uint8 {
			return uint8(deps.tables.objects[index].Weight)
		},
		loadObjectTypeInd: func(index int) uint32 {
			return deps.tables.objects[index].TypeInd
		},
		loadObjectKind: func(index int) uint32 {
			return deps.tables.objects[index].Kind
		},
		loadObjectSlots: func(index int) uint32 {
			return deps.tables.objects[index].Slots
		},
		objectTypeAllowed: deps.objectTypeAllowed,
		weaponTypeMask:    deps.weaponTypeMask,
		createObject:      deps.createObject,
		isNilObject: func(object *Object) bool {
			return object == nil
		},
		loadObjectClass: func(object *Object) uint32 {
			return uint32(object.ObjClass)
		},
		loadObjectSubClass: func(object *Object) uint32 {
			return uint32(object.ObjSubClass)
		},
		loadModifierName: func(table rewardWeaponModifierTable4F14E0, index int) string {
			return deps.modifierTable(table)[index].Name
		},
		loadModifierValue: func(table rewardWeaponModifierTable4F14E0, index int) *ModifierEff {
			return deps.modifierTable(table)[index].Modifier
		},
		loadModifierGroup: func(table rewardWeaponModifierTable4F14E0, index int) uint8 {
			return uint8(deps.modifierTable(table)[index].Group)
		},
		loadModifierSlots: func(table rewardWeaponModifierTable4F14E0, index int) uint32 {
			return deps.modifierTable(table)[index].Slots
		},
		loadModifierExcludeWeapon: func(table rewardWeaponModifierTable4F14E0, index int) uint32 {
			return deps.modifierTable(table)[index].ExcludeWeapon
		},
		modifierAllowWeapons: func(modifier *ModifierEff) uint32 {
			return modifier.AllowWeapons28
		},
		modifierAllowPos: func(modifier *ModifierEff) uint32 {
			return modifier.AllowPos36
		},
		loadReplenishment: deps.loadReplenishment,
		applyModifiers: func(object *Object, modifiers [4]*ModifierEff) {
			deps.applyModifiers(object, &ModifierInitData{Modifiers: modifiers})
		},
	})
}

// RewardWeapon4F14E0 binds GAME.EXE 004F14E0 to native object and modifier
// pointers, the per-server reward definitions, weapon registry, object
// registry, logic RNG, and object factory. The marker is intentionally ignored
// because the original function never reads its first argument.
//
//go:noinline
func (s *Server) RewardWeapon4F14E0(_ *Object, stage uint32) *Object {
	randomInt := func(minimum, maximum int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	}
	return rewardWeaponNative4F14E0(stage, rewardWeaponNativeDeps4F14E0{
		tables: rewardWeaponNativeTables4F14E0{
			objects:      s.rewardDefinitions.Objects[:],
			weaponPower:  s.rewardDefinitions.WeaponPower[:],
			material:     s.rewardDefinitions.Material[:],
			enchantments: s.rewardDefinitions.Enchantments[:],
		},
		pickSlots: func(stage uint32) uint32 {
			return rewardRandomSlots4F0B60(stage, randomInt)
		},
		randomInt: randomInt,
		objectTypeAllowed: func(typeInd uint32) bool {
			return s.Types.ByInd(int(typeInd)).Allowed()
		},
		weaponTypeMask: func(typeInd uint32) uint32 {
			return s.Weapons.Nox_xxx_ammoCheck_415880(int(typeInd))
		},
		createObject: func(typeInd uint32) *Object {
			return s.NewObjectByTypeInd(int(typeInd))
		},
		loadReplenishment: func() *ModifierEff {
			id := s.Modif.Nox_xxx_modifGetIdByName413290("Replenishment1")
			return s.Modif.Nox_xxx_modifGetDescById413330(id)
		},
		applyModifiers: func(object *Object, attrs *ModifierInitData) {
			s.applyModifierAttrs4E4990(object, attrs)
		},
	})
}
