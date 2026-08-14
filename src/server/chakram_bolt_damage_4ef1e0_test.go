package server

import (
	"math"
	"testing"
)

func TestChakramCalcBoltDamage4EF1E0Branches(t *testing.T) {
	const archerType = uint32(73)
	tests := []struct {
		name        string
		cooperative bool
		typeIndex   uint32
		want        float64
	}{
		{name: "ordinary mode ArcherBolt", typeIndex: archerType, want: 34.25},
		{name: "cooperative other projectile", cooperative: true, typeIndex: 74, want: 34.25},
		{name: "cooperative ArcherBolt", cooperative: true, typeIndex: archerType, want: 36.75},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modifier := &Modifier{
				TypeInd:              tc.typeIndex,
				ReqStrength60:        12,
				DamageCoeffOrArmor64: 1.25,
				DamageMin72:          3,
			}
			got := chakramCalcBoltDamage4EF1E0(37, modifier, tc.cooperative, archerType, 5.5)
			if got != tc.want {
				t.Fatalf("damage = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestChakramCalcBoltDamage4EF1E0WidensBinary32Coefficient(t *testing.T) {
	coefficient := math.Float32frombits(0x3dcccccd)
	modifier := &Modifier{
		TypeInd:              99,
		ReqStrength60:        12,
		DamageCoeffOrArmor64: coefficient,
		DamageMin72:          7,
	}
	want := float64(int32(-5)-int32(modifier.ReqStrength60))*float64(coefficient) +
		float64(modifier.DamageMin72)
	got := chakramCalcBoltDamage4EF1E0(-5, modifier, false, 73, 1234)
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("damage bits = %#016x, want %#016x", math.Float64bits(got), math.Float64bits(want))
	}
}

func TestChakramCalcBoltDamage4EF1E0SoloPathIgnoresModifierMinimum(t *testing.T) {
	modifier := &Modifier{
		TypeInd:              73,
		ReqStrength60:        40,
		DamageCoeffOrArmor64: 0.75,
		DamageMin72:          0xffff,
	}
	got := chakramCalcBoltDamage4EF1E0(20, modifier, true, 73, 4.5)
	if got != -10.5 {
		t.Fatalf("damage = %v, want -10.5", got)
	}
}
