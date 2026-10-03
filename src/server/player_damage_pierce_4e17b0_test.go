package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerDamagePierceFixture4E17B0(t *testing.T, playerSource bool) (*Object, *Object, *Object, PlayerDamageRuntime4E17B0, *[]int32) {
	t.Helper()
	target, source, arrow, tail := defaultDamagePierceFixture4E0B30(t)
	if playerSource {
		// Use a real PlayerUpdateData, not a relabeled MonsterUpdateData.
		// A player source must not depend on the monster hit-sound service.
		target, source, arrow, tail = defaultDamagePlayerPierceFixture4E0B30(t)
	}
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

func playerDamagePierceSources4E17B0(t *testing.T, check func(*testing.T, bool)) {
	t.Helper()
	for _, playerSource := range []bool{false, true} {
		t.Run(map[bool]string{false: "monster source", true: "player source"}[playerSource], func(t *testing.T) {
			check(t, playerSource)
		})
	}
}

func TestPlayerDamageNative4E17B0MissilePierceFractionalCarry(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
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
		if !reflect.DeepEqual(*damages, []int32{2, 2, 3}) || target.HealthData.Cur != 13 || update.Field40_0 != 0x1234 || target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 700 || target.Pos132 != source.PrevPos {
			t.Fatalf("PIERCE damage=%v HP=%d source=%p", *damages, target.HealthData.Cur, target.Obj130)
		}
	})
}

func TestPlayerDamageNative4E17B0PiercePureMissileStillAdmitted(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
		arrow.ObjClass = object.ClassMissile
		if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || !result {
			t.Fatalf("pure-missile PIERCE = %t/%t", h, result)
		}
		if !reflect.DeepEqual(*damages, []int32{3}) || target.HealthData.Cur != 17 || target.Pos132 != arrow.PrevPos {
			t.Fatal("pure-missile PIERCE or its previous-position contract regressed")
		}
	})
}

func TestPlayerDamageNative4E17B0PierceArmorBeforeGodMode(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
		update := target.UpdateDataPlayer()
		update.Field57 = math.Float32bits(0.25)
		carry := damageArmorCarryFixture4E17B0(0.25)
		armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), HealthData: &HealthData{Cur: 100, Max: 100}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
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
		if !reflect.DeepEqual(events, []string{"armor", "report", "god"}) || armor.HealthData.Cur != 98 || *carry != 0.25 || target.HealthData.Cur != 20 || len(*damages) != 0 {
			t.Fatalf("events=%v armor=%d carry=%g HP=%d", events, armor.HealthData.Cur, *carry, target.HealthData.Cur)
		}
	})
}

func TestPlayerDamageNative4E17B0PierceShieldPreservesMissileMarker(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
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
	})
}

func TestPlayerDamageNative4E17B0PierceMinimumAndQuest(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
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
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
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
	})
}

func TestPlayerDamageNative4E17B0PierceReflectShield(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, front := range []bool{false, true} {
			t.Run(map[bool]string{false: "behind damages", true: "front reflects"}[front], func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
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
	})
}

func TestPlayerDamageNative4E17B0PierceCachedArmorAndLiveTail(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
		old := target.UpdateDataPlayer()
		old.Field57 = math.Float32bits(0.25)
		old.State, old.Player.ArmorEquip = PlayerState16, 0x1000000
		// Direction runs before the type switch. Absorption uses the value
		// cached at 004E1865; durability reads the newly live value at 004E2180.
		r.BlockDirection = func(*Object, types.Pointf) bool { old.Field57 = math.Float32bits(0.5); return false }
		carry := damageArmorCarryFixture4E17B0(0)
		armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), HealthData: &HealthData{Cur: 50}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
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
	})
}

func TestPlayerDamageNative4E17B0PierceAdmissionBeforeMutation(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, tc := range []struct {
			name   string
			change func(*Object, *Object, *Object, *PlayerDamageRuntime4E17B0)
		}{
			{"default service", func(_, _, _ *Object, r *PlayerDamageRuntime4E17B0) { r.DefaultDamage = nil }},
			{"Quest service", func(_, _, _ *Object, r *PlayerDamageRuntime4E17B0) {
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = nil
			}},
			{"source update", func(_, a, _ *Object, _ *PlayerDamageRuntime4E17B0) { a.UpdateData = nil }},
			{"non-unit source", func(_, a, _ *Object, _ *PlayerDamageRuntime4E17B0) { a.ObjClass = object.ClassMissile }},
			{"melee weapon subclass", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjSubClass = 0 }},
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
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
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
	})
}

