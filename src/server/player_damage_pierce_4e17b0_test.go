package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerDamagePierceFixture4E17B0(t *testing.T) (*Object, *Object, *Object, PlayerDamageRuntime4E17B0, *[]int32) {
	t.Helper()
	target, source, arrow, tail := defaultDamagePierceFixture4E0B30(t)
	var damages []int32
	r := playerDamageRuntime4E17B0(t, target.DamageSound, &damages)
	r.BlockSourceExcluded = func(*Object) bool { return false }
	r.BlockDirection = func(*Object, types.Pointf) bool { return false }
	tail.DamageClear = r.DamageClear
	r.DefaultDamage = func(v, a, w *Object, damage int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(v, a, w, damage, typ, tail)
	}
	return target, source, arrow, r, &damages
}

func TestPlayerDamageNative4E17B0MissilePierceFractionalCarry(t *testing.T) {
	target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.Field57 = math.Float32bits(0.25)
	update.Field40_0 = 0x1234
	for i, carry := range []float32{0.25, 0.5, -0.25} {
		if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || !result {
			t.Fatalf("PIERCE %d = %t/%t", i, h, result)
		}
		if math.Float32frombits(update.Field21) != carry || update.Field75 != uint32(arrow.TypeInd) || update.Field76 != 1 {
			t.Fatalf("PIERCE %d carry=%g marker=%#x/%d", i, math.Float32frombits(update.Field21), update.Field75, update.Field76)
		}
	}
	if !reflect.DeepEqual(*damages, []int32{2, 2, 3}) || target.HealthData.Cur != 13 || update.Field40_0 != 0x1234 || target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 700 || target.Pos132 != arrow.PrevPos {
		t.Fatalf("PIERCE damage=%v HP=%d source=%p", *damages, target.HealthData.Cur, target.Obj130)
	}
}

func TestPlayerDamageNative4E17B0PierceArmorBeforeGodMode(t *testing.T) {
	target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.Field57 = math.Float32bits(0.25)
	carry := float32(0.25)
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 100, Max: 100}, UpdateData: unsafe.Pointer(&carry), InitData: unsafe.Pointer(&ModifierInitData{})}
	target.InvFirstItem = armor
	var events []string
	r.ItemArmorValue = func(*Object) float32 { return 0.25 }
	r.CanDamageArmor = func(item *Object) bool { return item == armor }
	r.DamageArmor = func(item, a, w *Object, damage int32, typ object.DamageType) bool {
		if item != armor || a != source || w != arrow || damage != 2 || typ != object.DamageImpale || update.Field75 != 529 || update.Field76 != 1 || update.Field21 != 0 {
			t.Fatal("absorbed PIERCE armor arguments/order")
		}
		events = append(events, "armor")
		item.HealthData.Cur -= uint16(damage)
		return true
	}
	r.ReportArmorHealth = func(v, item *Object, before, after uint16) {
		if v != target || item != armor || before != 100 || after != 98 {
			t.Fatal("armor health report")
		}
		events = append(events, "report")
	}
	r.GodMode = func() bool { events = append(events, "god"); return true }
	r.QuestMode = func() bool { return true }
	r.QuestDamageScale = func() float32 { t.Fatal("GodMode reached Quest scaling"); return 0 }
	r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
		t.Fatal("GodMode reached default damage")
		return false
	}
	if h, result := PlayerDamageNative4E17B0(target, source, arrow, 8, object.DamageImpale, r); !h || !result {
		t.Fatalf("GodMode PIERCE = %t/%t", h, result)
	}
	if !reflect.DeepEqual(events, []string{"armor", "report", "god"}) || armor.HealthData.Cur != 98 || carry != 0.25 || target.HealthData.Cur != 20 || len(*damages) != 0 {
		t.Fatalf("events=%v armor=%d carry=%g HP=%d", events, armor.HealthData.Cur, carry, target.HealthData.Cur)
	}
}

