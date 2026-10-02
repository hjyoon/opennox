package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDefaultDamageWorld4E0B30MonsterSword(t *testing.T) {
	for _, subclass := range []uint32{0x51242, 0x11012} {
		for _, hitSound := range []bool{false, true} {
			t.Run(fmt.Sprintf("subclass-%x/hit-sound-%t", subclass, hitSound), func(t *testing.T) {
				update := &MonsterUpdateData{Field547: 99}
				target := &Object{
					ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(subclass),
					UpdateData: unsafe.Pointer(update), HealthData: &HealthData{Cur: 80, Max: 80},
				}
				sourceUpdate := &MonsterUpdateData{}
				source := &Object{
					ObjClass: object.ClassMonster, ObjSubClass: 0x11012,
					UpdateData: unsafe.Pointer(sourceUpdate), PrevPos: types.Pointf{X: 12.5, Y: -40.25},
				}
				weapon := &Object{
					TypeInd: 1174, ObjClass: object.ClassWeapon, ObjSubClass: 0x100,
					PrevPos: types.Pointf{X: -200, Y: 15},
				}
				modifier := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				init := &ModifierInitData{}
				init.Modifiers[1] = modifier
				weapon.InitData = unsafe.Pointer(init)
				var events []string
				runtime := DefaultDamageWorldRuntime4E0B30{
					Frame:         func() uint32 { return 1401 },
					GameplayFlag1: func() bool { return true },
					IsEnemy: func(gotTarget, gotSource *Object) bool {
						if gotTarget != target || gotSource != source {
							t.Fatalf("IsEnemy(%p,%p)", gotTarget, gotSource)
						}
						return true
					},
					BuffOff: func(got *Object, enchant EnchantID) {
						if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 ||
							update.Field547 != 1 || update.Field546 != uint32(weapon.TypeInd) {
							t.Fatalf("BuffOff(%p,%d), hit latch=%d/%d", got, enchant, update.Field547, update.Field546)
						}
						events = append(events, "buff-off")
					},
					CanApplyPreDamage: func(got *ModifierEff) bool { return got == modifier },
					ApplyPreDamage: func(got *ModifierEff, gotWeapon, gotSource, gotTarget *Object, damage *int32) {
						if got != modifier || gotWeapon != weapon || gotSource != source || gotTarget != target ||
							*damage != 22 || target.Obj130 != weapon || target.Frame134 != 1401 ||
							!update.StatusFlags.Has(object.MonStatusInjured) {
							t.Fatal("invalid pre-damage callback arguments or hit prefix")
						}
						*damage -= 3
						events = append(events, "pre-damage")
					},
					MonsterHasHitSound: func(got *Object) bool {
						if got != source {
							t.Fatalf("MonsterHasHitSound(%p)", got)
						}
						events = append(events, "hit-sound")
						return hitSound
					},
					DefaultDamageSound: func(gotTarget, gotSource *Object) {
						if gotTarget != target || gotSource != weapon || hitSound {
							t.Fatalf("DefaultDamageSound(%p,%p), source hit sound=%t", gotTarget, gotSource, hitSound)
						}
						events = append(events, "sound")
					},
					DamageClear: func(got *Object, damage int32) {
						if got != target || damage != 19 || sourceUpdate.Field130 != 1401 {
							t.Fatalf("DamageClear(%p,%d), source frame=%d", got, damage, sourceUpdate.Field130)
						}
						got.HealthData.Cur -= uint16(damage)
						events = append(events, "damage")
					},
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						t.Fatalf("monster Sword damage rejected: %s", reason)
					},
				}
				if !DefaultDamageWorld4E0B30(target, source, weapon, 22, object.DamageBlade, runtime) {
					t.Fatal("monster Sword damage returned false")
				}
				if target.HealthData.Cur != 61 || target.Pos132 != source.PrevPos || target.Obj130 != weapon ||
					target.Field131 != uint32(object.DamageBlade) || target.Frame134 != 1401 ||
					update.Field547 != 1 || update.Field546 != uint32(weapon.TypeInd) {
					t.Fatalf("hit state: health=%d pos=%v source=%p type=%d frame=%d latch=%d/%d",
						target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134,
						update.Field547, update.Field546)
				}
				want := []string{"buff-off", "pre-damage", "hit-sound"}
				if !hitSound {
					want = append(want, "sound")
				}
				want = append(want, "damage")
				if !slices.Equal(events, want) {
					t.Fatalf("events=%v, want %v", events, want)
				}
				t.Logf("native pointers: target=%p source=%p weapon=%p", target, source, weapon)
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30MonsterSwordFriendlyAndMissingSound(t *testing.T) {
	for _, enemy := range []bool{false, true} {
		t.Run(fmt.Sprintf("enemy-%t", enemy), func(t *testing.T) {
			update := &MonsterUpdateData{Field547: 99}
			target := &Object{
				ObjClass: object.ClassMonster, HealthData: &HealthData{Cur: 80, Max: 80},
				UpdateData: unsafe.Pointer(update),
			}
			source := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&MonsterUpdateData{})}
			weapon := &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x100}
			reason, calls := "", 0
			runtime := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return true },
				IsEnemy:       func(*Object, *Object) bool { return enemy },
				BuffOff:       func(*Object, EnchantID) { calls++ },
				DamageClear:   func(*Object, int32) { calls++ },
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) {
					reason = why
				},
			}
			if !enemy {
				runtime.MonsterHasHitSound = func(*Object) bool { calls++; return false }
			}
			if !DefaultDamageWorld4E0B30(target, source, weapon, 22, object.DamageBlade, runtime) ||
				target.HealthData.Cur != 80 || calls != 0 || update.Field547 != 0 || target.Obj130 != nil {
				t.Fatal("friendly/missing-service branch changed the damage tail")
			}
			want := ""
			if enemy {
				want = "missing monster hit-sound lookup"
			}
			if reason != want {
				t.Fatalf("unsupported reason=%q, want %q", reason, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30MonsterSwordAdmissionBoundary(t *testing.T) {
	for _, tc := range []struct {
		name     string
		class    object.Class
		subclass object.SubClass
		typ      object.DamageType
		noUpdate bool
	}{
		{name: "ranged weapon", class: object.ClassWeapon, subclass: 2, typ: object.DamageBlade},
		{name: "non-blade weapon damage", class: object.ClassWeapon, typ: object.DamageFlame},
		{name: "missing source update", class: object.ClassWeapon, typ: object.DamageBlade, noUpdate: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{
				ObjClass: object.ClassMonster, HealthData: &HealthData{Cur: 80, Max: 80},
				UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
			}
			source := &Object{ObjClass: object.ClassMonster}
			if !tc.noUpdate {
				source.UpdateData = unsafe.Pointer(&MonsterUpdateData{})
			}
			weapon := &Object{ObjClass: tc.class, ObjSubClass: tc.subclass}
			reason := ""
			runtime := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1:      func() bool { return true },
				MonsterHasHitSound: func(*Object) bool { t.Fatal("unported branch looked up a hit sound"); return false },
				DamageClear:        func(*Object, int32) { t.Fatal("unported branch applied damage") },
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) {
					reason = why
				},
			}
			DefaultDamageWorld4E0B30(target, source, weapon, 22, tc.typ, runtime)
			if reason != "unsupported monster damage shape" || target.HealthData.Cur != 80 || target.Obj130 != nil {
				t.Fatalf("unported branch: reason=%q health=%d source=%p", reason, target.HealthData.Cur, target.Obj130)
			}
		})
	}
}
