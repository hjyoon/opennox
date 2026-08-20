package server

import "unsafe"

// PoisonProtectionRuntime4E0040 carries the identity of the original
// PoisonProtectEngage callback. GAME.EXE compares this identity but does not
// call it while calculating protection.
type PoisonProtectionRuntime4E0040 struct {
	PoisonProtectEngage unsafe.Pointer
}

type poisonProtectionNativeDeps4E0040 struct {
	poisonProtectEngage unsafe.Pointer
	loadBalance         func(string, int32) float64
}

func poisonProtectionNative4E0040(
	unit *Object,
	deps poisonProtectionNativeDeps4E0040,
) float64 {
	return poisonProtection4E0040(poisonProtectionHooks4E0040[
		*Object,
		*ModifierInitData,
		*ModifierEff,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadFirstItem: func(unit *Object) *Object {
			return unit.InvFirstItem
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadInitData: func(obj *Object) *ModifierInitData {
			return (*ModifierInitData)(obj.InitData)
		},
		loadModifier: func(init *ModifierInitData, slot int) *ModifierEff {
			return init.Modifiers[slot]
		},
		matchesProtection: func(modifier *ModifierEff) bool {
			return deps.poisonProtectEngage != nil && modifier.Engage112 == deps.poisonProtectEngage
		},
		loadModifierValue: func(modifier *ModifierEff) float32 {
			return modifier.EngageFloat120
		},
		loadNextItem: func(item *Object) *Object {
			return item.InvNextItem
		},
		testBuff: func(unit *Object, enchant uint32) int32 {
			if unit.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		loadBuffPower: func(unit *Object, enchant uint32) uint32 {
			return uint32(unit.BuffsPower[EnchantID(enchant)])
		},
		loadBalance: deps.loadBalance,
	})
}

// PoisonProtection4E0040 binds the original modifier walk to native-width
// Object, ModifierInitData, and ModifierEff layouts.
func (s *Server) PoisonProtection4E0040(
	unit *Object,
	runtime PoisonProtectionRuntime4E0040,
) float64 {
	return poisonProtectionNative4E0040(unit, poisonProtectionNativeDeps4E0040{
		poisonProtectEngage: runtime.PoisonProtectEngage,
		loadBalance: func(key string, index int32) float64 {
			return s.Balance.FloatInd(key, int(index))
		},
	})
}
