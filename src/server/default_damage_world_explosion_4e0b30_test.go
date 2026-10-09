package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func damageWorldExplosionSource4E17B0(t *testing.T, variant int) *Object {
	t.Helper()
	source := &Object{
		TypeInd: uint16(773 + variant), ObjClass: object.ClassSimple | object.ClassLight,
		ObjFlags: object.FlagActive | object.FlagEnabled | object.FlagNoCollide,
		PrevPos:  types.Ptf(23, -9),
	}
	if source.UpdateData != nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0))) {
		t.Fatal("world explosion must have no AI record and retain its native high address")
	}
	return source
}

func damageWorldExplosionTarget4E17B0(t *testing.T, kind string) *Object {
	t.Helper()
	target := damageMeleeUnitFixture4E17B0(t, kind == "player")
	if kind == "monster" {
		target.ObjSubClass = 0x202
	}
	return target
}

func TestPlayerDamageWorldExplosionShape4E17B0(t *testing.T) {
	for _, class := range []object.Class{object.ClassSimple, object.ClassSimple | object.ClassLight, object.ClassSimple | object.ClassObstacle | object.ClassLight} {
		for _, typ := range []object.DamageType{object.DamageExplosion, object.DamageBlade, object.DamageFlame, object.DamageCrush, object.DamageImpale, object.DamagePoison} {
			source := &Object{ObjClass: class}
			if got := playerDamageWorldExplosionShape4E17B0(source, nil, typ); got != (typ == object.DamageExplosion) {
				t.Fatalf("world class=%x type=%d admitted=%t", class, typ, got)
			}
		}
	}
	for _, class := range []object.Class{0, object.ClassLight, object.ClassObstacle, object.ClassPlayer, object.ClassMonster,
		object.ClassSimple | object.ClassPlayer, object.ClassSimple | object.ClassMonster,
		object.ClassSimple | object.ClassMissile, object.ClassSimple | object.ClassWeapon, object.ClassSimple | object.ClassWand} {
		if playerDamageWorldExplosionShape4E17B0(&Object{ObjClass: class}, nil, object.DamageExplosion) {
			t.Fatalf("non-world/mixed class=%x incorrectly admitted", class)
		}
	}
	if playerDamageWorldExplosionShape4E17B0(nil, nil, object.DamageExplosion) ||
		playerDamageWorldExplosionShape4E17B0(damageWorldExplosionSource4E17B0(t, 0), &Object{}, object.DamageExplosion) {
		t.Fatal("nil source or a distinct weapon entered the world-only predicate")
	}
}

