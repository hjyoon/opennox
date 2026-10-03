package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageNative4E17B0PlayerMissileExplosionArmorAndGod(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, splash := range []bool{false, true} {
			for _, god := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/splash-%t/god-%t", fromPlayer, splash, god), func(t *testing.T) {
					target, source, weapon, missile := damageExplosionFixture4E17B0(t, fromPlayer, true, splash)
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					r := damageFlameRuntime4E17B0(t, 0.5)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					r.GodMode = func() bool {
						if armor.HealthData.Cur != 21 || math.Abs(float64(math.Float32frombits(target.UpdateDataPlayer().Field21))+0.1) > 0.000001 {
							t.Fatal("GodMode precedes armor wear or fractional carry")
						}
						return god
					}
					if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageExplosion, r); !h || !result {
						t.Fatalf("handled=%t result=%t", h, result)
					}
					marker, kind, carry := damageMeleeMarker4E17B0(target)
					wantMarker, wantKind := uint32(1), uint32(missile.TypeInd)
					if splash {
						wantMarker, wantKind = 2, 7
					}
					wantHP := uint16(198)
					if god {
						wantHP = 200
					}
					if target.HealthData.Cur != wantHP || armor.HealthData.Cur != 21 || marker != wantMarker || kind != wantKind || math.Abs(float64(carry)+0.1) > 0.000001 {
						t.Fatalf("HP=%d armor=%d marker=%d/%d carry=%g", target.HealthData.Cur, armor.HealthData.Cur, marker, kind, carry)
					}
					if !god && (target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 7 || target.Frame134 != 1400) {
						t.Fatal("player EXPLOSION attribution changed")
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0PlayerMissileExplosionCachedLiveOrder(t *testing.T) {
	damageExplosionCachedLiveOrder4E17B0(t, true)
}

func TestPlayerDamageNative4E17B0PlayerMissileExplosionRounding(t *testing.T) {
	damageExplosionRounding4E17B0(t, true)
}

func TestPlayerDamageNative4E17B0PlayerMissileExplosionAdmission(t *testing.T) {
	damageExplosionAdmission4E17B0(t, true)
}

func TestPlayerDamageNative4E17B0PlayerMissileExplosionReflect(t *testing.T) {
	for _, splash := range []bool{false, true} {
		t.Run(fmt.Sprintf("splash-%t", splash), func(t *testing.T) {
			target, source, weapon, missile := damageExplosionFixture4E17B0(t, false, true, splash)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			target.Buffs = 1 << playerDamageReflectEnchant4E17B0
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 99, 77
			r := damageFlameRuntime4E17B0(t, 0)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			r.BlockDirection = func(*Object, types.Pointf) bool { return true }
			reflected := 0
			r.ProjectileReflect = func(got, owner *Object) {
				if got != missile || owner != target {
					t.Fatal("wrong reflected projectile")
				}
				reflected++
			}
			r.ClearOwner = func(got *Object) { got.ObjOwner = nil }
			r.SetOwner = func(owner, got *Object) { got.ObjOwner = owner }
			r.Audio = func(id int, owner *Object) {
				if id != 122 || owner != target {
					t.Fatal("wrong reflect sound")
				}
			}
			if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageExplosion, r); !h || result || reflected != 1 ||
				target.HealthData.Cur != 200 || armor.HealthData.Cur != 25 || ud.Field76 != 0 || ud.Field75 != 77 || ud.Field21 != math.Float32bits(0.4) || missile.ObjOwner != target {
				t.Fatal("Reflect Shield did not suppress EXPLOSION damage/wear")
			}
		})
	}
}
