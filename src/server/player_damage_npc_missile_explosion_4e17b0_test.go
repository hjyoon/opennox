package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageNative4E17B0NPCMissileExplosionArmor(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, splash := range []bool{false, true} {
			for _, protected := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/splash-%t/protected-%t", fromPlayer, splash, protected), func(t *testing.T) {
					target, source, weapon, missile := damageExplosionFixture4E17B0(t, fromPlayer, false, splash)
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					protection, wantHP := float64(0), uint16(195)
					if protected {
						protection, wantHP = 0.5, 198
					}
					r := damageFlameRuntime4E17B0(t, protection)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					r.GodMode = func() bool { t.Fatal("NPC queried player GodMode"); return true }
					if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageExplosion, r); !h || !result {
						t.Fatalf("handled=%t result=%t", h, result)
					}
					marker, kind, carry := damageMeleeMarker4E17B0(target)
					wantMarker, wantKind := uint32(1), uint32(missile.TypeInd)
					if splash {
						wantMarker, wantKind = 2, 7
					}
					if target.HealthData.Cur != wantHP || armor.HealthData.Cur != 21 || marker != wantMarker || kind != wantKind ||
						math.Abs(float64(carry)+0.1) > 0.000001 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
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

// Defense may replace live update data. Absorption and markers retain their
// entry-time base, whereas fractional carry and durability use the live base.
func damageExplosionCachedLiveOrder4E17B0(t *testing.T, toPlayer bool) {
	t.Helper()
	target, source, weapon, _ := damageExplosionFixture4E17B0(t, !toPlayer, toPlayer, false)
	armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
	var cachedMarker, cachedType, oldCarry, liveCarry *uint32
	var replace func()
	if toPlayer {
		old := target.UpdateDataPlayer()
		live := &PlayerUpdateData{Player: old.Player, Field57: math.Float32bits(0.25), Field21: math.Float32bits(0.25)}
		cachedMarker, cachedType, oldCarry, liveCarry = &old.Field76, &old.Field75, &old.Field21, &live.Field21
		replace = func() { target.UpdateData = unsafe.Pointer(live) }
	} else {
		old := target.UpdateDataMonster()
		live := &MonsterUpdateData{Field518: math.Float32bits(0.25), Field1: math.Float32bits(0.25)}
		cachedMarker, cachedType, oldCarry, liveCarry = &old.Field547, &old.Field546, &old.Field1, &live.Field1
		replace = func() { target.UpdateData = unsafe.Pointer(live) }
	}
	*cachedMarker, *cachedType = 99, 77
	target.Buffs = 1 << playerDamageReflectEnchant4E17B0
	r := damageFlameRuntime4E17B0(t, 0)
	damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
	var events []string
	r.BlockDirection = func(*Object, types.Pointf) bool {
		events = append(events, "defense")
		replace()
		return false
	}
	quest := false
	r.DamageArmor = func(item, s, w *Object, amount int32, typ object.DamageType) bool {
		if item != armor || s != source || w != weapon || amount != 4 || typ != object.DamageExplosion ||
			*cachedMarker != 1 || *cachedType != uint32(weapon.TypeInd) || *oldCarry != math.Float32bits(0.4) || *liveCarry != math.Float32bits(-0.25) {
			t.Fatalf("cached/live wear=%d marker=%d/%d carry=%#x/%#x", amount, *cachedMarker, *cachedType, *oldCarry, *liveCarry)
		}
		events = append(events, "armor")
		*cachedMarker = 0
		quest = true
		return true
	}
	r.GodMode = func() bool {
		if !toPlayer || *cachedMarker != 2 || *cachedType != 7 {
			t.Fatal("GodMode was queried before cached marker fallback or for an NPC")
		}
		events = append(events, "god")
		return false
	}
	r.QuestMode = func() bool {
		if *cachedMarker != 2 || *cachedType != 7 || !quest {
			t.Fatal("QuestMode was queried before armor and marker fallback")
		}
		events = append(events, "quest")
		return quest
	}
	r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
	r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
		if v != target || s != source || w != weapon || amount != 2 || typ != object.DamageExplosion {
			t.Fatalf("late Quest/default identity: damage=%d", amount)
		}
		events = append(events, "default")
		return false
	}
	h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageExplosion, r)
	want := []string{"defense", "armor", "quest", "scale", "default"}
	if toPlayer {
		want = []string{"defense", "armor", "god", "quest", "scale", "default"}
	}
	if !h || result || target.HealthData.Cur != 200 || !slices.Equal(events, want) {
		t.Fatalf("handled=%t result=%t HP=%d events=%v", h, result, target.HealthData.Cur, events)
	}
}

