package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestItemApplyDisengageEffect4F3030NativeLayouts(t *testing.T) {
	wantObjectInitData := uintptr(692)
	wantInitSize := uintptr(20)
	wantThirdModifier := uintptr(8)
	wantModifierSize := uintptr(144)
	wantDisengage := uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectInitData = 760
		wantInitSize = 40
		wantThirdModifier = 16
		wantModifierSize = 208
		wantDisengage = 168
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
		{"ModifierEff.Disengage116", unsafe.Offsetof(ModifierEff{}.Disengage116), wantDisengage},
		{"disengage callback width", unsafe.Sizeof(ModifierEff{}.Disengage116), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestItemApplyDisengageEffect4F3030NativePointerIdentityAndLiveLoads(t *testing.T) {
	firstCallback := unsafe.Pointer(new(byte))
	oldThirdCallback := unsafe.Pointer(new(byte))
	newThirdCallback := unsafe.Pointer(new(byte))
	first := &ModifierEff{Disengage116: firstCallback}
	oldThird := &ModifierEff{Disengage116: oldThirdCallback}
	newThird := &ModifierEff{Disengage116: newThirdCallback}
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
	itemApplyDisengageEffectNative4F3030(item, owner, itemApplyDisengageNativeDeps4F3030{
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

func TestItemApplyDisengageEffect4F3030NativeForwardsNilOwner(t *testing.T) {
	callback := unsafe.Pointer(new(byte))
	modifier := &ModifierEff{Disengage116: callback}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier}}
	item := &Object{InitData: unsafe.Pointer(data)}
	calls := 0
	itemApplyDisengageEffectNative4F3030(item, nil, itemApplyDisengageNativeDeps4F3030{
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

func TestItemApplyDisengageEffect4F3030NativePreservesUnguardedFaults(t *testing.T) {
	deps := itemApplyDisengageNativeDeps4F3030{
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
		itemApplyDisengageEffectNative4F3030(nil, new(Object), deps)
	})
	t.Run("nil InitData", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil InitData did not preserve the original fault")
			}
		}()
		itemApplyDisengageEffectNative4F3030(new(Object), new(Object), deps)
	})
}