func TestPlayerDamageNative4E17B0PierceSignedDamage(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, tc := range []struct {
			name                     string
			damage                   int32
			armor, carry, questScale float32
			wantDamage               int32
			wantCarry                float32
			wantWear                 int32
		}{
			{"zero", 0, 0.25, 0, 1, 0, 0, 0},
			{"zero with positive carry", 0, 0.25, 0.75, 1, 1, -0.25, 0},
			{"zero with negative carry wears armor", 0, 0.25, -0.75, 1, -1, 0.25, 1},
			{"negative", -3, 0.25, 0, 1, -2, -0.25, 0},
			{"negative with live carry", -3, 0.25, 0.25, 1, -2, 0, 0},
			{"negative fully absorbed is not minimum one", -1, 1, 0, 1, 0, 0, 0},
			{"negative Quest rounds to negative", -3, 0.25, 0, 0.5, -1, -0.25, 0},
			{"negative Quest rounds to zero", -3, 0.25, 0, 0, 0, -0.25, 0},
		} {
			t.Run(tc.name, func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				update := target.UpdateDataPlayer()
				update.Field57, update.Field21 = math.Float32bits(tc.armor), math.Float32bits(tc.carry)
				itemUpdate := &WeaponArmorUpdateData{Field0: math.Float32bits(0.125)}
				armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), UpdateData: unsafe.Pointer(itemUpdate), InitData: unsafe.Pointer(&ModifierInitData{}), HealthData: &HealthData{Cur: 10, Max: 10}}
				target.InvFirstItem = armor
				r.ItemArmorValue = func(*Object) float32 { return tc.armor }
				r.CanDamageArmor = func(item *Object) bool { return item == armor }
				var wears []int32
				r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
					if item != armor || a != source || w != arrow || typ != object.DamageImpale || d != tc.wantWear || update.Field76 != 1 || update.Field75 != 529 {
						t.Fatal("signed PIERCE armor arguments or marker order")
					}
					wears = append(wears, d)
					item.HealthData.Cur -= uint16(d)
					return true
				}
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = func() float32 { return tc.questScale }
				if h, result := PlayerDamageNative4E17B0(target, source, arrow, tc.damage, object.DamageImpale, r); !h || !result {
					t.Fatalf("signed PIERCE = %t/%t", h, result)
				}
				if !reflect.DeepEqual(*damages, []int32{tc.wantDamage}) || target.HealthData.Cur != uint16(20-tc.wantDamage) ||
					update.Field21 != math.Float32bits(tc.wantCarry) || itemUpdate.Field0 != math.Float32bits(0.125) || armor.HealthData.Cur != uint16(10-tc.wantWear) || len(wears) != int(tc.wantWear) ||
					update.Field76 != 1 || update.Field75 != 529 || target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 700 {
					t.Fatalf("damage=%v HP=%d carry=%g wears=%v armor=%d item carry=%g", *damages, target.HealthData.Cur, math.Float32frombits(update.Field21), wears, armor.HealthData.Cur, math.Float32frombits(itemUpdate.Field0))
				}
			})
		}
	})
}

func TestPlayerDamageNative4E17B0PierceFriendlyAndCoop(t *testing.T) {
	for _, tc := range []struct {
		name               string
		coop, own, quest   bool
		wantResult, wantHP bool
	}{
		{"campaign friendly", false, false, false, true, false},
		{"campaign self", false, true, false, true, true},
		{"Quest self", false, true, true, true, false},
		{"Coop own missile", true, true, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, true)
			if tc.own {
				source = target
			}
			arrow.ObjOwner = source
			update := target.UpdateDataPlayer()
			update.Field57, update.Field76, update.Field75 = math.Float32bits(0.25), 88, 77
			armorUD := &WeaponArmorUpdateData{}
			armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), UpdateData: unsafe.Pointer(armorUD), InitData: unsafe.Pointer(&ModifierInitData{}), HealthData: &HealthData{Cur: 10, Max: 10}}
			target.InvFirstItem = armor
			r.ItemArmorValue = func(*Object) float32 { return 0.25 }
			r.CanDamageArmor = func(item *Object) bool { return item == armor }
			wears := 0
			r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
				if item != armor || a != source || w != arrow || d != 2 || typ != object.DamageImpale || update.Field76 != 1 || update.Field75 != 529 {
					t.Fatal("friendly armor callback arguments")
				}
				wears++
				item.HealthData.Cur -= uint16(d)
				return true
			}
			r.CoopMode = func() bool { return tc.coop }
			r.QuestMode = func() bool { return tc.quest }
			tail := DefaultDamageWorldRuntime4E0B30{
				Frame: r.Frame, GameplayFlag1: func() bool { return false },
				QuestMode: r.QuestMode, IsEnemy: func(*Object, *Object) bool { return false },
				BuffOff: func(*Object, EnchantID) {}, PlayerSetState: func(*Object, PlayerState) bool { t.Fatal("small hit requested hurt state"); return false },
				PlayerDamageSoundC: target.DamageSound, PlayerDamageSound: func(*Object, *Object) {}, DamageClear: r.DamageClear,
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				return DefaultDamageWorld4E0B30(v, a, w, d, typ, tail)
			}
			before, beforeUpdate, beforeArmor := *target, *update, *armor
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, 8, object.DamageImpale, r); !h || result != tc.wantResult {
				t.Fatalf("friendly PIERCE = %t/%t", h, result)
			}
			if tc.coop {
				if *target != before || *update != beforeUpdate || *armor != beforeArmor || armor.HealthData.Cur != 10 || wears != 0 || len(*damages) != 0 {
					t.Fatal("Coop self gate ran armor/HP prefix")
				}
			} else {
				wantHP := uint16(20)
				if tc.wantHP {
					wantHP = 14
				}
				if target.HealthData.Cur != wantHP || wears != 1 || armor.HealthData.Cur != 8 || update.Field76 != 1 || update.Field75 != 529 || len(*damages) != map[bool]int{false: 0, true: 1}[tc.wantHP] {
					t.Fatalf("friendly HP=%d armor=%d wears=%d damage=%v marker=%d/%d", target.HealthData.Cur, armor.HealthData.Cur, wears, *damages, update.Field76, update.Field75)
				}
			}
		})
	}
}
