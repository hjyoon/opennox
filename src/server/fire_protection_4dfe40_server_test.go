package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestFireProtectionNative4DFE40ModifierSpellAndCaps(t *testing.T) {
	marker := unsafe.Pointer(new(byte))
	other := unsafe.Pointer(new(byte))
	first := &Object{
		ObjClass: object.ClassArmor,
		ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{
			{Engage112: marker, EngageFloat120: 0.2},
			{Engage112: other, EngageFloat120: 10},
		}}),
	}
	second := &Object{
		ObjClass: object.ClassWeapon,
		ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{
			{Engage112: marker, EngageFloat120: 0.1},
		}}),
	}
	first.InvNextItem = second
	unit := &Object{
		ObjClass:     object.ClassPlayer,
		InvFirstItem: first,
		Buffs:        uint32(1) << ENCHANT_PROTECT_FROM_FIRE,
	}
	unit.BuffsPower[ENCHANT_PROTECT_FROM_FIRE] = 3
	balanceCalls := 0
	got := fireProtectionNative4DFE40(unit, marker, func(key string, index int32) float64 {
		balanceCalls++
		if key != fireProtectionBalanceKey4DFE40 || index != 2 {
			t.Fatalf("balance args = (%q,%d), want (%q,2)", key, index, fireProtectionBalanceKey4DFE40)
		}
		return 0.25
	})
	want := 0.25 + float64(float32(float64(float32(0.2))+float64(float32(0.1))))
	if math.Float64bits(got) != math.Float64bits(want) || balanceCalls != 1 {
		t.Fatalf("result bits/calls = %#016x/%d, want %#016x/1", math.Float64bits(got), balanceCalls, math.Float64bits(want))
	}

	second.InitDataModifier().Modifiers[0].EngageFloat120 = 0.4
	got = fireProtectionNative4DFE40(unit, marker, func(string, int32) float64 { return 0.25 })
	want = float64(math.Float32frombits(fireProtectionFinalLimitBits))
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("clamped result bits = %#016x, want %#016x", math.Float64bits(got), math.Float64bits(want))
	}
}

func TestFireProtectionNative4DFE40NilAndIdentityGates(t *testing.T) {
	if got := fireProtectionNative4DFE40(nil, unsafe.Pointer(new(byte)), func(string, int32) float64 {
		t.Fatal("nil unit loaded balance")
		return 0
	}); math.Float64bits(got) != 0 {
		t.Fatalf("nil result bits = %#016x, want positive zero", math.Float64bits(got))
	}
	unit := &Object{
		ObjClass: object.ClassArmor,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{
			{EngageFloat120: 0.5},
		}}),
	}
	if got := fireProtectionNative4DFE40(unit, nil, func(string, int32) float64 {
		t.Fatal("unit without buff loaded balance")
		return 0
	}); got != 0 {
		t.Fatalf("nil callback identity result = %v, want 0", got)
	}
}

func TestFireProtectionNative4DFE40ZeroPowerKeepsSignedIndex(t *testing.T) {
	unit := &Object{Buffs: uint32(1) << ENCHANT_PROTECT_FROM_FIRE}
	got := fireProtectionNative4DFE40(unit, nil, func(key string, index int32) float64 {
		if key != fireProtectionBalanceKey4DFE40 || index != -1 {
			t.Fatalf("balance args = (%q,%d), want (%q,-1)", key, index, fireProtectionBalanceKey4DFE40)
		}
		return 0.125
	})
	if got != 0.125 {
		t.Fatalf("result = %v, want 0.125", got)
	}
}