func TestDefaultDamageWorldExplosion4E0B30Matrix(t *testing.T) {
	for _, kind := range []string{"player", "monster", "npc"} {
		for variant := 0; variant < 2; variant++ {
			for _, tc := range []struct {
				name       string
				damage     int32
				protection float64
				immune     bool
				want       int32
			}{
				{"stock", 30, 0, false, 30}, {"resistance", 30, .5, false, 15},
				{"tie-even", 9, .5, false, 4}, {"tie-odd", 11, .5, false, 6},
				{"minimum", 1, 1, false, 1}, {"zero", 0, 0, false, 1}, {"signed", -9, .5, false, -4},
				{"immune", 9, 0, true, 4}, {"immune-resistance", 11, .5, true, 2}, {"immune-signed", -9, .5, true, -2},
			} {
				if tc.immune && kind == "player" {
					continue
				}
				t.Run(fmt.Sprintf("%s/barrel-%d/%s", kind, variant+1, tc.name), func(t *testing.T) {
					target, source := damageWorldExplosionTarget4E17B0(t, kind), damageWorldExplosionSource4E17B0(t, variant)
					if tc.immune {
						target.ObjSubClass |= 0x400
					}
					r := damageMeleeWorldRuntime4E0B30(t)
					r.FireProtection = func(*Object) float64 { return tc.protection }
					r.IsEnemy = func(*Object, *Object) bool {
						t.Fatal("unowned world explosion entered the melee enemy gate")
						return false
					}
					if !DefaultDamageWorld4E0B30(target, source, nil, tc.damage, object.DamageExplosion, r) ||
						target.HealthData.Cur != uint16(200-tc.want) || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
						t.Fatalf("HP=%d/%d source=%p/%p position=%v/%v type/frame=%d/%d", target.HealthData.Cur, 200-tc.want, target.Obj130, source, target.Pos132, source.PrevPos, target.Field131, target.Frame134)
					}
					if kind != "player" {
						ud := target.UpdateDataMonster()
						mask := object.MonStatusInjured | object.MonStatusOnFire
						if ud.Field547 != 2 || ud.Field546 != 7 || ud.StatusFlags&mask != mask {
							t.Fatalf("monster marker/status=%d/%d/%x", ud.Field547, ud.Field546, ud.StatusFlags)
						}
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorldExplosion4E0B30OwnerGate(t *testing.T) {
	for _, kind := range []string{"player", "monster", "npc"} {
		for _, owned := range []bool{false, true} {
			for _, gameplay := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/owned-%t/gameplay-%t", kind, owned, gameplay), func(t *testing.T) {
					target, source := damageWorldExplosionTarget4E17B0(t, kind), damageWorldExplosionSource4E17B0(t, 0)
					if owned {
						source.ObjOwner = damageMeleeUnitFixture4E17B0(t, true)
					}
					r := damageMeleeWorldRuntime4E0B30(t)
					r.GameplayFlag1 = func() bool { return gameplay }
					r.FireProtection = func(*Object) float64 { return 0 }
					queries := 0
					r.IsEnemy = func(v, s *Object) bool {
						if v != target || s != source.ObjOwner {
							t.Fatal("world-source owner gate lost identity")
						}
						queries++
						return false
					}
					if !DefaultDamageWorld4E0B30(target, source, nil, 30, object.DamageExplosion, r) {
						t.Fatal("owner-gated explosion returned false")
					}
					wantHP, wantQueries := uint16(170), 0
					if owned && !gameplay {
						wantHP, wantQueries = 200, 1
					}
					if target.HealthData.Cur != wantHP || queries != wantQueries {
						t.Fatalf("HP=%d/%d queries=%d/%d", target.HealthData.Cur, wantHP, queries, wantQueries)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorldExplosion4E0B30OrderShieldAndNoShock(t *testing.T) {
	for _, kind := range []string{"player", "monster", "npc"} {
		t.Run(kind, func(t *testing.T) {
			target, source := damageWorldExplosionTarget4E17B0(t, kind), damageWorldExplosionSource4E17B0(t, 0)
			target.Buffs = 1<<defaultDamageShieldEnchant4E0B30 | 1<<defaultDamageShockEnchant4E0B30
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.FireProtection = func(*Object) float64 { events = append(events, "fire"); return .5 }
			r.Audio = func(id int, v *Object) {
				if id != 104 || v != target {
					t.Fatal("unexpected protection audio")
				}
				events = append(events, "fire-sound")
			}
			r.BuffOff = func(v *Object, e EnchantID) {
				if v != target || e != defaultDamageInvisibleEnchant4E0B30 || v.Pos132 != source.PrevPos {
					t.Fatal("BuffOff before original position store")
				}
				events = append(events, "buff")
			}
			r.DefaultDamageSound = func(v, s *Object) {
				if v != target || s != source || v.Obj130 != source || v.Frame134 != 1400 {
					t.Fatal("sound before attribution")
				}
				events = append(events, "sound")
			}
			r.ShieldReduce = func(v *Object, d *int32, typ object.DamageType, s *Object) {
				if v != target || *d != 15 || typ != object.DamageExplosion || s != source {
					t.Fatal("Shield lost protected amount or barrel identity")
				}
				events = append(events, "shield")
				*d = 0
			}
			r.CallDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("nil-weapon barrel explosion triggered Shock")
				return false
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("zero Shield result reached HP") }
			if DefaultDamageWorld4E0B30(target, source, nil, 30, object.DamageExplosion, r) || target.HealthData.Cur != 200 || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
				t.Fatal("Shield/ Shock contract changed")
			}
			if !slices.Equal(events, []string{"fire", "fire-sound", "buff", "sound", "shield"}) {
				t.Fatalf("events=%v", events)
			}
		})
	}
}
