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

func npcHarpoonFixture4E17B0(t *testing.T) (*Object, *Object, *Object, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, r := npcChargeFixture4E17B0(t)
	bolt := &Object{TypeInd: 66, ObjClass: object.ClassMissile, ObjOwner: source, PrevPos: types.Ptf(20, 7)}
	return target, source, bolt, r
}

// Original case 11 shares 004E1F84 with PIERCE, not half-armor CRUSH.
// HarpoonCollide passes parent Player as source and the distinct bolt as weapon.
func TestPlayerDamageNPCHarpoon4E17B0FractionalCarry(t *testing.T) {
	target, source, bolt, r := npcHarpoonFixture4E17B0(t)
	ud := target.UpdateDataMonster()
	ud.Field518, ud.Field547, ud.Field546 = math.Float32bits(0.25), 99, 77
	beforeSource, beforeUD, beforeBolt := *source, *source.UpdateDataPlayer(), *bolt
	for i, hit := range []struct {
		hp    uint16
		carry float32
	}{{198, 0.25}, {196, 0.5}, {193, -0.25}} {
		if h, result := PlayerDamageNative4E17B0(target, source, bolt, 3, object.DamageImpact, r); !h || !result {
			t.Fatalf("harpoon %d handled/result=%t/%t", i, h, result)
		}
		if target.HealthData.Cur != hit.hp || math.Float32frombits(ud.Field1) != hit.carry ||
			ud.Field547 != 1 || ud.Field546 != uint32(bolt.TypeInd) || target.Obj130 != bolt ||
			target.Pos132 != bolt.PrevPos || target.Field131 != uint32(object.DamageImpact) || target.Frame134 != 1400 {
			t.Fatalf("hit=%d HP=%d carry=%g marker=%d/%d attribution=%p/%d", i,
				target.HealthData.Cur, math.Float32frombits(ud.Field1), ud.Field547, ud.Field546, target.Obj130, target.Field131)
		}
	}
	if *source != beforeSource || *source.UpdateDataPlayer() != beforeUD || *bolt != beforeBolt {
		t.Fatal("NPC damage mutated player or projectile")
	}
}

func TestPlayerDamageNPCHarpoon4E17B0StockRangedClassFlags(t *testing.T) {
	target, source, bolt, r := npcHarpoonFixture4E17B0(t)
	source.ObjClass |= object.ClassComplex | object.ClassLight
	bolt.ObjClass |= object.ClassWeapon | object.ClassComplex | object.ClassNotStackable
	bolt.ObjSubClass = 0x10
	bolt.InitData = unsafe.Pointer(&ModifierInitData{})
	source.PrevPos = types.Ptf(44, 9)
	if h, result := PlayerDamageNative4E17B0(target, source, bolt, 1, object.DamageImpact, r); !h || !result ||
		target.HealthData.Cur != 199 || target.Obj130 != bolt || target.Pos132 != source.PrevPos || target.Field131 != uint32(object.DamageImpact) {
		t.Fatalf("stock HarpoonBolt handled/result=%t/%t HP=%d", h, result, target.HealthData.Cur)
	}
}

