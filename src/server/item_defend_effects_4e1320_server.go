package server

import "unsafe"

// ItemDefendEffectsRuntime4E1320 receives the original two-DWORD context,
// including its raw signed damage/type bits. Only context[0] is copied back.
type ItemDefendEffectsRuntime4E1320 struct {
	ApplyDefend func(*ModifierEff, *Object, *Object, *Object, *Object, *[2]int32)
}

func ItemDefendEffects4E1320(target, source, weapon *Object, damage *int32, typ int32, r ItemDefendEffectsRuntime4E1320) int32 {
	return itemDefendEffects4E1320(target, source, weapon, typ, itemDefendEffectsHooks4E1320[*Object, *ModifierInitData, *ModifierEff]{
		lowDWORD:    func(obj *Object) int32 { return int32(uint32(uintptr(unsafe.Pointer(obj)))) },
		firstItem:   func(obj *Object) *Object { return obj.InvFirstItem },
		flags:       func(obj *Object) uint32 { return uint32(obj.ObjFlags) },
		initData:    func(obj *Object) *ModifierInitData { return obj.InitDataModifier() },
		modifier:    func(initData *ModifierInitData, slot int) *ModifierEff { return initData.Modifiers[slot] },
		hasDefend:   func(effect *ModifierEff) bool { return effect.Defend76.Fnc != nil },
		loadDamage:  func() int32 { return *damage },
		storeDamage: func(value int32) { *damage = value },
		applyDefend: r.ApplyDefend,
		nextItem:    func(obj *Object) *Object { return obj.InvNextItem },
	})
}
