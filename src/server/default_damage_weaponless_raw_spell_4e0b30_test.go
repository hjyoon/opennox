package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func damageRawSpellUnit4E17B0(t *testing.T, kind string) *Object {
	t.Helper()
	u := damageMeleeUnitFixture4E17B0(t, kind == "player")
	if kind == "monster" {
		u.ObjSubClass = 0x10202 // Stock Troll, including its poison immunity.
	}
	return u
}

func TestDefaultDamageWorld4E0B30WeaponlessRawSpellPairs(t *testing.T) {
	for _, from := range []string{"player", "monster", "NPC"} {
		for _, to := range []string{"player", "monster", "NPC"} {
			if from == to {
				continue
			}
			for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
				t.Run(fmt.Sprintf("%s-to-%s/type-%d", from, to, typ), func(t *testing.T) {
					target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, from)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.FireProtection = func(*Object) float64 { t.Fatal("raw spell used fire protection"); return 1 }
					r.ElectricProtection = func(*Object) float64 { t.Fatal("raw spell used electric protection"); return 1 }
					if !DefaultDamageWorld4E0B30(target, source, nil, 25, typ, r) || target.HealthData.Cur != 175 || source.HealthData.Cur != 200 ||
						target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 {
						t.Fatalf("raw spell HP/source/type/frame=%d/%p/%d/%d", target.HealthData.Cur, target.Obj130, target.Field131, target.Frame134)
					}
					if to != "player" {
						ud := target.UpdateDataMonster()
						if ud.Field547 != 2 || ud.Field546 != uint32(typ) || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) {
							t.Fatalf("raw monster marker=%d/%d flags=%x", ud.Field547, ud.Field546, ud.StatusFlags)
						}
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessRawSpellNilAndSigned(t *testing.T) {
	for _, to := range []string{"player", "monster", "NPC"} {
		for _, sourceKind := range []string{"none", "imaginary"} {
			for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
				for _, damage := range []int32{0, -3, 9} {
					t.Run(fmt.Sprintf("%s/%s/type-%d/raw-%d", to, sourceKind, typ, damage), func(t *testing.T) {
						target := damageRawSpellUnit4E17B0(t, to)
						var source *Object
						if sourceKind == "imaginary" {
							source = &Object{PrevPos: types.Ptf(14, -9)}
						}
						r := damageMeleeWorldRuntime4E0B30(t)
						gotDamage, calls := int32(999), 0
						r.DamageClear = func(got *Object, raw int32) {
							if got != target {
								t.Fatal("wrong target")
							}
							gotDamage, calls = raw, calls+1
						}
						if !DefaultDamageWorld4E0B30(target, source, nil, damage, typ, r) || calls != 1 || gotDamage != damage || target.Obj130 != source {
							t.Fatalf("raw tail calls/damage/source=%d/%d/%p", calls, gotDamage, target.Obj130)
						}
					})
				}
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessRawSpellFriendlyAndImmunity(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
		for _, gameplay := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/gameplay-%t", typ, gameplay), func(t *testing.T) {
				target, source := damageRawSpellUnit4E17B0(t, "monster"), damageRawSpellUnit4E17B0(t, "player")
				target.ObjSubClass |= 0xe00 // Poison/fire/electric immunity is not raw-spell immunity.
				r := damageMeleeWorldRuntime4E0B30(t)
				r.GameplayFlag1 = func() bool { return gameplay }
				r.IsEnemy = func(*Object, *Object) bool { return false }
				DefaultDamageWorld4E0B30(target, source, nil, 9, typ, r)
				want := uint16(200)
				if gameplay {
					want = 191
				}
				if target.HealthData.Cur != want {
					t.Fatalf("friendly raw HP=%d want %d", target.HealthData.Cur, want)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessRawSpellLateDefense(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
			t.Run(fmt.Sprintf("%s/type-%d", to, typ), func(t *testing.T) {
				target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, "monster")
				late := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, late}})}
				target.InvFirstItem, target.Buffs = item, 1<<defaultDamageShieldEnchant4E0B30
				var events []string
				r := damageMeleeWorldRuntime4E0B30(t)
				r.BuffOff = func(got *Object, id EnchantID) {
					if got != target || id != ENCHANT_INVISIBLE {
						t.Fatal("wrong invisibility service")
					}
					events = append(events, "invisible")
				}
				r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == late }
				r.ApplyLateDefend = func(m *ModifierEff, gotItem, gotOwner, gotWeapon, gotSource *Object, raw int32, gotType object.DamageType) int32 {
					if m != late || gotItem != item || gotOwner != target || gotWeapon != nil || gotSource != source || raw != 19 || gotType != typ {
						t.Fatal("wrong late defense context")
					}
					events = append(events, "defend")
					return 11
				}
				r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
				r.ShieldReduce = func(got *Object, raw *int32, gotType object.DamageType, gotSource *Object) {
					if got != target || *raw != 11 || gotType != typ || gotSource != source || target.Obj130 != source {
						t.Fatal("wrong Shield context")
					}
					events = append(events, "shield")
					*raw = 3
				}
				clear := r.DamageClear
				r.DamageClear = func(got *Object, raw int32) { events = append(events, "hp"); clear(got, raw) }
				if !DefaultDamageWorld4E0B30(target, source, nil, 19, typ, r) || target.HealthData.Cur != 197 || !slices.Equal(events, []string{"invisible", "defend", "sound", "shield", "hp"}) {
					t.Fatalf("HP=%d events=%v", target.HealthData.Cur, events)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessRawSpellSelfManaBombShield(t *testing.T) {
	for _, kind := range []string{"player", "monster", "NPC"} {
		t.Run(kind, func(t *testing.T) {
			target := damageRawSpellUnit4E17B0(t, kind)
			target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			r := damageMeleeWorldRuntime4E0B30(t)
			r.ShieldReduce = func(*Object, *int32, object.DamageType, *Object) { t.Fatal("self Mana Bomb reduced by Shield") }
			if !DefaultDamageWorld4E0B30(target, target, nil, 9, object.DamageManaBomb, r) || target.HealthData.Cur != 191 {
				t.Fatalf("self Mana Bomb HP=%d", target.HealthData.Cur)
			}
		})
	}
}
