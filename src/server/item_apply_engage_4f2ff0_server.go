package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

type itemApplyEngageNativeDeps4F2FF0 struct {
	call func(unsafe.Pointer, *ModifierEff, *Object, *Object)
}

func itemApplyEngageEffectNative4F2FF0(
	item, owner *Object,
	deps itemApplyEngageNativeDeps4F2FF0,
) {
	itemApplyEngageEffect4F2FF0(item, owner, itemApplyEngageEffectHooks4F2FF0[
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
		loadEngage: func(modifier *ModifierEff) unsafe.Pointer {
			return modifier.Engage112
		},
		callEngage: func(callback unsafe.Pointer, modifier *ModifierEff, owner, item *Object) {
			deps.call(callback, modifier, owner, item)
		},
	})
}

// ItemApplyEngageEffect4F2FF0 binds GAME.EXE 004F2FF0 to native Object,
// ModifierInitData, ModifierEff, and callback pointers. The callback crosses
// the C ABI without narrowing any pointer and receives modifier, owner, item.
//
//go:noinline
func ItemApplyEngageEffect4F2FF0(item, owner *Object) {
	itemApplyEngageEffectNative4F2FF0(item, owner, itemApplyEngageNativeDeps4F2FF0{
		call: func(callback unsafe.Pointer, modifier *ModifierEff, owner, item *Object) {
			ccall.CallVoidPtr3(callback, modifier.C(), owner.CObj(), item.CObj())
		},
	})
}
