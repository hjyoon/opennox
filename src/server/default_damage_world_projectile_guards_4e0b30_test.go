package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
)

func TestWorldProjectileShape4E17B0Bounds(t *testing.T) {
	for _, tc := range worldProjectileCases4E0B30() {
		t.Run(tc.name, func(t *testing.T) {
			a, w, m := worldProjectileFixture4E0B30(tc)
			if !playerDamageWorldProjectileShape4E17B0(a, w, tc.typ) {
				t.Fatal("stock shape rejected")
			}
			for _, bad := range []object.Class{object.ClassPlayer, object.ClassMonster, object.ClassWand} {
				before := m.ObjClass
				m.ObjClass |= bad
				if playerDamageWorldProjectileShape4E17B0(a, w, tc.typ) {
					t.Fatalf("mixed missile class %v admitted", bad)
				}
				m.ObjClass = before
			}
			if playerDamageWorldProjectileShape4E17B0(a, w, object.DamageBite) || playerDamageWorldProjectileShape4E17B0(nil, w, tc.typ) {
				t.Fatal("unrelated shape admitted")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30WorldProjectileEnemyBeforeNoUpdate(t *testing.T) {
	for _, tc := range worldProjectileCases4E0B30() {
		t.Run(tc.name, func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, true)
			v.ObjFlags |= object.FlagNoUpdate
			a, w, _ := worldProjectileFixture4E0B30(tc)
			r := damageMeleeWorldRuntime4E0B30(t)
			queries := 0
			r.IsEnemy = func(got, owner *Object) bool {
				if got != v || owner != a {
					t.Fatal("wrong enemy inputs")
				}
				queries++
				return false
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("NoUpdate reached HP") }
			if !DefaultDamageWorld4E0B30(v, a, w, 25, tc.typ, r) {
				t.Fatal("NoUpdate failed")
			}
			want := 1
			if tc.radial {
				want = 0
			}
			if queries != want {
				t.Fatalf("enemy queries=%d want=%d", queries, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30WorldProjectileLiveFriendlyQualifier(t *testing.T) {
	for _, qualifying := range []bool{false, true} {
		t.Run(fmt.Sprintf("qualifying-%t", qualifying), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			a, w, _ := worldProjectileFixture4E0B30(worldProjectileCases4E0B30()[0])
			r := damageMeleeWorldRuntime4E0B30(t)
			queries := 0
			r.IsEnemy = func(*Object, *Object) bool {
				queries++
				if qualifying {
					w.ObjSubClass = 0
				}
				return false
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 25, object.DamageImpale, r) {
				t.Fatal("friendly query failed")
			}
			want := uint16(175)
			if qualifying {
				want = 200
			}
			if queries != 1 || v.HealthData.Cur != want {
				t.Fatalf("queries=%d HP=%d want=%d", queries, v.HealthData.Cur, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30WorldProjectileFireImmunityProtectionShield(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, immune := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-%t/immune-%t", player, immune), func(t *testing.T) {
				v := damageMeleeUnitFixture4E17B0(t, player)
				if immune {
					v.ObjSubClass |= 0x400
				}
				v.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
				a, w, _ := worldProjectileFixture4E0B30(worldProjectileCases4E0B30()[2])
				r := damageMeleeWorldRuntime4E0B30(t)
				r.IsEnemy = func(*Object, *Object) bool { return false }
				protected, shielded := 0, 0
				r.FireProtection = func(*Object) float64 { protected++; return 0.5 }
				r.ShieldReduce = func(got *Object, d *int32, k object.DamageType, effective *Object) {
					if got != v || effective != w || k != object.DamageFlame || *d != 12 {
						t.Fatalf("Shield inputs: damage=%d", *d)
					}
					shielded++
					*d -= 2
				}
				if !DefaultDamageWorld4E0B30(v, a, w, 25, object.DamageFlame, r) {
					t.Fatal("fireball failed")
				}
				want := uint16(190)
				calls := 1
				if immune && !player {
					want = 200
					calls = 0
				}
				if v.HealthData.Cur != want || protected != calls || shielded != calls {
					t.Fatalf("HP=%d protection=%d Shield=%d", v.HealthData.Cur, protected, shielded)
				}
			})
		}
	}
	t.Run("immune-before-health-and-tail-services", func(t *testing.T) {
		v := damageMeleeUnitFixture4E17B0(t, false)
		v.ObjSubClass |= 0x400
		v.HealthData = nil
		a, w, _ := worldProjectileFixture4E0B30(worldProjectileCases4E0B30()[2])
		r := damageMeleeWorldRuntime4E0B30(t)
		r.BuffOff = nil
		r.DamageClear = nil
		r.FireProtection = nil
		if !DefaultDamageWorld4E0B30(v, a, w, 25, object.DamageFlame, r) {
			t.Fatal("immune hit rejected")
		}
	})
}

func TestDefaultDamageWorld4E0B30DeathBallTerminalOwnerGate(t *testing.T) {
	for _, radial := range []bool{false, true} {
		for _, kind := range []string{"World", "Player", "NPC", "Imaginary", "Unowned"} {
			for _, gameplay := range []bool{false, true} {
				t.Run(fmt.Sprintf("radial-%t/%s/gameplay-%t", radial, kind, gameplay), func(t *testing.T) {
					v := damageMeleeUnitFixture4E17B0(t, true)
					tc := worldProjectileCases4E0B30()[3]
					tc.radial = radial
					a, w, m := worldProjectileFixture4E0B30(tc)
					owner := m.ObjOwner
					if kind == "Player" || kind == "NPC" {
						owner = damageMeleeUnitFixture4E17B0(t, kind == "Player")
					}
					if kind == "Imaginary" {
						owner = &Object{}
					}
					if kind == "Unowned" {
						m.ObjOwner = nil
						owner = m
					} else {
						m.ObjOwner = owner
					}
					if !radial {
						a = owner
					}
					if !playerDamageWorldProjectileShape4E17B0(a, w, object.DamageCrush) {
						t.Fatal("DeathBall owner shape rejected")
					}
					r := damageMeleeWorldRuntime4E0B30(t)
					r.GameplayFlag1 = func() bool { return gameplay }
					r.IsEnemy = func(*Object, *Object) bool { return false }
					if !DefaultDamageWorld4E0B30(v, a, w, 25, object.DamageCrush, r) {
						t.Fatal("DeathBall failed")
					}
					want := uint16(175)
					if !gameplay && (kind == "Player" || kind == "NPC") {
						want = 200
					}
					if v.HealthData.Cur != want {
						t.Fatalf("HP=%d want=%d", v.HealthData.Cur, want)
					}
				})
			}
		}
	}
}
