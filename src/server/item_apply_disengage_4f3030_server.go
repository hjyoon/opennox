package server

import "unsafe"

type itemApplyDisengageNativeDeps4F3030 struct {
	call func(unsafe.Pointer, *ModifierEff, *Object, *Object)
}

func itemApplyDisengageEffectNative4F3030(
	item, owner *Object,
	deps itemApplyDisengageNativeDeps4F3030,
) {
	itemApplyDisengageEffect4F3030(item, owner, itemApplyDisengageEffectHooks4F3030[
		*Object,
		*ModifierInitData,
		*ModifierEff,
		unsafe.Pointer,
	]{
		loadInitData: func(item *Object) *ModifierInitData {
			return (*ModifierInitData)(item.InitData)
		},
		loadModifier: func(data *ModifierInitData, slot int) *ModifierEff {
			return data.Modifiers[slot]
		},
		loadDisengage: func(modifier *ModifierEff) unsafe.Pointer {
			return modifier.Disengage116
		},
		callDisengage: func(callback unsafe.Pointer, modifier *ModifierEff, owner, item *Object) {
			deps.call(callback, modifier, owner, item)
		},
	})
}

// ItemApplyDisengageEffect4F3030 binds GAME.EXE 004F3030 to native Object,
// ModifierInitData, ModifierEff, and callback pointers. Each callback receives
// the modifier, owner, and item at native width.
//
//go:noinline
func ItemApplyDisengageEffect4F3030(item, owner *Object) {
	itemApplyDisengageEffectNative4F3030(item, owner, itemApplyDisengageNativeDeps4F3030{
		call: callModifierEngageEffect,
	})
}
