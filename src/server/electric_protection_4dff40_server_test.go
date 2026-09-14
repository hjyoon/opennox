package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestElectricProtectionNative4DFF40ModifiersAndEnchant(t *testing.T) {
	marker := unsafe.Pointer(new(byte))
	other := unsafe.Pointer(new(byte))
	armor := &Object{
		ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{
			{Engage112: marker, EngageFloat120: 0.125},
			{Engage112: other, EngageFloat120: 0.5},
		}}),
	}
	unit := &Object{
		ObjClass: object.ClassPlayer, InvFirstItem: armor,
		Buffs: uint32(1) << ENCHANT_PROTECT_FROM_ELECTRICITY,
	}
	unit.BuffsPower[ENCHANT_PROTECT_FROM_ELECTRICITY] = 2
	got := elementalProtectionNative4DFE40(unit, marker, func(key string, index int32) float64 {
		if key != electricProtectionBalanceKey4DFF40 || index != 1 {
			t.Fatalf("balance args = (%q,%d), want (%q,1)", key, index, electricProtectionBalanceKey4DFF40)
		}
		return 0.25
	}, ENCHANT_PROTECT_FROM_ELECTRICITY, electricProtectionBalanceKey4DFF40)
	if got != 0.375 {
		t.Fatalf("electric protection = %v, want 0.375", got)
	}
	armor.InitData = nil
	got = elementalProtectionNative4DFE40(unit, marker, func(string, int32) float64 { return 1 },
		ENCHANT_PROTECT_FROM_ELECTRICITY, electricProtectionBalanceKey4DFF40)
	if got != float64(math.Float32frombits(fireProtectionFinalLimitBits)) {
		t.Fatalf("capped protection = %v", got)
	}
}
