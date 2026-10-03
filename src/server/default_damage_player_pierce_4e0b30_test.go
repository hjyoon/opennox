package server

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultDamagePlayerPierceFixture4E0B30(t *testing.T) (*Object, *Object, *Object, DefaultDamageWorldRuntime4E0B30) {
	t.Helper()
	v, _, w, r := defaultDamagePierceFixture4E0B30(t)
	a, _, _ := playerDamageFixture4E17B0(t)
	a.PrevPos = types.Ptf(56, 78)
	a.DamageSound = r.PlayerDamageSoundC
	// 004E1008 queries a hit-sound set only for a MONSTER source. A
	// player-fired missile must not depend on that unrelated service.
	r.MonsterHasHitSound = nil
	return v, a, w, r
}

func TestDefaultDamageWorld4E0B30PlayerMissilePierce(t *testing.T) {
	for _, targetKind := range []string{"player", "monster", "NPC"} {
		for _, pure := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/pure-%t", targetKind, pure), func(t *testing.T) {
				v, a, w, r := defaultDamagePlayerPierceFixture4E0B30(t)
				if targetKind != "player" {
					v = monsterActionTestObject50A910(t)
					v.ObjClass, v.ObjSubClass = object.ClassMonster, 0
					v.HealthData = &HealthData{Cur: 20, Max: 20}
					if targetKind == "NPC" {
						v.ObjSubClass = 0x10
					}
				}
				if pure {
					w.ObjClass = object.ClassMissile
				}
				w.ObjOwner = a
				v.Buffs = 1 << defaultDamageShockEnchant4E0B30
				beforeSource, beforeWeapon, beforeSourceUD := *a, *w, *a.UpdateDataPlayer()
				wantPos := a.PrevPos
				if pure {
					wantPos = w.PrevPos
				}
				for _, obj := range []*Object{v, a, w} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
						t.Fatalf("object %p is not a native high pointer", obj)
					}
				}
				var events []string
				r.CallDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
					t.Fatal("ranged missile triggered melee Shock")
					return false
				}
				r.FireProtection = func(*Object) float64 { t.Fatal("PIERCE used fire protection"); return 0 }
				r.ElectricProtection = func(*Object) float64 { t.Fatal("PIERCE used electric protection"); return 0 }
				r.BuffOff = func(got *Object, e EnchantID) {
					if got != v || e != 0 || v.Pos132 != wantPos {
						t.Fatal("missile hit position/invisibility order")
					}
					events = append(events, "buff-off")
				}
				sound := func(got, attack *Object) {
					if got != v || attack != w || v.Obj130 != w || v.Field131 != 3 || v.Frame134 != 700 {
						t.Fatal("sound/attribution order")
					}
					events = append(events, "sound")
				}
				r.DefaultDamageSound, r.PlayerDamageSound = sound, sound
				if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageImpale, r) || v.HealthData.Cur != 17 ||
					!reflect.DeepEqual(events, []string{"buff-off", "sound"}) || *a != beforeSource || *w != beforeWeapon ||
					*a.UpdateDataPlayer() != beforeSourceUD || !v.HasEnchant(defaultDamageShockEnchant4E0B30) {
					t.Fatalf("events=%v HP=%d source/weapon/buff changed", events, v.HealthData.Cur)
				}
				if targetKind != "player" {
					ud := v.UpdateDataMonster()
					if ud.Field547 != 1 || ud.Field546 != uint32(w.TypeInd) || !ud.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("missile/injured marker")
					}
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerPierceOrderedNPCTail(t *testing.T) {
	for _, absorb := range []bool{false, true} {
		t.Run(fmt.Sprintf("Shield-absorbs-%t", absorb), func(t *testing.T) {
			_, a, w, r := defaultDamagePlayerPierceFixture4E0B30(t)
			v := monsterActionTestObject50A910(t)
			v.ObjClass, v.ObjSubClass, v.HealthData = object.ClassMonster, 0x10, &HealthData{Cur: 100, Max: 100}
			ud := v.UpdateDataMonster()
			v.Buffs = 1<<defaultDamageShieldEnchant4E0B30 | 1<<defaultDamageShockEnchant4E0B30
			a.Buffs = 1 << damageVampirismEnchant4E0B30
			owner := monsterActionTestObject50A910(t)
			owner.ObjClass, a.ObjOwner = object.ClassMonster, owner
			defend := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			pre := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
				InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, defend, nil}})}
			v.InvFirstItem = armor
			w.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{pre, nil, nil, pre}})
			var events []string
			r.BuffOff = func(got *Object, e EnchantID) {
				if got != v || e != 0 || ud.Field547 != 1 || ud.Field546 != uint32(w.TypeInd) {
					t.Fatal("missile marker before invisibility")
				}
				events = append(events, "buff-off")
			}
			r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == defend }
			r.ApplyLateDefend = func(m *ModifierEff, item, target, weapon, source *Object, d int32, typ object.DamageType) int32 {
				if m != defend || item != armor || target != v || weapon != w || source != a || d != 29 || typ != object.DamageImpale {
					t.Fatal("late defend arguments")
				}
				events = append(events, "defend")
				ud.Field547 = 0 // 004E0FC9 restores the raw type after this callback.
				return d + 2
			}
			r.CanApplyPreDamage = func(m *ModifierEff) bool { return m == pre }
			r.ApplyPreDamage = func(m *ModifierEff, weapon, source, target *Object, d *int32) {
				if m != pre || weapon != w || source != a || target != v || ud.Field547 != 2 || ud.Field546 != 3 || v.Obj130 != w {
					t.Fatal("pre-damage attribution/fallback order")
				}
				events = append(events, "pre")
				*d++
			}
			r.DefaultDamageSound = func(target, weapon *Object) {
				if target != v || weapon != w {
					t.Fatal("damage sound")
				}
				events = append(events, "sound")
			}
			r.Audio = func(id int, got *Object) {
				if id != 163 || got != w {
					t.Fatal("Vampirism audio")
				}
				events = append(events, "vamp-audio")
			}
			r.BalanceFloatInd = func(key string, index int) float64 {
				if key != "VampirismCoeff" || index != -1 {
					t.Fatal("Vampirism balance")
				}
				return 0.5
			}
			r.AdjustHP = func(got *Object, d int32) {
				if got != a || d != 16 {
					t.Fatalf("heal target=%p amount=%d", got, d)
				}
				events = append(events, "heal")
			}
			r.VampirismFX = func(_ int, _, _ image.Point, d uint16) {
				if d != 16 {
					t.Fatal("heal FX")
				}
				events = append(events, "vamp-fx")
			}
			r.AdjustFieldGuide = func(source, target *Object, d int32) int32 {
				if source != a || target != v || d != 33 || owner.UpdateDataMonster().Field130 != 0 {
					t.Fatal("field guide before owner latch")
				}
				events = append(events, "guide")
				return 7
			}
			r.ShieldReduce = func(target *Object, d *int32, typ object.DamageType, weapon *Object) {
				if target != v || *d != 7 || typ != object.DamageImpale || weapon != w || owner.UpdateDataMonster().Field130 != 700 {
					t.Fatal("Shield order/actual missile")
				}
				events = append(events, "shield")
				if absorb {
					*d = 0
				} else {
					*d = 5
				}
			}
			r.DamageClear = func(got *Object, d int32) {
				if got != v || d != 5 {
					t.Fatal("final HP")
				}
				events = append(events, "damage")
				v.HealthData.Cur -= uint16(d)
			}
			result := DefaultDamageWorld4E0B30(v, a, w, 29, object.DamageImpale, r)
			want := []string{"buff-off", "defend", "pre", "pre", "sound", "vamp-audio", "heal", "vamp-fx", "guide", "shield"}
			wantHP := uint16(100)
			if !absorb {
				want = append(want, "damage")
				wantHP = 95
			}
			if result == absorb || !reflect.DeepEqual(events, want) || v.HealthData.Cur != wantHP || !v.HasEnchant(defaultDamageShockEnchant4E0B30) {
				t.Fatalf("result=%t events=%v HP=%d", result, events, v.HealthData.Cur)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerPierceCampaignGates(t *testing.T) {
	for _, tc := range []struct {
		name                               string
		gameplay, enemy, quest, self, want bool
	}{
		{"enemy campaign", false, true, false, false, true},
		{"friendly campaign", false, false, false, false, false},
		{"friendly Quest", false, false, true, false, false},
		{"friendly gameplay flag", true, false, true, false, true},
		{"self campaign", false, false, false, true, true},
		{"self Quest", false, false, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, a, w, r := defaultDamagePlayerPierceFixture4E0B30(t)
			if tc.self {
				v = a
			}
			r.GameplayFlag1 = func() bool { return tc.gameplay }
			r.QuestMode = func() bool { return tc.quest }
			r.IsEnemy = func(target, owner *Object) bool {
				if target != v || owner != a {
					t.Fatal("campaign owner predicate")
				}
				return tc.enemy
			}
			before := *v
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageImpale, r) || (v.HealthData.Cur == 17) != tc.want {
				t.Fatalf("HP=%d want damage=%t", v.HealthData.Cur, tc.want)
			}
			if !tc.want && *v != before {
				t.Fatal("friendly hit mutated target")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerPierceMissingServicesBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name string
		omit func(*Object, *DefaultDamageWorldRuntime4E0B30)
	}{
		{"invisibility", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.BuffOff = nil }},
		{"enemy", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.IsEnemy = nil }},
		{"HP", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.DamageClear = nil }},
		{"hurt state", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.PlayerSetState = nil }},
		{"player sound", func(_ *Object, r *DefaultDamageWorldRuntime4E0B30) { r.PlayerDamageSound = nil }},
		{"Shield", func(v *Object, _ *DefaultDamageWorldRuntime4E0B30) { v.Buffs = 1 << defaultDamageShieldEnchant4E0B30 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, a, w, r := defaultDamagePlayerPierceFixture4E0B30(t)
			tc.omit(v, &r)
			before, beforeUpdate := *v, *v.UpdateDataPlayer()
			reason := ""
			r.Unsupported = func(got string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = got }
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageImpale, r) || reason == "" ||
				*v != before || *v.UpdateDataPlayer() != beforeUpdate || v.HealthData.Cur != 20 || a.HealthData.Cur != 20 {
				t.Fatalf("missing service reason=%q target/source mutated", reason)
			}
		})
	}
}
