package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
)

func damageFlameRuntime4E17B0(t *testing.T, protection float64) PlayerDamageRuntime4E17B0 {
	t.Helper()
	world := damageMeleeWorldRuntime4E0B30(t)
	world.FireProtection = func(*Object) float64 { return protection }
	return PlayerDamageRuntime4E17B0{
		Frame: world.Frame,
		DefaultDamage: func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
			return DefaultDamageWorld4E0B30(target, source, weapon, damage, typ, world)
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("missile FLAME rejected: %s", reason)
		},
	}
}

func TestPlayerDamageNative4E17B0NPCMissileFlameArmor(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, splash := range []bool{false, true} {
			for _, protected := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/splash-%t/protected-%t", fromPlayer, splash, protected), func(t *testing.T) {
					target, source, weapon, missile := damageFlameFixture4E17B0(t, fromPlayer, false, splash)
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					protection, wantHP := float64(0), uint16(191)
					if protected {
						protection, wantHP = 0.5, 196
					}
					r := damageFlameRuntime4E17B0(t, protection)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					r.GodMode = func() bool { t.Fatal("NPC queried player GodMode"); return true }
					if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageFlame, r); !h || !result {
						t.Fatalf("handled=%t result=%t", h, result)
					}
					marker, kind, carry := damageMeleeMarker4E17B0(target)
					wantMarker, wantKind := uint32(1), uint32(missile.TypeInd)
					if splash {
						wantMarker, wantKind = 2, 1
					}
					if target.HealthData.Cur != wantHP || armor.HealthData.Cur != 16 || marker != wantMarker || kind != wantKind ||
						carry != 0.4 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 1 || target.Frame134 != 1400 {
						t.Fatalf("HP=%d armor=%d marker=%d/%d carry=%g source=%p", target.HealthData.Cur, armor.HealthData.Cur, marker, kind, carry, target.Obj130)
					}
					if !target.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured | object.MonStatusOnFire) {
						t.Fatal("missing injured/fire status")
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0NPCMissileFlameLiveQuestOrder(t *testing.T) {
	target, source, weapon, _ := damageFlameFixture4E17B0(t, true, false, false)
	armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
	r := damageFlameRuntime4E17B0(t, 0)
	damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
	quest := false
	var events []string
	wear := r.DamageArmor
	r.DamageArmor = func(item, s, w *Object, amount int32, typ object.DamageType) bool {
		if amount != 9 || target.UpdateDataMonster().Field547 != 1 {
			t.Fatal("raw wear/marker prefix changed")
		}
		events = append(events, "armor")
		quest = true
		return wear(item, s, w, amount, typ)
	}
	r.QuestMode = func() bool { events = append(events, "quest"); return quest }
	r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
	damage := r.DefaultDamage
	r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
		if amount != 4 {
			t.Fatalf("Quest rounded FLAME=%d", amount)
		}
		events = append(events, "default")
		return damage(v, s, w, amount, typ)
	}
	if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageFlame, r); !h || !result ||
		target.HealthData.Cur != 196 || !slices.Equal(events, []string{"armor", "quest", "scale", "default"}) {
		t.Fatalf("HP=%d events=%v", target.HealthData.Cur, events)
	}
}

func TestPlayerDamageNative4E17B0NPCMissileFlameAdmission(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*PlayerDamageRuntime4E17B0)
	}{
		{"default", func(r *PlayerDamageRuntime4E17B0) { r.DefaultDamage = nil }},
		{"armor", func(r *PlayerDamageRuntime4E17B0) { r.DamageArmor = nil }},
		{"quest", func(r *PlayerDamageRuntime4E17B0) { r.QuestMode = func() bool { return true } }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, weapon, _ := damageFlameFixture4E17B0(t, true, false, false)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			ud := target.UpdateDataMonster()
			ud.Field547, ud.Field546 = 99, 77
			r := damageFlameRuntime4E17B0(t, 0)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			tc.mutate(&r)
			if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageFlame, r); h || result || reason == "" ||
				target.HealthData.Cur != 200 || armor.HealthData.Cur != 25 || ud.Field547 != 99 || ud.Field546 != 77 || ud.Field1 != math.Float32bits(0.4) {
				t.Fatalf("admission mutated target: handled=%t result=%t reason=%s", h, result, reason)
			}
		})
	}
}
