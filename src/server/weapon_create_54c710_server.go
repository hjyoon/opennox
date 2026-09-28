package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type weaponCreateNativeDeps54C710 struct {
	cache             *weaponCreateTypeCache54C710
	findDefinition    func(uint16) *Modifier
	resolveObjectType func(string) uint32
	modifierID        func(string) int
	modifierDesc      func(int) *ModifierEff
	gameFlag          func(uint32) int32
	loadBalance       func(string) float32
	floatToInt        func(float32) int32
}

func weaponCreateNative54C710(obj *Object, deps weaponCreateNativeDeps54C710) {
	weaponCreate54C710(obj, weaponCreateHooks54C710[
		*Object,
		*Modifier,
		*HealthData,
		*ModifierInitData,
		*ModifierEff,
		unsafe.Pointer,
	]{
		loadTypeInd: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadInitData: func(obj *Object) *ModifierInitData {
			return (*ModifierInitData)(obj.InitData)
		},
		findDefinition: deps.findDefinition,
		loadHeartType: func() uint32 {
			return deps.cache.heart
		},
		storeHeartType: func(value uint32) {
			deps.cache.heart = value
		},
		loadWierdlingType: func() uint32 {
			return deps.cache.wierdling
		},
		storeWierdlingType: func(value uint32) {
			deps.cache.wierdling = value
		},
		storeOrbType: func(value uint32) {
			deps.cache.orb = value
		},
		resolveObjectType: deps.resolveObjectType,
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		loadDurability: func(definition *Modifier) uint16 {
			return uint16(definition.Durability52)
		},
		storeCurrent: func(health *HealthData, value uint16) {
			health.Cur = value
		},
		storeMaximum: func(health *HealthData, value uint16) {
			health.Max = value
		},
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		loadMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
		modifierID:   deps.modifierID,
		modifierDesc: deps.modifierDesc,
		storeModifier: func(data *ModifierInitData, index int, modifier *ModifierEff) {
			data.Modifiers[index] = modifier
		},
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadSubclass: func(obj *Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		loadUseData: func(obj *Object) unsafe.Pointer {
			return obj.UseData.Ptr
		},
		loadUseByte: func(data unsafe.Pointer, offset uintptr) uint8 {
			return *(*uint8)(unsafe.Add(data, offset))
		},
		storeUseByte: func(data unsafe.Pointer, offset uintptr, value uint8) {
			*(*uint8)(unsafe.Add(data, offset)) = value
		},
		gameFlag:    deps.gameFlag,
		loadBalance: deps.loadBalance,
		floatToInt:  deps.floatToInt,
	})
}

// WeaponCreate54C710 initializes weapon durability, fixed Oblivion
// modifiers, ammo, and Quest staff charge through native-width Go pointers.
// The original C address remains registered only as the thing.bin callback
// identity.
//
//go:noinline
func (s *Server) WeaponCreate54C710(obj *Object) {
	weaponCreateNative54C710(obj, weaponCreateNativeDeps54C710{
		cache: &s.weaponCreateTypes54C710,
		findDefinition: func(typeInd uint16) *Modifier {
			return s.Modif.Nox_xxx_getProjectileClassById413250(int(typeInd))
		},
		resolveObjectType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		modifierID: func(name string) int {
			return s.Modif.Nox_xxx_modifGetIdByName413290(name)
		},
		modifierDesc: func(id int) *ModifierEff {
			return s.Modif.Nox_xxx_modifGetDescById413330(id)
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadBalance: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		floatToInt: armorCreateRoundFloat32ToInt32_54C950,
	})
}