func TestPlayerDamageNPCHarpoon4E17B0ArmorQuestShield(t *testing.T) {
	target, source, bolt, r := npcHarpoonFixture4E17B0(t)
	armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0)
	armor.HealthData.Cur, armor.HealthData.Max = 100, 100
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor.InitDataModifier().Modifiers[2] = modifier
	target.Buffs = 1<<ENCHANT_INVISIBLE | 1<<ENCHANT_SHIELD
	var events []string
	damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
	r.DamageArmor = func(item, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
		ud := target.UpdateDataMonster()
		if item != armor || attacker != source || weapon != bolt || damage != 40 || typ != object.DamageImpact ||
			ud.Field547 != 1 || ud.Field546 != uint32(bolt.TypeInd) {
			t.Fatal("harpoon armor identity, absorption or marker")
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
			t.Fatal("harpoon buff removal")
		}
		v.Buffs &^= 1 << enchant
		events = append(events, "invisible-off")
	}
	world.CanApplyLateDefend = func(m *ModifierEff) bool { return m == modifier }
	world.ApplyLateDefend = func(m *ModifierEff, item, victim, weapon, attacker *Object, damage int32, typ object.DamageType) int32 {
		if m != modifier || item != armor || victim != target || weapon != bolt || attacker != source || damage != 20 || typ != object.DamageImpact {
			t.Fatal("harpoon late Defend arguments")
		}
		events = append(events, "late-defend")
		return damage - 4
	}
	world.ShieldReduce = func(v *Object, damage *int32, typ object.DamageType, weapon *Object) {
		if v != target || *damage != 16 || typ != object.DamageImpact || weapon != bolt {
			t.Fatal("harpoon Shield arguments")
		}
		events = append(events, "shield")
		*damage -= 6
	}
	world.DamageClear = func(v *Object, damage int32) {
		if v != target || damage != 10 {
			t.Fatal("harpoon final HP damage")
		}
		events = append(events, "HP")
		v.HealthData.Cur -= uint16(damage)
	}
	r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(v, a, w, d, typ, world)
	}
	if h, result := PlayerDamageNative4E17B0(target, source, bolt, 80, object.DamageImpact, r); !h || !result {
		t.Fatal("NPC harpoon rejected")
	}
	if !slices.Equal(events, []string{"armor", "quest", "invisible-off", "late-defend", "shield", "HP"}) ||
		target.HealthData.Cur != 190 || armor.HealthData.Cur != 60 || target.HasEnchant(ENCHANT_INVISIBLE) {
		t.Fatalf("events=%v HP=%d armor=%d", events, target.HealthData.Cur, armor.HealthData.Cur)
	}
}

func TestPlayerDamageNPCHarpoon4E17B0Gates(t *testing.T) {
	for _, tc := range []struct {
		name                                                          string
		flags                                                         object.Flags
		invulnerable, shield, campaign, enemy, wantDamage, wantResult bool
	}{
		{name: "regular enemy", enemy: true, wantDamage: true, wantResult: true},
		{name: "regular non-enemy missile", wantDamage: true, wantResult: true},
		{name: "campaign enemy", campaign: true, enemy: true, wantDamage: true, wantResult: true},
		{name: "campaign friendly owner gate", campaign: true, wantResult: true},
		{name: "NoUpdate", flags: object.FlagNoUpdate},
		{name: "dead", flags: object.FlagDead},
		{name: "invulnerable", invulnerable: true, wantResult: true},
		{name: "Shield absorbs all", shield: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, bolt, r := npcHarpoonFixture4E17B0(t)
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
			if h, result := PlayerDamageNative4E17B0(target, source, bolt, 20, object.DamageImpact, r); !h || result != tc.wantResult || called != tc.wantDamage {
				t.Fatalf("handled/result/damage=%t/%t/%t", h, result, called)
			}
		})
	}
}

