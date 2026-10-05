package server

import (
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
)

func worldImpaleFixture4E17B0(t *testing.T, owner string) (target, source, hazard *Object) {
	target, source, hazard = spellMissileImpactFixture4E17B0(t, owner)
	hazard.ObjClass, hazard.ObjSubClass = object.ClassDangerous|object.ClassImmobile|object.ClassVisibleEnable, 0
	return
}

func TestDefaultDamagePlayerWorldImpale4E0B30(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, simple := range []bool{false, true} {
			for _, raw := range []int32{1, 0, -3, 33} {
				for _, enemy := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/simple-%t/raw-%d/enemy-%t", owner, simple, raw, enemy), func(t *testing.T) {
						v, a, w := worldImpaleFixture4E17B0(t, owner)
						if simple {
							w.ObjClass = object.ClassObstacle | object.ClassDangerous | object.ClassSimple
						}
						v.Buffs = 1 << 22
						r := damageMeleeWorldRuntime4E0B30(t)
						var events []string
						r.IsEnemy = func(target, source *Object) bool {
							if target != v || source != a {
								t.Fatal("enemy identity")
							}
							events = append(events, "enemy")
							return enemy
						}
						r.FireProtection = func(*Object) float64 { t.Fatal("IMPALE fire protection"); return 0 }
						r.ElectricProtection = func(*Object) float64 { t.Fatal("IMPALE electric protection"); return 0 }
						r.CallDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
							t.Fatal("hazard retaliated Shock")
							return false
						}
						r.BuffOff = func(target *Object, enchant EnchantID) {
							if target != v || enchant != 0 || v.Pos132 != w.PrevPos || v.Obj130 != nil {
								t.Fatal("visibility/position order")
							}
							events = append(events, "buff")
						}
						r.MonsterHasHitSound = func(source *Object) bool {
							if source != a || owner != "NPC" {
								t.Fatal("hit sound identity")
							}
							events = append(events, "lookup")
							return false
						}
						r.Frame = func() uint32 { events = append(events, "frame"); return 1407 }
						r.DefaultDamageSound = func(target, hazard *Object) {
							if target != v || hazard != w || v.Obj130 != w || v.Field131 != 3 || v.Frame134 != 1407 {
								t.Fatal("hit metadata")
							}
							events = append(events, "sound")
						}
						r.PlayerSetState = func(target *Object, state PlayerState) bool {
							if target != v || state != PlayerState30 {
								t.Fatal("hurt state")
							}
							events = append(events, "hurt")
							return true
						}
						r.DamageClear = func(target *Object, amount int32) {
							if target != v || amount != raw || !v.HasEnchant(22) {
								t.Fatal("HP/retaliation")
							}
							events = append(events, "hp")
						}
						if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageImpale, r) {
							t.Fatal("world IMPALE failed")
						}
						want := []string{"enemy", "buff", "frame"}
						if owner == "NPC" {
							want = append(want, "lookup")
						}
						want = append(want, "sound")
						if raw >= 20 {
							want = append(want, "hurt")
						}
						if owner == "NPC" {
							want = append(want, "enemy")
							if enemy {
								want = append(want, "frame")
							}
						}
						want = append(want, "hp")
						if !slices.Equal(events, want) {
							t.Fatalf("events=%v want=%v", events, want)
						}
					})
				}
			}
		}
	}
}

func TestDefaultDamagePlayerWorldImpaleEarly4E0B30(t *testing.T) {
	for _, gate := range []string{"invulnerable", "dead", "campaign-friendly", "no-update", "live-no-update", "live-qualifier"} {
		t.Run(gate, func(t *testing.T) {
			v, a, w := worldImpaleFixture4E17B0(t, "player")
			r := damageMeleeWorldRuntime4E0B30(t)
			queries := 0
			r.IsEnemy = func(*Object, *Object) bool {
				queries++
				if gate == "live-no-update" {
					v.ObjFlags |= object.FlagNoUpdate
				}
				if gate == "live-qualifier" {
					w.ObjClass |= object.ClassWeapon
				}
				return false
			}
			want := 1
			switch gate {
			case "invulnerable":
				v.Buffs |= 1 << 23
				want = 0
			case "dead":
				v.ObjFlags |= object.FlagDead
				want = 0
			case "campaign-friendly":
				r.GameplayFlag1 = func() bool { return false }
			case "no-update":
				v.ObjFlags |= object.FlagNoUpdate
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("early gate reached HP") }
			r.BuffOff = func(*Object, EnchantID) { t.Fatal("early gate reached visibility") }
			if !DefaultDamageWorld4E0B30(v, a, w, 1, object.DamageImpale, r) || queries != want || v.Obj130 != nil {
				t.Fatalf("queries=%d want=%d", queries, want)
			}
		})
	}
}

func TestDefaultDamagePlayerWorldImpaleShield4E0B30(t *testing.T) {
	for _, absorbed := range []bool{false, true} {
		t.Run(fmt.Sprint(absorbed), func(t *testing.T) {
			v, a, w := worldImpaleFixture4E17B0(t, "self")
			v.Buffs = 1<<22 | 1<<26
			r := damageMeleeWorldRuntime4E0B30(t)
			hp, shield := 0, 0
			r.ShieldReduce = func(target *Object, amount *int32, typ object.DamageType, hazard *Object) {
				if target != v || hazard != w || *amount != 4 || typ != object.DamageImpale || v.Obj130 != w {
					t.Fatal("Shield identity/order")
				}
				shield++
				*amount = 2
				if absorbed {
					*amount = 0
				}
			}
			r.DamageClear = func(target *Object, amount int32) {
				if target != v || amount != 2 || shield != 1 {
					t.Fatal("HP before Shield")
				}
				hp++
			}
			got := DefaultDamageWorld4E0B30(v, a, w, 4, object.DamageImpale, r)
			if got == absorbed || shield != 1 || hp != map[bool]int{false: 1, true: 0}[absorbed] || !v.HasEnchant(22) {
				t.Fatal("Shield result")
			}
		})
	}
}

func TestPlayerDamageWorldImpaleShape4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		t.Run(owner, func(t *testing.T) {
			_, a, w := worldImpaleFixture4E17B0(t, owner)
			w.TypeInd = 0
			if !playerDamageWorldImpaleShape4E17B0(a, w, object.DamageImpale) {
				t.Fatal("stock non-SIMPLE/zero type shape")
			}
			for _, class := range []object.Class{object.ClassPlayer, object.ClassMonster, object.ClassWeapon, object.ClassWand, object.ClassMissile} {
				w.ObjClass |= class
				if playerDamageWorldImpaleShape4E17B0(a, w, object.DamageImpale) {
					t.Fatal("separate shape admitted")
				}
				w.ObjClass &^= class
			}
			w.ObjClass &^= object.ClassDangerous
			if playerDamageWorldImpaleShape4E17B0(a, w, object.DamageImpale) || playerDamageWorldImpaleShape4E17B0(nil, w, object.DamageImpale) || playerDamageWorldImpaleShape4E17B0(a, nil, object.DamageImpale) || playerDamageWorldImpaleShape4E17B0(a, w, object.DamageImpact) {
				t.Fatal("unrelated shape admitted")
			}
		})
	}
}