func TestPlayerDamageNative4E17B0PierceShieldPreservesMissileMarker(t *testing.T) {
	target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.State, update.Player.ArmorEquip = PlayerState16, 0x1000000
	shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10}}
	target.InvFirstItem = shield
	r.BlockDirection = func(v *Object, p types.Pointf) bool { return v == target && p == arrow.PrevPos }
	r.BlockDamagePercent = func() float64 { return 0.5 }
	r.CanDamageBlockItem = func(item *Object) bool { return item == shield }
	r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("intact shield changed state"); return false }
	r.ProjectileReflect = func(_, _ *Object) { t.Fatal("subclass 0x10 GolemArrow reflected from ordinary shield") }
	r.Audio = func(id int, v *Object) {
		if id != 878 || v != target || update.Field75 != 529 || update.Field76 != 1 {
			t.Fatal("shield marker/audio order")
		}
	}
	r.DamageBlockItem = func(item, v, a, w *Object, damage float32, typ object.DamageType) bool {
		if item != shield || v != target || a != source || w != arrow || damage != 4 || typ != object.DamageImpale {
			t.Fatal("shield durability arguments")
		}
		shield.HealthData.Cur -= 4
		return true
	}
	if h, result := PlayerDamageNative4E17B0(target, source, arrow, 8, object.DamageImpale, r); !h || result {
		t.Fatalf("shield PIERCE = %t/%t", h, result)
	}
	if target.HealthData.Cur != 20 || shield.HealthData.Cur != 6 || len(*damages) != 0 || update.Field21 != 0 {
		t.Fatal("shield block reached armor/HP pass")
	}
}