func TestPlayerDamageNPCHarpoon4E17B0ShieldFrontRear(t *testing.T) {
	for _, front := range []bool{false, true} {
		t.Run(fmt.Sprint(front), func(t *testing.T) {
			target, source, bolt, r := npcHarpoonFixture4E17B0(t)
			ud := target.UpdateDataMonster()
			ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
			shield := damageMeleeArmorFixture4E17B0(target, 0.5, 0)
			shield.ObjSubClass = 2
			damageMeleeArmorRuntime4E17B0(&r, shield, 0.5)
			r.BlockSourceExcluded = func(v *Object) bool {
				if v != bolt {
					t.Fatal("shield attack")
				}
				return false
			}
			r.BlockDirection = func(v *Object, pos types.Pointf) bool {
				if v != target || pos != bolt.PrevPos {
					t.Fatal("shield position")
				}
				return front
			}
			var events []string
			r.Audio = func(sound int, v *Object) {
				if sound != 878 || v != target || ud.Field547 != 1 || ud.Field546 != uint32(bolt.TypeInd) {
					t.Fatal("shield marker/audio")
				}
				events = append(events, "audio")
			}
			r.ProjectileReflect = func(v, owner *Object) {
				if v != bolt || owner != target {
					t.Fatal("reflect identity")
				}
				events = append(events, "reflect")
			}
			r.ClearOwner = func(v *Object) {
				if v != bolt {
					t.Fatal("clear owner")
				}
				v.ObjOwner = nil
				events = append(events, "clear")
			}
			r.SetOwner = func(owner, v *Object) {
				if owner != target || v != bolt || bolt.ObjOwner != nil {
					t.Fatal("set owner")
				}
				v.ObjOwner = owner
				events = append(events, "owner")
			}
			r.BlockDamagePercent = func() float64 { return 0.25 }
			r.CanDamageBlockItem = func(v *Object) bool { return v == shield }
			r.DamageBlockItem = func(item, owner, attacker, weapon *Object, wear float32, typ object.DamageType) bool {
				if item != shield || owner != target || attacker != source || weapon != bolt || wear != 2 || typ != object.DamageImpact {
					t.Fatal("shield wear")
				}
				events = append(events, "wear")
				return true
			}
			r.Melee.MonsterPopBlockAction = func(*Object) { t.Fatal("unbroken shield popped block") }
			if h, result := PlayerDamageNative4E17B0(target, source, bolt, 8, object.DamageImpact, r); !h || result == front {
				t.Fatal("NPC harpoon front/rear result")
			}
			wantHP := uint16(196)
			if front {
				wantHP = 200
			}
			if target.HealthData.Cur != wantHP || front && !slices.Equal(events, []string{"audio", "reflect", "clear", "owner", "wear"}) ||
				!front && len(events) != 0 || ud.AIStack[0].Action != uint32(ai.ACTION_BLOCK_ATTACK) {
				t.Fatalf("HP=%d events=%v", target.HealthData.Cur, events)
			}
		})
	}
}

func TestPlayerDamageNPCHarpoon4E17B0SignedAndMissingService(t *testing.T) {
	for _, raw := range []int32{0, -4, 1} {
		t.Run(fmt.Sprint(raw), func(t *testing.T) {
			target, source, bolt, r := npcHarpoonFixture4E17B0(t)
			target.UpdateDataMonster().Field518 = math.Float32bits(1)
			var amount int32
			r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				if v != target || a != source || w != bolt || typ != object.DamageImpact {
					t.Fatal("signed harpoon identity")
				}
				amount = d
				return true
			}
			if h, result := PlayerDamageNative4E17B0(target, source, bolt, raw, object.DamageImpact, r); !h || !result {
				t.Fatal("signed harpoon rejected")
			}
			want := int32(0)
			if raw > 0 {
				want = 1
			}
			if amount != want {
				t.Fatalf("damage=%d want=%d", amount, want)
			}
		})
	}
	target, source, bolt, r := npcHarpoonFixture4E17B0(t)
	before := *target.UpdateDataMonster()
	r.DefaultDamage = nil
	why := ""
	r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
	if h, result := PlayerDamageNative4E17B0(target, source, bolt, 20, object.DamageImpact, r); h || result || why != "missing default damage service" || *target.UpdateDataMonster() != before {
		t.Fatalf("handled/result=%t/%t reason=%q", h, result, why)
	}
}

