package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

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
// ModifierInitData, ModifierEff, and callback pointers. The callback crosses
// the C ABI without narrowing any pointer and receives modifier, owner, item.
//
//go:noinline
func ItemApplyDisengageEffect4F3030(item, owner *Object) {
	itemApplyDisengageEffectNative4F3030(item, owner, itemApplyDisengageNativeDeps4F3030{
		call: func(callback unsafe.Pointer, modifier *ModifierEff, owner, item *Object) {
			ccall.CallVoidPtr3(callback, modifier.C(), owner.CObj(), item.CObj())
		},
	})
}
