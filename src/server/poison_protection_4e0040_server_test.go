package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPoisonProtection4E0040NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantBuffs := uintptr(340)
	wantBuffPower := uintptr(408)
	wantNext := uintptr(496)
	wantFirst := uintptr(504)
	wantInit := uintptr(692)
	wantInitSize := uintptr(20)
	wantInitField16 := uintptr(16)
	wantModifierSize := uintptr(144)
	wantEngage := uintptr(112)
	wantEngageFloat := uintptr(120)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantBuffs = 344
		wantBuffPower = 412
		wantNext = 528
		wantFirst = 544
		wantInit = 760
		wantInitSize = 40
		wantInitField16 = 32
		wantModifierSize = 208
		wantEngage = 160
		wantEngageFloat = 176
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Buffs", unsafe.Offsetof(Object{}.Buffs), wantBuffs},
		{"Object.BuffsPower", unsafe.Offsetof(Object{}.BuffsPower), wantBuffPower},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInit},
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantInitSize},
		{"ModifierInitData.Modifiers", unsafe.Offsetof(ModifierInitData{}.Modifiers), 0},
		{"ModifierInitData.Field16", unsafe.Offsetof(ModifierInitData{}.Field16), wantInitField16},
		{"ModifierEff size", unsafe.Sizeof(ModifierEff{}), wantModifierSize},
		{"ModifierEff.Engage112", unsafe.Offsetof(ModifierEff{}.Engage112), wantEngage},
		{"ModifierEff.EngageFloat120", unsafe.Offsetof(ModifierEff{}.EngageFloat120), wantEngageFloat},
		{"modifier value width", unsafe.Sizeof(ModifierEff{}.EngageFloat120), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPoisonProtectionNative4E0040BindsModifierIdentityAndBalance(t *testing.T) {
	marker := unsafe.Pointer(new(byte))
	otherMarker := unsafe.Pointer(new(byte))
	itemMatch := &ModifierEff{Engage112: marker, EngageFloat120: 0.2}
	itemOther := &ModifierEff{Engage112: otherMarker, EngageFloat120: 50}
	itemInit := &ModifierInitData{Modifiers: [4]*ModifierEff{itemMatch, itemOther}}
	item := &Object{
		ObjFlags: object.Flags(poisonProtectionEquippedFlag4E0040),
		ObjClass: object.Class(poisonProtectionClassMask4E0040),
		InitData: unsafe.Pointer(itemInit),
	}
	unitMatch := &ModifierEff{Engage112: marker, EngageFloat120: 0.1}
	unitInit := &ModifierInitData{Modifiers: [4]*ModifierEff{unitMatch}}
	unit := &Object{
		ObjClass:     object.Class(poisonProtectionClassMask4E0040),
		InvFirstItem: item,
		InitData:     unsafe.Pointer(unitInit),
		Buffs:        uint32(1) << poisonProtectionEnchant4E0040,
	}
	unit.BuffsPower[poisonProtectionEnchant4E0040] = 3
	balanceCalls := 0
	got := poisonProtectionNative4E0040(unit, poisonProtectionNativeDeps4E0040{
		poisonProtectEngage: marker,
		loadBalance: func(key string, index int32) float64 {
			balanceCalls++
			if key != poisonProtectionBalanceKey4E0040 || index != 2 {
				t.Fatalf("balance args = (%q,%d), want (%q,2)", key, index, poisonProtectionBalanceKey4E0040)
			}
			return 0.25
		},
	})
	want := 0.25 + float64(float32(float64(float32(0.2))+float64(float32(0.1))))
	if math.Float64bits(got) != math.Float64bits(want) || balanceCalls != 1 {
		t.Fatalf("result bits/calls = %#016x/%d, want %#016x/1", math.Float64bits(got), balanceCalls, math.Float64bits(want))
	}
}

func TestPoisonProtectionNative4E0040NilIdentityDoesNotMatchNilCallbacks(t *testing.T) {
	modifier := &ModifierEff{EngageFloat120: 0.5}
	init := &ModifierInitData{Modifiers: [4]*ModifierEff{modifier}}
	unit := &Object{
		ObjClass: object.Class(poisonProtectionClassMask4E0040),
		InitData: unsafe.Pointer(init),
	}
	got := poisonProtectionNative4E0040(unit, poisonProtectionNativeDeps4E0040{
		loadBalance: func(string, int32) float64 {
			t.Fatal("unit without buff loaded balance")
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("nil identity result = %v, want 0", got)
	}
}

func TestPoisonProtectionNative4E0040NilUnitAndNilInit(t *testing.T) {
	deps := poisonProtectionNativeDeps4E0040{
		poisonProtectEngage: unsafe.Pointer(new(byte)),
		loadBalance: func(string, int32) float64 {
			t.Fatal("balance reached")
			return 0
		},
	}
	if got := poisonProtectionNative4E0040(nil, deps); math.Float64bits(got) != 0 {
		t.Fatalf("nil unit result bits = %#016x, want positive zero", math.Float64bits(got))
	}

	unit := &Object{ObjClass: object.Class(poisonProtectionClassMask4E0040)}
	defer func() {
		if recover() == nil {
			t.Fatal("nil native ModifierInitData did not fault")
		}
	}()
	poisonProtectionNative4E0040(unit, deps)
}
