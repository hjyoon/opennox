package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

func damageExplosionFixture4E17B0(t *testing.T, fromPlayer, toPlayer, splash bool) (target, source, weapon, missile *Object) {
	t.Helper()
	target, source, weapon, missile = damageFlameFixture4E17B0(t, fromPlayer, toPlayer, splash)
	missile.TypeInd = 697 // independent synthetic missile type, not a damage result
	return
}

func TestDefaultDamageWorld4E0B30MissileExplosionMatrix(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, toPlayer := range []bool{false, true} {
			for _, splash := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/splash-%t", fromPlayer, toPlayer, splash), func(t *testing.T) {
					target, source, weapon, missile := damageExplosionFixture4E17B0(t, fromPlayer, toPlayer, splash)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.FireProtection = func(*Object) float64 { return 0.5 }
					if !DefaultDamageWorld4E0B30(target, source, weapon, 9, object.DamageExplosion, r) ||
						target.HealthData.Cur != 196 || target.Obj130 != missile || target.Pos132 != missile.PrevPos ||
						target.Field131 != 7 || target.Frame134 != 1400 {
						t.Fatalf("EXPLOSION HP=%d attribution=%p position=%v type/frame=%d/%d", target.HealthData.Cur, target.Obj130, target.Pos132, target.Field131, target.Frame134)
					}
					if !toPlayer {
						ud := target.UpdateDataMonster()
						wantMarker, wantType := uint32(1), uint32(missile.TypeInd)
						if splash {
							wantMarker, wantType = 2, 7
						}
						if ud.Field547 != wantMarker || ud.Field546 != wantType || !ud.StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) {
							t.Fatalf("NPC marker=%d/%d flags=%x", ud.Field547, ud.Field546, ud.StatusFlags)
						}
					}
					if !fromPlayer && source.FindOwnerChainPlayer().UpdateDataMonster().Field130 != 1400 {
						t.Fatal("NPC successful-hit timestamp missing")
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30MissileExplosionFireImmuneHalving(t *testing.T) {
	// 004E0D5A divides signed EXPLOSION by two before fire protection.
	// FLAME/LAVA immunity is a rejection, not this type-7 attenuation.
	for _, tc := range []struct {
		damage     int32
		protection float64
		want       int32
	}{
		{9, 0.5, 2}, {11, 0.5, 2}, {-9, 0.5, -2}, {-11, 0.5, -2},
		{9, 0, 4}, {-9, 0, -4}, {1, 0, 1}, {0, 0, 1}, {-1, 0, 1},
		{math.MinInt32, 0, -1073741824}, {math.MaxInt32, 0, 1073741824},
	} {
		t.Run(fmt.Sprintf("%d/%g", tc.damage, tc.protection), func(t *testing.T) {
			target, source, weapon, _ := damageExplosionFixture4E17B0(t, true, false, false)
			target.ObjSubClass |= 0x400
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 { return tc.protection }
			got := int32(123456)
			r.DamageClear = func(_ *Object, amount int32) { got = amount }
			DefaultDamageWorld4E0B30(target, source, weapon, tc.damage, object.DamageExplosion, r)
			if got != tc.want {
				t.Fatalf("damage=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30MissileExplosionCampaignGate(t *testing.T) {
	for _, enemy := range []bool{false, true} {
		target, source, weapon, _ := damageExplosionFixture4E17B0(t, true, false, false)
		r := damageMeleeWorldRuntime4E0B30(t)
		r.GameplayFlag1 = func() bool { return false }
		r.IsEnemy = func(*Object, *Object) bool { return enemy }
		r.FireProtection = func(*Object) float64 { return 0 }
		DefaultDamageWorld4E0B30(target, source, weapon, 9, object.DamageExplosion, r)
		want := uint16(200)
		if enemy {
			want = 191
		}
		if target.HealthData.Cur != want {
			t.Fatalf("enemy=%t HP=%d want=%d", enemy, target.HealthData.Cur, want)
		}
	}
}
