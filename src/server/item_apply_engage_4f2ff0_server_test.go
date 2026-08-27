package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestItemApplyEngageEffect4F2FF0NativeLayouts(t *testing.T) {
	wantObjectInitData := uintptr(692)
	wantInitSize := uintptr(20)
	wantThirdModifier := uintptr(8)
	wantModifierSize := uintptr(144)
	wantEngage := uintptr(112)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectInitData = 760
		wantInitSize = 40
		wantThirdModifier = 16
		wantModifierSize = 208
		wantEngage = 160
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantObjectInitData},
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantInitSize},
		{"ModifierInitData.Modifiers", unsafe.Offsetof(ModifierInitData{}.Modifiers), 0},
		{"ModifierInitData.Modifiers[2]", 2 * unsafe.Sizeof((*ModifierEff)(nil)), wantThirdModifier},
		{"ModifierEff size", unsafe.Sizeof(ModifierEff{}), wantModifierSize},
		{"ModifierEff.Engage112", unsafe.Offsetof(ModifierEff{}.Engage112), wantEngage},
		{"engage callback width", unsafe.Sizeof(ModifierEff{}.Engage112), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestItemApplyEngageEffect4F2FF0NativePointerIdentityAndLiveLoads(t *testing.T) {
	firstCallback := unsafe.Pointer(new(byte))
	oldThirdCallback := unsafe.Pointer(new(byte))
	newThirdCallback := unsafe.Pointer(new(byte))
	first := &ModifierEff{Engage112: firstCallback}
	oldThird := &ModifierEff{Engage112: oldThirdCallback}
	newThird := &ModifierEff{Engage112: newThirdCallback}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, first, oldThird}}
	replacementData := &ModifierInitData{}
	item := &Object{InitData: unsafe.Pointer(data)}
	owner := &Object{}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{
			unsafe.Pointer(item),
			unsafe.Pointer(owner),
			unsafe.Pointer(data),
			unsafe.Pointer(first),
			unsafe.Pointer(oldThird),
			unsafe.Pointer(newThird),
			firstCallback,
			oldThirdCallback,
			newThirdCallback,
		}
		for index, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want a native high address", index, pointer)
			}
		}
	}

	type call struct {
		callback unsafe.Pointer
		modifier *ModifierEff
		owner    *Object
		item     *Object
	}
	var calls []call
	itemApplyEngageEffectNative4F2FF0(item, owner, itemApplyEngageNativeDeps4F2FF0{
		call: func(callback unsafe.Pointer, modifier *ModifierEff, gotOwner, gotItem *Object) {
			calls = append(calls, call{callback, modifier, gotOwner, gotItem})
			if modifier == first {
				item.InitData = unsafe.Pointer(replacementData)
				data.Modifiers[3] = newThird
			}
		},
	})
	want := []call{
		{firstCallback, first, owner, item},
		{newThirdCallback, newThird, owner, item},
	}
	if len(calls) != len(want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
	for index := range want {
		if calls[index] != want[index] {
			t.Fatalf("call %d = %#v, want %#v", index, calls[index], want[index])
		}
	}
	runtime.KeepAlive(oldThird)
	runtime.KeepAlive(replacementData)
}

func TestItemApplyEngageEffect4F2FF0NativeForwardsNilOwner(t *testing.T) {
	callback := unsafe.Pointer(new(byte))
	modifier := &ModifierEff{Engage112: callback}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier}}
	item := &Object{InitData: unsafe.Pointer(data)}
	calls := 0
	itemApplyEngageEffectNative4F2FF0(item, nil, itemApplyEngageNativeDeps4F2FF0{
		call: func(gotCallback unsafe.Pointer, gotModifier *ModifierEff, gotOwner, gotItem *Object) {
			calls++
			if gotCallback != callback || gotModifier != modifier || gotOwner != nil || gotItem != item {
				t.Fatalf("callback args = (%p,%p,%p,%p), want (%p,%p,nil,%p)", gotCallback, gotModifier, gotOwner, gotItem, callback, modifier, item)
			}
		},
	})
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestItemApplyEngageEffect4F2FF0NativePreservesUnguardedFaults(t *testing.T) {
	deps := itemApplyEngageNativeDeps4F2FF0{
		call: func(unsafe.Pointer, *ModifierEff, *Object, *Object) {
			t.Fatal("callback reached")
		},
	}
	t.Run("nil item", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil item did not preserve the original fault")
			}
		}()
		itemApplyEngageEffectNative4F2FF0(nil, new(Object), deps)
	})
	t.Run("nil InitData", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil InitData did not preserve the original fault")
			}
		}()
		itemApplyEngageEffectNative4F2FF0(new(Object), new(Object), deps)
	})
}