func TestPlayerDamageNative4E17B0NPCMissileExplosionCachedLiveOrder(t *testing.T) {
	damageExplosionCachedLiveOrder4E17B0(t, false)
}

func damageExplosionRounding4E17B0(t *testing.T, toPlayer bool) {
	t.Helper()
	for _, tc := range []struct {
		damage, effective int32
		armor, carry      float32
		wantCarry         uint32
	}{
		{9, 4, 0.5, 0, math.Float32bits(0.5)},
		{11, 6, 0.5, 0, math.Float32bits(-0.5)},
		{-9, -4, 0.5, 0, math.Float32bits(-0.5)},
		{1, 1, 1, 0, 0},
		{0, 0, 0.5, 0, 0},
		{3, 3, 0.1, 0, 0xbe999998},
		{9, 7, 0.2, 0, 0x3e4cccc0},
		{3, 2, 0.4, 0, 0xbe4cccd0},
	} {
		t.Run(fmt.Sprintf("%d/%g", tc.damage, tc.armor), func(t *testing.T) {
			target, source, weapon, _ := damageExplosionFixture4E17B0(t, !toPlayer, toPlayer, true)
			armor := damageMeleeArmorFixture4E17B0(target, tc.armor, tc.carry)
			r := damageFlameRuntime4E17B0(t, 0)
			damageMeleeArmorRuntime4E17B0(&r, armor, tc.armor)
			got := int32(math.MaxInt32)
			r.DefaultDamage = func(_ *Object, s, w *Object, amount int32, typ object.DamageType) bool {
				if s != source || w != nil || typ != object.DamageExplosion {
					t.Fatal("splash identity changed")
				}
				got = amount
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.damage, object.DamageExplosion, r)
			marker, kind, carry := damageMeleeMarker4E17B0(target)
			if !h || !result || got != tc.effective || marker != 2 || kind != 7 || math.Float32bits(carry) != tc.wantCarry {
				t.Fatalf("handled=%t result=%t damage=%d marker=%d/%d carry=%#x", h, result, got, marker, kind, math.Float32bits(carry))
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCMissileExplosionRounding(t *testing.T) {
	damageExplosionRounding4E17B0(t, false)
}

func damageExplosionAdmission4E17B0(t *testing.T, toPlayer bool) {
	t.Helper()
	for _, tc := range []struct {
		name   string
		mutate func(*PlayerDamageRuntime4E17B0)
	}{
		{"default", func(r *PlayerDamageRuntime4E17B0) { r.DefaultDamage = nil }},
		{"armor", func(r *PlayerDamageRuntime4E17B0) { r.DamageArmor = nil }},
		{"quest", func(r *PlayerDamageRuntime4E17B0) { r.QuestMode = func() bool { return true } }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, weapon, _ := damageExplosionFixture4E17B0(t, !toPlayer, toPlayer, false)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			if toPlayer {
				target.UpdateDataPlayer().Field76, target.UpdateDataPlayer().Field75 = 99, 77
			} else {
				target.UpdateDataMonster().Field547, target.UpdateDataMonster().Field546 = 99, 77
			}
			r := damageFlameRuntime4E17B0(t, 0)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			tc.mutate(&r)
			h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageExplosion, r)
			marker, kind, carry := damageMeleeMarker4E17B0(target)
			if h || result || reason == "" || target.HealthData.Cur != 200 || armor.HealthData.Cur != 25 || marker != 99 || kind != 77 || carry != 0.4 {
				t.Fatalf("admission mutated target: handled=%t result=%t reason=%s", h, result, reason)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCMissileExplosionAdmission(t *testing.T) {
	damageExplosionAdmission4E17B0(t, false)
}