func TestPlayerDamageNative4E17B0PierceMinimumAndQuest(t *testing.T) {
	for _, tc := range []struct {
		name              string
		armor, questScale float32
		damage, want      int32
	}{
		{"minimum", 1, 1, 1, 1},
		{"Quest minimum", 1, 0, 1, 1},
		{"Quest scaling after armor", 0.25, 0.5, 5, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
			target.UpdateDataPlayer().Field57 = math.Float32bits(tc.armor)
			r.QuestMode = func() bool { return true }
			r.QuestDamageScale = func() float32 { return tc.questScale }
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, tc.damage, object.DamageImpale, r); !h || !result {
				t.Fatalf("PIERCE = %t/%t", h, result)
			}
			if !reflect.DeepEqual(*damages, []int32{tc.want}) || target.HealthData.Cur != uint16(20-tc.want) {
				t.Fatalf("damage=%v HP=%d", *damages, target.HealthData.Cur)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PierceReflectShield(t *testing.T) {
	for _, front := range []bool{false, true} {
		t.Run(map[bool]string{false: "behind damages", true: "front reflects"}[front], func(t *testing.T) {
			target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
			target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
			var events []string
			r.BlockDirection = func(v *Object, p types.Pointf) bool {
				if v != target || p != arrow.PosVec {
					t.Fatal("Reflect Shield uses current missile position")
				}
				return front
			}
			r.ProjectileReflect = func(w, v *Object) {
				if w != arrow || v != target {
					t.Fatal("reflect arguments")
				}
				events = append(events, "reflect")
			}
			r.ClearOwner = func(w *Object) {
				if w != arrow {
					t.Fatal("clear owner")
				}
				events = append(events, "clear")
			}
			r.SetOwner = func(v, w *Object) {
				if v != target || w != arrow {
					t.Fatal("set owner")
				}
				events = append(events, "owner")
			}
			r.Audio = func(id int, v *Object) {
				if id != 122 || v != target {
					t.Fatal("reflect audio")
				}
				events = append(events, "audio")
			}
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || result == front {
				t.Fatalf("reflect PIERCE = %t/%t", h, result)
			}
			if front {
				if !reflect.DeepEqual(events, []string{"reflect", "clear", "owner", "audio"}) || len(*damages) != 0 || target.HealthData.Cur != 20 || target.UpdateDataPlayer().Field76 != 0 || target.UpdateDataPlayer().Field21 != 0 {
					t.Fatalf("reflection events=%v damage=%v", events, *damages)
				}
			} else if len(events) != 0 || !reflect.DeepEqual(*damages, []int32{3}) || target.HealthData.Cur != 17 {
				t.Fatalf("behind events=%v damage=%v", events, *damages)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PierceCachedArmorAndLiveTail(t *testing.T) {
	target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
	old := target.UpdateDataPlayer()
	old.Field57 = math.Float32bits(0.25)
	old.State, old.Player.ArmorEquip = PlayerState16, 0x1000000
	// Direction runs before the type switch. Absorption uses the value
	// cached at 004E1865; durability reads the newly live value at 004E2180.
	r.BlockDirection = func(*Object, types.Pointf) bool { old.Field57 = math.Float32bits(0.5); return false }
	carry := float32(0)
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 50}, UpdateData: unsafe.Pointer(&carry), InitData: unsafe.Pointer(&ModifierInitData{})}
	target.InvFirstItem = armor
	r.ItemArmorValue = func(*Object) float32 { return 0.25 }
	r.CanDamageArmor = func(item *Object) bool { return item == armor }
	live := &PlayerUpdateData{Player: old.Player, State: PlayerState13}
	r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
		if item != armor || a != source || w != arrow || d != 4 || typ != object.DamageImpale || old.Field76 != 1 || old.Field75 != 529 {
			t.Fatalf("cached/live armor damage=%d", d)
		}
		item.HealthData.Cur -= uint16(d)
		target.UpdateData = unsafe.Pointer(live)
		return true
	}
	if h, result := PlayerDamageNative4E17B0(target, source, arrow, 32, object.DamageImpale, r); !h || !result {
		t.Fatalf("cached/live PIERCE = %t/%t", h, result)
	}
	if !reflect.DeepEqual(*damages, []int32{24}) || target.HealthData.Cur != 0 || armor.HealthData.Cur != 46 || old.Field75 != 529 || old.Field76 != 1 || live.Field75 != 0 || live.Field76 != 0 {
		t.Fatalf("damage=%v armor=%d old marker=%d/%d live marker=%d/%d", *damages, armor.HealthData.Cur, old.Field75, old.Field76, live.Field75, live.Field76)
	}
}

func TestPlayerDamageNative4E17B0PierceAdmissionBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Object, *Object, *Object, *PlayerDamageRuntime4E17B0)
	}{
		{"default service", func(_, _, _ *Object, r *PlayerDamageRuntime4E17B0) { r.DefaultDamage = nil }},
		{"Quest service", func(_, _, _ *Object, r *PlayerDamageRuntime4E17B0) {
			r.QuestMode = func() bool { return true }
			r.QuestDamageScale = nil
		}},
		{"monster update", func(_, a, _ *Object, _ *PlayerDamageRuntime4E17B0) { a.UpdateData = nil }},
		{"player source", func(_, a, _ *Object, _ *PlayerDamageRuntime4E17B0) { a.ObjClass = object.ClassPlayer }},
		{"weapon class", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjClass |= object.ClassWeapon }},
		{"wand class", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjClass |= object.ClassWand }},
		{"unit class", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjClass |= object.ClassMonster }},
		{"GreatStaff defense", func(v, _, _ *Object, _ *PlayerDamageRuntime4E17B0) { v.UpdateDataPlayer().Player.WeaponEquip = 0x400 }},
		{"armor callback", func(v, _, _ *Object, r *PlayerDamageRuntime4E17B0) {
			v.UpdateDataPlayer().Field57 = math.Float32bits(0.5)
			r.ItemArmorValue = func(*Object) float32 { return 0.5 }
			carry := float32(0.25)
			v.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10}, UpdateData: unsafe.Pointer(&carry), InitData: unsafe.Pointer(&ModifierInitData{})}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t)
			tc.change(target, source, arrow, &r)
			update := target.UpdateDataPlayer()
			update.Field75, update.Field76 = 77, 88
			before, beforeUpdate := *target, *update
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, 8, object.DamageImpale, r); h || result || reason == "" {
				t.Fatalf("admission = %t/%t reason=%q", h, result, reason)
			}
			if *target != before || *update != beforeUpdate || target.HealthData.Cur != 20 || len(*damages) != 0 || source.Class().Has(object.ClassMonster) && source.UpdateData != nil && source.UpdateDataMonster().Field130 != 0 {
				t.Fatal("unsupported PIERCE mutated player/source")
			}
		})
	}
}
