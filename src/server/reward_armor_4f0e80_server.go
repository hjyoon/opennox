package server

import "github.com/opennox/libs/object"

type rewardArmorNativeTables4F0E80 struct {
	objects      []rewardObjectDefinition4F0640
	armorQuality []rewardModifierDefinition4F0640
	material     []rewardModifierDefinition4F0640
	enchantments []rewardModifierDefinition4F0640
}

type rewardArmorNativeDeps4F0E80 struct {
	tables            rewardArmorNativeTables4F0E80
	pickSlots         func(uint32) uint32
	randomInt         func(int32, int32) int32
	objectTypeAllowed func(uint32) bool
	armorTypeMask     func(uint32) uint32
	createObject      func(uint32) *Object
	applyModifiers    func(*Object, *ModifierInitData)
}

func (deps rewardArmorNativeDeps4F0E80) modifierTable(
	table rewardArmorModifierTable4F0E80,
) []rewardModifierDefinition4F0640 {
	switch table {
	case rewardArmorQualityTable4F0E80:
		return deps.tables.armorQuality
	case rewardArmorMaterialTable4F0E80:
		return deps.tables.material
	case rewardArmorEnchantmentTable4F0E80:
		return deps.tables.enchantments
	default:
		panic("invalid reward armor modifier table")
	}
}

func rewardArmorNative4F0E80(stage uint32, deps rewardArmorNativeDeps4F0E80) *Object {
	return rewardArmor4F0E80(stage, rewardArmorHooks4F0E80[*Object, *ModifierEff]{
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
		armorTypeMask:     deps.armorTypeMask,
		createObject:      deps.createObject,
		isNilObject: func(object *Object) bool {
			return object == nil
		},
		loadModifierName: func(table rewardArmorModifierTable4F0E80, index int) string {
			return deps.modifierTable(table)[index].Name
		},
		loadModifierValue: func(table rewardArmorModifierTable4F0E80, index int) *ModifierEff {
			return deps.modifierTable(table)[index].Modifier
		},
		loadModifierGroup: func(table rewardArmorModifierTable4F0E80, index int) uint8 {
			return uint8(deps.modifierTable(table)[index].Group)
		},
		loadModifierSlots: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return deps.modifierTable(table)[index].Slots
		},
		loadModifierExcludeArmor: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return deps.modifierTable(table)[index].ExcludeArmor
		},
		loadModifierAllowArmor: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return deps.modifierTable(table)[index].Modifier.AllowArmor32
		},
		loadModifierAllowPos: func(table rewardArmorModifierTable4F0E80, index int) uint32 {
			return deps.modifierTable(table)[index].Modifier.AllowPos36
		},
		applyModifiers: func(object *Object, modifiers [4]*ModifierEff) {
			deps.applyModifiers(object, &ModifierInitData{Modifiers: modifiers})
		},
	})
}

// applyModifierAttrs4E4990 keeps the observable part of the original wrapper
// needed by native callers: empty ordinary attributes return before TeamBase
// lookup, a zero type-ID lookup is retried, and successful lookups are cached.
// The original mixed pointer/integer C return value has no native consumer.
func (s *Server) applyModifierAttrs4E4990(obj *Object, attrs *ModifierInitData) bool {
	forced := object.Class(obj.ObjClass).Has(object.ClassWand) && uint32(obj.ObjSubClass)&0x047F0000 != 0
	if !forced && !attrs.HasModifiers() {
		return false
	}
	if s.Modif.teamBaseTypeInd4E4990 == 0 {
		s.Modif.teamBaseTypeInd4E4990 = uint32(s.Types.IndByID("TeamBase"))
	}
	return obj.SetModifierAttrs(attrs, s.Modif.teamBaseTypeInd4E4990)
}

// RewardArmor4F0E80 binds GAME.EXE 004F0E80 to the per-server reward tables,
// object registry, armor registry, logic RNG, native object factory, and
// native-width modifier pointers. The marker is intentionally ignored because
// the original function never reads its first argument.
//
//go:noinline
func (s *Server) RewardArmor4F0E80(_ *Object, stage uint32) *Object {
	randomInt := func(minimum, maximum int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	}
	return rewardArmorNative4F0E80(stage, rewardArmorNativeDeps4F0E80{
		tables: rewardArmorNativeTables4F0E80{
			objects:      s.rewardDefinitions.Objects[:],
			armorQuality: s.rewardDefinitions.ArmorQuality[:],
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
		armorTypeMask: func(typeInd uint32) uint32 {
			return s.Armor.Sub_415D10(int(typeInd))
		},
		createObject: func(typeInd uint32) *Object {
			return s.NewObjectByTypeInd(int(typeInd))
		},
		applyModifiers: func(object *Object, attrs *ModifierInitData) {
			s.applyModifierAttrs4E4990(object, attrs)
		},
	})
}
