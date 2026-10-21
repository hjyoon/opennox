package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageSelfMissileFlame4E17B0HPArmorAndAttribution(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, protection := range []float64{0, .5} {
			for _, god := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-%t/protection-%g/god-%t", player, protection, god), func(t *testing.T) {
					v, _, _, w := damageFlameFixture4E17B0(t, false, player, false)
					if player {
						v.TypeInd, v.ObjClass, v.ObjFlags = 713, 2621444, 16777732
					}
					w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = 694, 2621441, 1, 553665028
					armor := damageMeleeArmorFixture4E17B0(v, .5, .4)
					armor.HealthData.Cur, armor.HealthData.Max = 200, 200
					r := damageFlameRuntime4E17B0(t, protection)
					damageMeleeArmorRuntime4E17B0(&r, armor, .5)
					r.GodMode = func() bool { return god }
					before := *w
					handled, result := PlayerDamageNative4E17B0(v, w, w, 96, object.DamageFlame, r)
					wantHP := uint16(104)
					if protection != 0 {
						wantHP = 152
					}
					if player && god {
						wantHP = 200
					}
					marker, kind, carry := damageMeleeMarker4E17B0(v)
					if !handled || !result || v.HealthData.Cur != wantHP || armor.HealthData.Cur != 104 ||
						marker != 2 || kind != 1 || carry != .4 || *w != before {
						t.Fatalf("self-sourced FLAME: handled=%t result=%t HP=%d/%d armor=%d marker=%d/%d carry=%g", handled, result, v.HealthData.Cur, wantHP, armor.HealthData.Cur, marker, kind, carry)
					}
					if !(player && god) && (v.Obj130 != w || v.Pos132 != w.PrevPos || v.Field131 != 1 || v.Frame134 != 1400) {
						t.Fatal("self-sourced FLAME lost the missile attribution")
					}
				})
			}
		}
	}
}

func TestSelfMissileFlameAdmission4E17B0(t *testing.T) {
	for _, id := range []uint16{0, 694, 695, math.MaxUint16} {
		for _, extra := range []object.Class{0, object.ClassPlayer, object.ClassMonster, object.ClassWeapon, object.ClassWand} {
			t.Run(fmt.Sprintf("type-%d/extra-%x", id, uint32(extra)), func(t *testing.T) {
				w := &Object{TypeInd: id, ObjClass: object.Class(2621441) | extra, ObjSubClass: 1}
				for typ := object.DamageType(0); typ <= object.DamageAirborneElectric; typ++ {
					// Self-sourced CRUSH already supports DeathBall. Adding FLAME
					// must preserve that admission and reject all other types.
					want := extra == 0 && (typ == object.DamageFlame || typ == object.DamageCrush)
					if got := playerDamageWorldProjectileShape4E17B0(w, w, typ); got != want {
						t.Fatalf("class=%x type=%d admitted=%t want=%t", uint32(w.Class()), typ, got, want)
					}
				}
			})
		}
	}
}
