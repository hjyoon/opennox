package server

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultDamagePierceFixture4E0B30(t *testing.T) (*Object, *Object, *Object, DefaultDamageWorldRuntime4E0B30) {
	t.Helper()
	target, source, sound := playerDamageFixture4E17B0(t)
	source.PrevPos = types.Ptf(56, 78)
	// Observed from the immutable stock definition during actual stage-5 AI:
	// MISSILE|WEAPON|COMPLEX|NOT_STACKABLE, not a synthetic pure missile.
	arrow := &Object{ObjClass: object.Class(0x05200001), ObjSubClass: 16, TypeInd: 529, PrevPos: types.Ptf(12, 34), InitData: unsafe.Pointer(&ModifierInitData{})}
	r := DefaultDamageWorldRuntime4E0B30{
		Frame: func() uint32 { return 700 }, GameplayFlag1: func() bool { return true },
		IsEnemy: func(*Object, *Object) bool { return true }, BuffOff: func(*Object, EnchantID) {},
		MonsterHasHitSound: func(*Object) bool { return false },
		PlayerDamageSoundC: sound, PlayerDamageSound: func(*Object, *Object) {},
		PlayerSetState: func(*Object, PlayerState) bool { return true },
		DamageClear:    func(v *Object, damage int32) { v.HealthData.Cur -= uint16(damage) },
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("missile PIERCE rejected: %s", reason)
		},
	}
	return target, source, arrow, r
}

