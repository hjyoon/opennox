package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageWorldExplosionTail4E17B0ArmorAndProtection(t *testing.T) {
	for variant := 0; variant < 2; variant++ {
		for _, protected := range []bool{false, true} {
			t.Run(fmt.Sprintf("barrel-%d/protected-%t", variant+1, protected), func(t *testing.T) {
				target, source := damageMeleeUnitFixture4E17B0(t, true), damageWorldExplosionSource4E17B0(t, variant)
				armor := damageMeleeArmorFixture4E17B0(target, .5, .25)
				protection, wantHP := float64(0), uint16(185)
				if protected {
					protection, wantHP = .5, 192
				}
				r := damageFlameRuntime4E17B0(t, protection)
				damageMeleeArmorRuntime4E17B0(&r, armor, .5)
				handled, result := playerDamagePlayerWeaponlessExplosionTail4E17B0(target, source, target.UpdateDataPlayer(), .5, 30, object.DamageExplosion, r)
				marker, typ, carry := damageMeleeMarker4E17B0(target)
				if !handled || !result || target.HealthData.Cur != wantHP || armor.HealthData.Cur != 10 || marker != 2 || typ != 7 || carry != .25 ||
					target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
					t.Fatalf("handled/result=%t/%t HP=%d/%d armor=%d marker=%d/%d carry=%g", handled, result, target.HealthData.Cur, wantHP, armor.HealthData.Cur, marker, typ, carry)
				}
			})
		}
	}
}

func TestPlayerDamageWorldExplosionTail4E17B0CachedAbsorptionLiveCarryAndQuest(t *testing.T) {
	target, source := damageMeleeUnitFixture4E17B0(t, true), damageWorldExplosionSource4E17B0(t, 0)
	armor := damageMeleeArmorFixture4E17B0(target, .5, .25)
	cached := target.UpdateDataPlayer()
	live := &PlayerUpdateData{Player: &Player{}, State: PlayerState13, Field57: math.Float32bits(.1), Field21: math.Float32bits(.4), Field76: 33, Field75: 44}
	target.UpdateData = unsafe.Pointer(live)
	r := damageFlameRuntime4E17B0(t, 0)
	damageMeleeArmorRuntime4E17B0(&r, armor, .5)
	var events []string
	quest := false
	wear := r.DamageArmor
	r.DamageArmor = func(item, s, w *Object, amount int32, typ object.DamageType) bool {
		// The cached .5 absorption produces four absorbed points; armor wear
		// reloads the live .1 denominator, so this .5 item receives twenty.
		if item != armor || s != source || w != nil || amount != 20 || typ != object.DamageExplosion || cached.Field76 != 0 {
			t.Fatalf("world explosion armor identity/order: amount=%d cached marker=%d", amount, cached.Field76)
		}
		events = append(events, "wear")
		quest = true
		return wear(item, s, w, amount, typ)
	}
	r.GodMode = func() bool { events = append(events, "god"); return false }
	r.QuestMode = func() bool { events = append(events, "quest"); return quest }
	r.QuestDamageScale = func() float32 { events = append(events, "scale"); return .5 }
	ordinary := r.DefaultDamage
	r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
		if v != target || s != source || w != nil || amount != 2 || typ != object.DamageExplosion {
			t.Fatal("world explosion changed Quest amount or original source")
		}
		events = append(events, "default")
		return ordinary(v, s, w, amount, typ)
	}
	if h, result := playerDamagePlayerWeaponlessExplosionTail4E17B0(target, source, cached, .5, 9, object.DamageExplosion, r); !h || !result {
		t.Fatal("world-source carry tail rejected")
	}
	if target.HealthData.Cur != 198 || armor.HealthData.Cur != 5 || cached.Field76 != 2 || cached.Field75 != 7 || live.Field76 != 33 || live.Field75 != 44 ||
		math.Abs(float64(math.Float32frombits(live.Field21))+.1) > 1e-6 || cached.Field21 != math.Float32bits(.25) || !slices.Equal(events, []string{"wear", "god", "quest", "scale", "default"}) {
		t.Fatalf("HP=%d armor=%d cached marker/carry=%d/%d/%x live marker/carry=%d/%d/%x events=%v", target.HealthData.Cur, armor.HealthData.Cur, cached.Field76, cached.Field75, cached.Field21, live.Field76, live.Field75, live.Field21, events)
	}
}

func TestPlayerDamageWorldExplosionTail4E17B0KeepsUnitRecordGuard(t *testing.T) {
	for _, class := range []object.Class{object.ClassPlayer, object.ClassMonster, object.ClassSimple | object.ClassPlayer, object.ClassSimple | object.ClassMissile} {
		target := damageMeleeUnitFixture4E17B0(t, true)
		source := &Object{ObjClass: class}
		r := damageFlameRuntime4E17B0(t, 0)
		reason := ""
		r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
		if h, result := playerDamagePlayerWeaponlessExplosionTail4E17B0(target, source, target.UpdateDataPlayer(), 0, 30, object.DamageExplosion, r); h || result || reason == "" || target.HealthData.Cur != 200 || target.Obj130 != nil {
			t.Fatalf("unsupported class=%x reached HP or lost diagnostic", class)
		}
	}
}
