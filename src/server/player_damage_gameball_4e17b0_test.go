package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageNative4E17B0ChargeDropsGameBallBeforeShield(t *testing.T) {
	target, source, runtime, _ := playerChargeFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	ball := &Object{TypeInd: 77, ObjOwner: target, ObjFlags: object.FlagNoCollide}
	target.Field129 = ball
	runtime.GameBallType = 77
	target.Buffs = 1 << playerDamageShieldEnchant4E17B0
	var events []string
	runtime.GameBallOnDamage = func(attacker, victim *Object, damage int32) {
		if attacker != source || victim != target || damage != 150 || victim.HealthData.Cur != 2000 || victim.Obj130 != source {
			t.Fatal("GameBall tail arguments/order")
		}
		events = append(events, "drop")
		victim.Field129 = nil
		ball.ObjOwner = nil
		ball.ObjFlags &^= object.FlagNoCollide
	}
	runtime.PlayerSetState = func(*Object, PlayerState) bool { events = append(events, "hurt"); return true }
	runtime.ShieldReduce = func(_ *Object, damage *int32, _ object.DamageType, _ *Object) {
		events = append(events, "shield")
		*damage = 0
	}
	handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime)
	if !handled || result || !reflect.DeepEqual(events, []string{"drop", "hurt", "shield"}) || target.HealthData.Cur != 2000 || ball.ObjOwner != nil {
		t.Fatalf("handled=%t result=%t events=%v HP=%d owner=%p", handled, result, events, target.HealthData.Cur, ball.ObjOwner)
	}
}

func TestPlayerDamageNative4E17B0GameBallUsesFinalDefaultDamage(t *testing.T) {
	for _, tc := range []struct {
		name         string
		damage, want int32
		armor, quest float32
		late         int32
		typ          object.DamageType
	}{
		{"below threshold", 29, 29, 0, 1, 0, object.DamageCrush},
		{"at threshold", 30, 30, 0, 1, 0, object.DamageCrush},
		{"after charge armor", 150, 15, 1.8, 1, 0, object.DamageCrush},
		{"after quest scale", 150, 15, 0, 0.1, 0, object.DamageCrush},
		{"late defend increases", 10, 30, 0, 1, 20, object.DamageCrush},
		{"late defend decreases", 40, 29, 0, 1, -11, object.DamageCrush},
		{"Sentry zap ray", 30, 30, 0, 1, 0, object.DamageZapRay},
		{"monster bite", 30, 30, 0, 1, 0, object.DamageBite},
		{"monster missile", 30, 30, 0, 1, 0, object.DamageImpact},
		{"Flame", 30, 30, 0, 1, 0, object.DamageFlame},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, r, _ := playerChargeFixture4E17B0(t)
			target.UpdateDataPlayer().Field57 = math.Float32bits(tc.armor)
			weapon := source
			switch tc.typ {
			case object.DamageZapRay:
				r.SentryGlobeType = 51
				weapon = &Object{TypeInd: 51, ObjClass: object.ClassSimple | object.ClassImmobile}
				r.PlayerDamageSound = func(*Object, *Object) {}
			case object.DamageBite, object.DamageImpact:
				_, source, _ = playerDamageFixture4E17B0(t)
				weapon = source
				if tc.typ == object.DamageImpact {
					weapon = &Object{ObjClass: object.ClassMissile}
				}
			case object.DamageFlame:
				source = &Object{ObjClass: object.ClassFire}
				weapon = source
			}
			ball := &Object{TypeInd: 77, ObjOwner: target, ObjFlags: object.FlagNoCollide}
			target.Field129 = ball
			r.GameBallType = 77
			if tc.quest != 1 {
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = func() float32 { return tc.quest }
			}
			if tc.late != 0 {
				modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				target.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier, nil}})}
				r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
				r.ApplyLateDefend = func(_ *ModifierEff, _, _, _, _ *Object, d int32, _ object.DamageType) int32 { return d + tc.late }
			}
			calls := 0
			r.GameBallOnDamage = func(a, v *Object, d int32) {
				calls++
				if a != source || v != target || d != tc.want || v.HealthData.Cur != 2000 {
					t.Fatalf("tail damage=%d want=%d HP=%d", d, tc.want, v.HealthData.Cur)
				}
				if d >= 30 {
					ball.ObjOwner = nil
					target.Field129 = nil
				}
			}
			if tc.typ == object.DamageZapRay {
				bindPlayerZapRayDefault4E17B0(&r)
			}
			if h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.damage, tc.typ, r); !h || !result {
				t.Fatalf("handled=%t result=%t", h, result)
			}
			if calls != 1 || (ball.ObjOwner == nil) != (tc.want >= 30) || target.HealthData.Cur != 2000-uint16(tc.want) {
				t.Fatalf("calls=%d owner=%p HP=%d", calls, ball.ObjOwner, target.HealthData.Cur)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0GameBallNilSourceAdmission(t *testing.T) {
	for _, damage := range []int32{4, 30} {
		t.Run(fmt.Sprint(damage), func(t *testing.T) {
			target, _, r, _ := playerChargeFixture4E17B0(t)
			target.Field129 = &Object{TypeInd: 77}
			r.GameBallType = 77
			target.UpdateDataPlayer().Field76 = 19
			called := false
			reason := ""
			r.GameBallOnDamage = func(a, v *Object, d int32) {
				called = true
				if a != nil || v != target || d != 4 {
					t.Fatal("unsafe nil-source call")
				}
			}
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			h, result := PlayerDamageNative4E17B0(target, nil, nil, damage, object.DamagePoison, r)
			if damage == 4 {
				if !h || !result || !called || target.HealthData.Cur != 1996 {
					t.Fatal("small poison incorrectly blocked")
				}
			} else if h || result || called || reason != "source-less GameBall drop" || target.HealthData.Cur != 2000 || target.UpdateDataPlayer().Field76 != 19 {
				t.Fatalf("nil-source preflight: %t/%t/%t %q", h, result, called, reason)
			}
		})
	}
}