func TestDefaultDamageWorld4E0B30MonsterMissilePierce(t *testing.T) {
	for _, npc := range []bool{false, true} {
		t.Run(map[bool]string{false: "player", true: "scripted monster"}[npc], func(t *testing.T) {
			target, source, arrow, r := defaultDamagePierceFixture4E0B30(t)
			if npc {
				target = monsterActionTestObject50A910(t)
				target.ObjClass, target.ObjSubClass = object.ClassMonster, 0x10
				target.HealthData = &HealthData{Cur: 20, Max: 20}
			} else {
				target.UpdateDataPlayer().Field75, target.UpdateDataPlayer().Field76 = 529, 1
				target.UpdateDataPlayer().Field40_0 = 0x1234
			}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, obj := range []*Object{target, source, arrow} {
					if uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
						t.Fatalf("object %p is not a native high pointer", obj)
					}
				}
			}
			var events []string
			r.FireProtection = func(*Object) float64 { t.Fatal("PIERCE used fire protection"); return 0 }
			r.ElectricProtection = func(*Object) float64 { t.Fatal("PIERCE used electric protection"); return 0 }
			r.BuffOff = func(v *Object, enchant EnchantID) {
				if v != target || enchant != 0 || v.Pos132 != source.PrevPos {
					t.Fatal("hit position/invisibility order")
				}
				events = append(events, "buff-off")
			}
			r.MonsterHasHitSound = func(got *Object) bool {
				if got != source {
					t.Fatal("hit-sound source")
				}
				events = append(events, "hit-sound")
				return false
			}
			sound := func(v, a *Object) {
				if v != target || a != arrow || v.Obj130 != arrow || v.Field131 != 3 || v.Frame134 != 700 {
					t.Fatal("damage sound/attribution order")
				}
				events = append(events, "sound")
			}
			r.PlayerDamageSound, r.DefaultDamageSound = sound, sound
			r.DamageClear = func(v *Object, damage int32) {
				if v != target || damage != 3 || source.UpdateDataMonster().Field130 != 700 {
					t.Fatal("HP/attacker latch order")
				}
				v.HealthData.Cur -= uint16(damage)
				events = append(events, "damage")
			}
			if !DefaultDamageWorld4E0B30(target, source, arrow, 3, object.DamageImpale, r) {
				t.Fatal("hit returned false")
			}
			if !reflect.DeepEqual(events, []string{"buff-off", "hit-sound", "sound", "damage"}) || target.HealthData.Cur != 17 {
				t.Fatalf("events=%v HP=%d", events, target.HealthData.Cur)
			}
			if npc {
				update := target.UpdateDataMonster()
				if update.Field546 != 529 || update.Field547 != 1 || !update.StatusFlags.Has(object.MonStatusInjured) {
					t.Fatal("distinct missile/injured marker")
				}
			} else if update := target.UpdateDataPlayer(); update.Field75 != 529 || update.Field76 != 1 || update.Field40_0 != 0x1234 {
				t.Fatal("default tail overwrote player marker or electric word")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PiercePureMissileStillAdmitted(t *testing.T) {
	target, source, arrow, r := defaultDamagePierceFixture4E0B30(t)
	arrow.ObjClass = object.ClassMissile
	if !DefaultDamageWorld4E0B30(target, source, arrow, 3, object.DamageImpale, r) ||
		target.HealthData.Cur != 17 || target.Pos132 != arrow.PrevPos || target.Obj130 != arrow {
		t.Fatal("pure-missile PIERCE or its previous-position contract regressed")
	}
}

func TestDefaultDamageWorld4E0B30PiercePlayerDefenseTail(t *testing.T) {
	target, source, arrow, r := defaultDamagePierceFixture4E0B30(t)
	target.HealthData.Cur = 100
	old := target.UpdateDataPlayer()
	live := &PlayerUpdateData{Player: old.Player, State: PlayerState13}
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier, nil}})}
	target.InvFirstItem, target.Field129 = armor, &Object{TypeInd: 77, ObjOwner: target}
	target.Buffs = 1<<defaultDamageShieldEnchant4E0B30 | 1<<defaultDamageShockEnchant4E0B30
	source.Buffs = 1 << damageVampirismEnchant4E0B30
	var events []string
	r.CallDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
		t.Fatal("missile triggered melee Shock")
		return false
	}
	r.BuffOff = func(v *Object, e EnchantID) {
		if v != target || e != 0 {
			t.Fatal("PIERCE removed Shock")
		}
		events = append(events, "buff-off")
	}
	r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == modifier }
	r.ApplyLateDefend = func(m *ModifierEff, item, v, w, a *Object, d int32, typ object.DamageType) int32 {
		if m != modifier || item != armor || v != target || w != arrow || a != source || d != 29 || typ != object.DamageImpale {
			t.Fatal("late defend arguments")
		}
		events = append(events, "defend")
		v.UpdateData = unsafe.Pointer(live)
		return d + 2
	}
	r.MonsterHasHitSound = func(*Object) bool { events = append(events, "hit-sound"); return false }
	r.PlayerDamageSound = func(v, w *Object) {
		if v != target || w != arrow || v.Obj130 != arrow {
			t.Fatal("player damage sound")
		}
		events = append(events, "sound")
	}
	r.Audio = func(id int, obj *Object) {
		if id != 163 || obj != arrow {
			t.Fatal("Vampirism audio")
		}
		events = append(events, "vamp-audio")
	}
	r.BalanceFloatInd = func(name string, _ int) float64 {
		if name != "VampirismCoeff" {
			t.Fatal("balance")
		}
		return 0.5
	}
	r.AdjustHP = func(v *Object, d int32) {
		if v != source || d != 16 {
			t.Fatal("Vampirism heal")
		}
		events = append(events, "heal")
	}
	r.VampirismFX = func(_ int, _, _ image.Point, d uint16) {
		if d != 16 {
			t.Fatal("Vampirism FX amount")
		}
		events = append(events, "vamp-fx")
	}
	r.GameBallType = 77
	r.GameBallOnDamage = func(a, v *Object, d int32) {
		if a != source || v != target || d != 31 || v.HealthData.Cur != 100 {
			t.Fatal("GameBall before HP")
		}
		events = append(events, "ball")
	}
	r.PlayerSetState = func(v *Object, state PlayerState) bool {
		if v != target || state != PlayerState30 || v.UpdateDataPlayer() != live {
			t.Fatal("hurt live update reload")
		}
		events = append(events, "hurt")
		live.State = state
		return true
	}
	r.ShieldReduce = func(v *Object, d *int32, typ object.DamageType, w *Object) {
		if v != target || *d != 31 || typ != object.DamageImpale || w != arrow || live.State != PlayerState30 || source.UpdateDataMonster().Field130 != 700 {
			t.Fatal("Shield order/arguments")
		}
		events = append(events, "shield")
		*d = 0
	}
	r.DamageClear = func(*Object, int32) { t.Fatal("absorbed damage reached HP") }
	if DefaultDamageWorld4E0B30(target, source, arrow, 29, object.DamageImpale, r) {
		t.Fatal("absorbed damage returned true")
	}
	want := []string{"buff-off", "defend", "hit-sound", "sound", "vamp-audio", "heal", "vamp-fx", "ball", "hurt", "shield"}
	if !reflect.DeepEqual(events, want) || target.HealthData.Cur != 100 || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
		t.Fatalf("events=%v HP=%d buffs=%#x", events, target.HealthData.Cur, target.Buffs)
	}
}