func TestPlayerDamageNPCHarpoon4E17B0AdjacentShapes(t *testing.T) {
	for _, tc := range []struct {
		name                                      string
		sourceClass, boltClass                    object.Class
		typ                                       object.DamageType
		nilSource, nilBolt, nilUpdate, selfWeapon bool
	}{
		{name: "nil source", boltClass: object.ClassMissile, typ: object.DamageImpact, nilSource: true},
		{name: "nil bolt", sourceClass: object.ClassPlayer, typ: object.DamageImpact, nilBolt: true},
		{name: "uninitialized player", sourceClass: object.ClassPlayer, boltClass: object.ClassMissile, typ: object.DamageImpact, nilUpdate: true},
		{name: "monster source", sourceClass: object.ClassMonster, boltClass: object.ClassMissile, typ: object.DamageImpact},
		{name: "player-monster hybrid", sourceClass: object.ClassPlayer | object.ClassMonster, boltClass: object.ClassMissile, typ: object.DamageImpact},
		{name: "player-weapon hybrid", sourceClass: object.ClassPlayer | object.ClassWeapon, boltClass: object.ClassMissile, typ: object.DamageImpact},
		{name: "player-wand hybrid", sourceClass: object.ClassPlayer | object.ClassWand, boltClass: object.ClassMissile, typ: object.DamageImpact},
		{name: "player-missile hybrid", sourceClass: object.ClassPlayer | object.ClassMissile, boltClass: object.ClassMissile, typ: object.DamageImpact},
		{name: "missile-weapon hybrid", sourceClass: object.ClassPlayer, boltClass: object.ClassMissile | object.ClassWeapon, typ: object.DamageImpact},
		{name: "missile-wand hybrid", sourceClass: object.ClassPlayer, boltClass: object.ClassMissile | object.ClassWand, typ: object.DamageImpact},
		{name: "missile-unit hybrid", sourceClass: object.ClassPlayer, boltClass: object.ClassMissile | object.ClassMonster, typ: object.DamageImpact},
		{name: "self weapon", sourceClass: object.ClassPlayer, typ: object.DamageImpact, selfWeapon: true},
		{name: "ZapRay type", sourceClass: object.ClassPlayer, boltClass: object.ClassMissile, typ: object.DamageAirborneElectric},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, bolt, r := npcHarpoonFixture4E17B0(t)
			source.ObjClass, bolt.ObjClass = tc.sourceClass, tc.boltClass
			if tc.nilUpdate {
				source.UpdateData = nil
			}
			if tc.selfWeapon {
				bolt = source
			}
			if tc.nilSource {
				source = nil
			}
			if tc.nilBolt {
				bolt = nil
			}
			before := *target.UpdateDataMonster()
			why := ""
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
			if tc.selfWeapon {
				// This is not a HarpoonBolt, but it is the original case 11
				// caller shape used by a player's Earthquake. Its restored
				// caster path must now damage the NPC rather than reject it.
				beforeSource, beforePlayer := *source, *source.UpdateDataPlayer()
				godQueries := 0
				r.GodMode = func() bool { godQueries++; return true }
				if h, result := PlayerDamageNative4E17B0(target, source, bolt, 20, tc.typ, r); !h || !result || why != "" ||
					target.HealthData.Cur != 180 || target.Obj130 != source || target.Pos132 != source.PrevPos ||
					target.Field131 != uint32(object.DamageImpact) || target.Frame134 != 1400 || godQueries != 1 {
					t.Fatalf("caster IMPACT handled/result=%t/%t reason=%q HP=%d God queries=%d", h, result, why, target.HealthData.Cur, godQueries)
				}
				before.Field547, before.Field546 = 2, uint32(object.DamageImpact)
				before.StatusFlags |= object.MonStatusInjured
				if *target.UpdateDataMonster() != before || *source != beforeSource || *source.UpdateDataPlayer() != beforePlayer {
					t.Fatal("caster IMPACT changed unrelated NPC/player state")
				}
				return
			}
			if h, result := PlayerDamageNative4E17B0(target, source, bolt, 20, tc.typ, r); h || result || why != "unsupported monster damage shape" ||
				target.HealthData.Cur != 200 || *target.UpdateDataMonster() != before {
				t.Fatalf("handled/result=%t/%t reason=%q", h, result, why)
			}
		})
	}
}
