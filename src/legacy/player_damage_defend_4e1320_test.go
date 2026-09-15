package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageLateDefendNative4E1320KnownEffects(t *testing.T) {
	fns := playerDamageLateDefendFunctionsNative4E1320()
	tests := []struct {
		name       string
		function   unsafe.Pointer
		defend     float32
		collision  int32
		damage     int32
		wantDamage int32
	}{
		{
			name: "armor multiplier", function: fns.armorMultiplier, defend: 0.25,
			damage: int32(math.Float32bits(8)), wantDamage: int32(math.Float32bits(2)),
		},
		{
			name: "durability multiplier", function: fns.durabilityMultiplier, defend: 1.25,
			damage: int32(math.Float32bits(8)), wantDamage: int32(math.Float32bits(6)),
		},
		{name: "resilience", function: fns.resilience, damage: 37, wantDamage: 37},
		{name: "breaking", function: fns.breaking, damage: 38, wantDamage: 38},
		{name: "puncture prone", function: fns.punctureProne, damage: 39, wantDamage: 39},
		{name: "inversion enabled", function: fns.inversion, collision: 1, damage: 40, wantDamage: 1},
		{name: "inversion unsigned high", function: fns.inversion, collision: math.MinInt32, damage: 40, wantDamage: 1},
		{name: "inversion disabled", function: fns.inversion, collision: 0, damage: 41, wantDamage: 0},
		{name: "grip enabled", function: fns.grip, collision: 0, damage: 42, wantDamage: 1},
		{name: "grip unsigned high", function: fns.grip, collision: math.MinInt32, damage: 42, wantDamage: 0},
		{name: "grip disabled", function: fns.grip, collision: 1, damage: 43, wantDamage: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			effect := &server.ModifierEff{
				Defend76:        server.ModifierEffFnc{Fnc: test.function, Valf: test.defend},
				DefendCollide88: server.ModifierEffFnc{Val: test.collision},
			}
			if !playerDamageCanApplyLateDefendNative4E1320(effect) {
				t.Fatal("known callback was rejected")
			}
			got := playerDamageApplyLateDefendNative4E1320(
				effect, nil, nil, nil, nil, test.damage, object.DamageBite,
			)
			if got != test.wantDamage {
				t.Fatalf("damage = %#x, want %#x", uint32(got), uint32(test.wantDamage))
			}
		})
	}
}

func TestPlayerDamageLateDefendNative4E1320RejectsUnknownOnWideHost(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		t.Skip("PE32 can invoke arbitrary legacy modifier callbacks")
	}
	effect := &server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	if playerDamageCanApplyLateDefendNative4E1320(effect) {
		t.Fatal("unknown callback accepted on wide host")
	}
}
