package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func npcChargeFixture4E17B0(t *testing.T) (*Object, *Object, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
	r := damageMeleeRuntimeFixture4E17B0(t)
	r.GodMode = func() bool { t.Fatal("NPC charge read player GodMode"); return false }
	return target, source, r
}

// PlayerCollide supplies the Warrior as BOTH source and weapon. NPC case 2 at
// 004E1EE8 uses half armor absorption, not a MONSTER-only source restriction.
func TestPlayerDamageNPCBerserker4E17B0FractionalCarry(t *testing.T) {
	target, source, r := npcChargeFixture4E17B0(t)
	ud := target.UpdateDataMonster()
	ud.Field518, ud.Field547, ud.Field546 = math.Float32bits(0.25), 99, 77
	beforeSource, beforePlayer := *source, *source.UpdateDataPlayer()
	for i, hit := range []struct {
		hp    uint16
		carry float32
	}{{197, -0.375}, {195, 0.25}, {192, -0.125}} {
		if h, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageCrush, r); !h || !result {
			t.Fatalf("charge %d handled=%t result=%t", i, h, result)
		}
		if target.HealthData.Cur != hit.hp || math.Float32frombits(ud.Field1) != hit.carry ||
			ud.Field547 != 2 || ud.Field546 != uint32(object.DamageCrush) ||
			target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(object.DamageCrush) || target.Frame134 != 1400 {
			t.Fatalf("charge %d HP=%d carry=%g marker=%d/%d", i, target.HealthData.Cur, math.Float32frombits(ud.Field1), ud.Field547, ud.Field546)
		}
	}
	if *source != beforeSource || *source.UpdateDataPlayer() != beforePlayer {
		t.Fatal("NPC damage mutated the player or used a monster update for the source")
	}
}

func TestPlayerDamageNPCBerserker4E17B0ArmorQuestShield(t *testing.T) {
	target, source, r := npcChargeFixture4E17B0(t)
	armor := damageMeleeArmorFixture4E17B0(target, 0.4, 0.25)
	armor.HealthData.Cur, armor.HealthData.Max = 100, 100
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor.InitDataModifier().Modifiers[2] = modifier
	target.Buffs = 1<<ENCHANT_INVISIBLE | 1<<ENCHANT_SHIELD | 1<<ENCHANT_SHOCK
	var events []string
	damageMeleeArmorRuntime4E17B0(&r, armor, 0.4)
	r.DamageArmor = func(item, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
		if item != armor || attacker != source || weapon != source || damage != 30 || typ != object.DamageCrush {
			t.Fatal("charge armor lost source/self-weapon or absorbed damage")
		}
		events = append(events, "armor")
		item.HealthData.Cur -= uint16(damage)
		return true
	}
	r.QuestMode = func() bool { return true }
	r.QuestDamageScale = func() float32 { events = append(events, "quest"); return 0.5 }
	world := damageMeleeWorldRuntime4E0B30(t)
	world.BuffOff = func(v *Object, enchant EnchantID) {
		if v != target || enchant != ENCHANT_INVISIBLE {
			t.Fatal("wrong charge buff removal")
		}
		v.Buffs &^= 1 << enchant
		events = append(events, "invisible-off")
	}
	world.CanApplyLateDefend = func(m *ModifierEff) bool { return m == modifier }
	world.ApplyLateDefend = func(m *ModifierEff, item, victim, weapon, attacker *Object, damage int32, typ object.DamageType) int32 {
		if m != modifier || item != armor || victim != target || weapon != source || attacker != source || damage != 60 || typ != object.DamageCrush {
			t.Fatal("charge late Defend lost ordered arguments")
		}
		events = append(events, "late-defend")
		return damage - 10
	}
	world.ShieldReduce = func(v *Object, damage *int32, typ object.DamageType, weapon *Object) {
		if v != target || *damage != 50 || typ != object.DamageCrush || weapon != source {
			t.Fatal("charge Shield arguments")
		}
		events = append(events, "shield")
		*damage -= 10
	}
	world.DamageClear = func(v *Object, damage int32) {
		if v != target || damage != 40 {
			t.Fatal("charge final HP damage")
		}
		events = append(events, "HP")
		v.HealthData.Cur -= uint16(damage)
	}
	r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(v, a, w, d, typ, world)
	}
	if h, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, r); !h || !result {
		t.Fatal("NPC charge rejected")
	}
	if !slices.Equal(events, []string{"armor", "quest", "invisible-off", "late-defend", "shield", "HP"}) ||
		target.HealthData.Cur != 160 || armor.HealthData.Cur != 70 || math.Float32frombits(target.UpdateDataMonster().Field1) != 0.25 ||
		!target.HasEnchant(ENCHANT_SHOCK) || target.HasEnchant(ENCHANT_INVISIBLE) {
		t.Fatalf("events=%v HP=%d armor=%d buffs=%#x", events, target.HealthData.Cur, armor.HealthData.Cur, target.Buffs)
	}
}