func TestDefaultDamageWorld4E0B30PierceFriendlyAndHurtGates(t *testing.T) {
	for _, tc := range []struct {
		name                                    string
		owner, gameplay, enemy, quest, hitSound bool
		damage                                  int32
		state                                   PlayerState
		wantDamage, wantHurt, wantSound         bool
	}{
		{"ordinary", false, false, true, false, false, 20, PlayerState13, true, true, true},
		{"below hurt threshold", false, false, true, false, false, 19, PlayerState13, true, false, true},
		{"attacking", false, false, true, false, false, 20, PlayerState1, true, false, true},
		{"special state", false, false, true, false, false, 20, PlayerState15, true, false, true},
		{"source hit sound", false, false, true, false, true, 20, PlayerState13, true, true, false},
		{"friendly fire non-enemy missile is not melee", false, true, false, false, false, 3, PlayerState13, true, false, true},
		{"friendly owner", true, false, false, false, false, 3, PlayerState13, false, false, false},
		{"friendly fire enabled", true, true, false, false, false, 3, PlayerState13, true, false, true},
		{"enemy owner", true, false, true, true, false, 3, PlayerState13, true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, a, w, r := defaultDamagePierceFixture4E0B30(t)
			v.UpdateDataPlayer().State = tc.state
			v.Buffs = 1 << defaultDamageShockEnchant4E0B30
			if tc.owner {
				a.ObjOwner = &Object{ObjClass: object.ClassPlayer}
			}
			r.GameplayFlag1 = func() bool { return tc.gameplay }
			r.QuestMode = func() bool { return tc.quest }
			r.IsEnemy = func(*Object, *Object) bool { return tc.enemy }
			r.MonsterHasHitSound = func(*Object) bool { return tc.hitSound }
			damaged, hurt, sound := false, false, false
			r.DamageClear = func(*Object, int32) { damaged = true }
			r.PlayerSetState = func(*Object, PlayerState) bool { hurt = true; return true }
			r.PlayerDamageSound = func(*Object, *Object) { sound = true }
			if !DefaultDamageWorld4E0B30(v, a, w, tc.damage, object.DamageImpale, r) || damaged != tc.wantDamage || hurt != tc.wantHurt || sound != tc.wantSound {
				t.Fatalf("damage=%t hurt=%t sound=%t", damaged, hurt, sound)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PierceMissingServicesBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name string
		omit func(*Object, *DefaultDamageWorldRuntime4E0B30)
	}{
		{"hit sound", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.MonsterHasHitSound = nil }},
		{"hurt", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.PlayerSetState = nil }},
		{"invisibility", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.BuffOff = nil }},
		{"enemy", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.IsEnemy = nil }},
		{"HP", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.DamageClear = nil }},
		{"player sound", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.PlayerDamageSound = nil }},
		{"Shield", func(v *Object, r *DefaultDamageWorldRuntime4E0B30) {
			v.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			r.ShieldReduce = nil
		}},
		{"GameBall", func(v *Object, r *DefaultDamageWorldRuntime4E0B30) {
			v.Field129 = &Object{TypeInd: 77, ObjOwner: v}
			r.GameBallType = 77
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, a, w, r := defaultDamagePierceFixture4E0B30(t)
			v.Pos132, v.Obj130, v.Frame134, v.Field131 = types.Ptf(98, 76), a, 42, 99
			v.UpdateDataPlayer().Field75, v.UpdateDataPlayer().Field76 = 529, 1
			tc.omit(v, &r)
			before := *v
			beforeUpdate := *v.UpdateDataPlayer()
			reason := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageImpale, r) || reason == "" {
				t.Fatal("missing service was not rejected")
			}
			if *v != before || *v.UpdateDataPlayer() != beforeUpdate || v.HealthData.Cur != 20 || a.UpdateDataMonster().Field130 != 0 {
				t.Fatal("missing service mutated target/source")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PierceAdmissionBoundary(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Object, *Object, *Object)
		typ    object.DamageType
	}{
		{"no monster update", func(_, a, _ *Object) { a.UpdateData = nil }, object.DamageImpale},
		{"non-unit source is a separate port", func(_, a, _ *Object) { a.ObjClass = object.ClassSimple }, object.DamageImpale},
		{"mixed missile melee weapon", func(_, _, w *Object) { w.ObjSubClass = 0 }, object.DamageImpale},
		{"mixed missile wand", func(_, _, w *Object) { w.ObjClass |= object.ClassWand }, object.DamageImpale},
		{"mixed missile monster", func(_, _, w *Object) { w.ObjClass |= object.ClassMonster }, object.DamageImpale},
		{"different type", func(_, _, _ *Object) {}, object.DamageDrain},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, a, w, r := defaultDamagePierceFixture4E0B30(t)
			tc.change(v, a, w)
			reason := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			r.DamageClear = func(*Object, int32) { t.Fatal("unsupported shape reached HP") }
			DefaultDamageWorld4E0B30(v, a, w, 3, tc.typ, r)
			if reason == "" || v.Obj130 != nil || v.Frame134 != 0 {
				t.Fatal("unsupported shape was admitted/mutated")
			}
		})
	}
}