func TestPlayerDamageNPCBerserker4E17B0Gates(t *testing.T) {
	for _, tc := range []struct {
		name                                                          string
		flags                                                         object.Flags
		invulnerable, shield, campaign, enemy, wantDamage, wantResult bool
	}{
		{name: "regular hostile", enemy: true, wantDamage: true, wantResult: true},
		{name: "regular friendly non-melee", wantDamage: true, wantResult: true},
		{name: "campaign enemy", campaign: true, enemy: true, wantDamage: true, wantResult: true},
		{name: "campaign friendly owner gate", campaign: true, wantResult: true},
		{name: "NoUpdate", flags: object.FlagNoUpdate},
		{name: "dead", flags: object.FlagDead},
		{name: "invulnerable", invulnerable: true, wantResult: true},
		{name: "Shield absorbs all", shield: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, r := npcChargeFixture4E17B0(t)
			target.ObjFlags |= tc.flags
			if tc.invulnerable {
				target.Buffs |= 1 << ENCHANT_INVULNERABLE
			}
			if tc.shield {
				target.Buffs |= 1 << ENCHANT_SHIELD
			}
			world := damageMeleeWorldRuntime4E0B30(t)
			world.GameplayFlag1 = func() bool { return !tc.campaign }
			world.IsEnemy = func(*Object, *Object) bool { return tc.enemy }
			world.ShieldReduce = func(_ *Object, d *int32, _ object.DamageType, _ *Object) { *d = 0 }
			called := false
			world.DamageClear = func(v *Object, d int32) { called = true; v.HealthData.Cur -= uint16(d) }
			r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				return DefaultDamageWorld4E0B30(v, a, w, d, typ, world)
			}
			if h, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, r); !h || result != tc.wantResult || called != tc.wantDamage {
				t.Fatalf("handled/result/damage=%t/%t/%t", h, result, called)
			}
		})
	}
}

func TestPlayerDamageNPCBerserker4E17B0ShieldFrontRear(t *testing.T) {
	for _, front := range []bool{false, true} {
		t.Run(fmt.Sprintf("front-%t", front), func(t *testing.T) {
			target, source, r := npcChargeFixture4E17B0(t)
			ud := target.UpdateDataMonster()
			ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
			shield := damageMeleeArmorFixture4E17B0(target, 0.5, 0.25)
			shield.ObjSubClass = 2
			damageMeleeArmorRuntime4E17B0(&r, shield, 0.5)
			r.BlockSourceExcluded = func(v *Object) bool {
				if v != source {
					t.Fatal("shield attack identity")
				}
				return false
			}
			r.BlockDirection = func(v *Object, pos types.Pointf) bool {
				if v != target || pos != source.PrevPos {
					t.Fatal("shield needs prior attack position")
				}
				return front
			}
			r.Audio = func(sound int, v *Object) {
				if sound != 878 || v != target || ud.Field547 != 0 {
					t.Fatal("charge shield marker/audio")
				}
			}
			r.BlockDamagePercent = func() float64 { return 0.2 }
			r.CanDamageBlockItem = func(v *Object) bool { return v == shield }
			blocked := false
			r.DamageBlockItem = func(item, owner, attacker, weapon *Object, wear float32, typ object.DamageType) bool {
				if item != shield || owner != target || attacker != source || weapon != source || wear != 1.6 || typ != object.DamageCrush {
					t.Fatal("charge shield wear arguments")
				}
				blocked = true
				return true
			}
			r.Melee.MonsterPopBlockAction = func(*Object) { t.Fatal("unbroken shield popped block action") }
			if h, result := PlayerDamageNative4E17B0(target, source, source, 8, object.DamageCrush, r); !h || result == front || blocked != front {
				t.Fatal("NPC charge front/rear result")
			}
			wantHP := uint16(194)
			if front {
				wantHP = 200
			}
			if target.HealthData.Cur != wantHP || ud.AIStack[0].Action != uint32(ai.ACTION_BLOCK_ATTACK) {
				t.Fatal("shield did not preserve HP/action")
			}
		})
	}
}

func TestPlayerDamageNPCBerserker4E17B0SignedAndMissingService(t *testing.T) {
	for _, raw := range []int32{0, -4, 1} {
		t.Run(fmt.Sprint(raw), func(t *testing.T) {
			target, source, r := npcChargeFixture4E17B0(t)
			target.UpdateDataMonster().Field518 = math.Float32bits(1)
			var amount int32
			r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				if v != target || a != source || w != source || typ != object.DamageCrush {
					t.Fatal("signed charge identity")
				}
				amount = d
				return true
			}
			if h, result := PlayerDamageNative4E17B0(target, source, source, raw, object.DamageCrush, r); !h || !result {
				t.Fatal("signed NPC charge rejected")
			}
			want := raw / 2
			if raw > 0 && want == 0 {
				want = 1
			}
			if amount != want {
				t.Fatalf("damage=%d want=%d", amount, want)
			}
		})
	}
	target, source, r := npcChargeFixture4E17B0(t)
	before := *target.UpdateDataMonster()
	r.DefaultDamage = nil
	why := ""
	r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
	if h, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, r); h || result || why != "missing default damage service" || *target.UpdateDataMonster() != before {
		t.Fatalf("missing service handled/result=%t/%t reason=%q", h, result, why)
	}
}

func TestPlayerDamageNPCBerserker4E17B0AdjacentShapes(t *testing.T) {
	for _, tc := range []struct {
		name      string
		class     object.Class
		typ       object.DamageType
		nilUpdate bool
	}{
		{"player Blade", object.ClassPlayer, object.DamageBlade, false},
		{"player PIERCE", object.ClassPlayer, object.DamageImpale, false},
		{"uninitialized player", object.ClassPlayer, object.DamageCrush, true},
		{"player/monster hybrid", object.ClassPlayer | object.ClassMonster, object.DamageCrush, false},
		{"player/weapon hybrid", object.ClassPlayer | object.ClassWeapon, object.DamageCrush, false},
		{"player/wand hybrid", object.ClassPlayer | object.ClassWand, object.DamageCrush, false},
		{"player/missile hybrid", object.ClassPlayer | object.ClassMissile, object.DamageCrush, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, r := npcChargeFixture4E17B0(t)
			source.ObjClass = tc.class
			if tc.nilUpdate {
				source.UpdateData = nil
			}
			before := *target.UpdateDataMonster()
			why := ""
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
			if h, result := PlayerDamageNative4E17B0(target, source, source, 150, tc.typ, r); h || result || why != "unsupported monster damage shape" ||
				target.HealthData.Cur != 200 || *target.UpdateDataMonster() != before {
				t.Fatalf("adjacent shape handled/result=%t/%t reason=%q", h, result, why)
			}
		})
	}
}
