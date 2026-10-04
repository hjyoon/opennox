package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func spellMissileImpactFixture4E17B0(t *testing.T, owner string) (target, source, missile *Object) {
	t.Helper()
	target = damageMeleeUnitFixture4E17B0(t, true)
	// Stock Pixie class=0x180001/subclass=3. No update is needed to
	// identify an unowned projectile's terminal-parent/self-weapon pair.
	missile = &Object{TypeInd: 991, ObjClass: object.ClassMissile | object.ClassSimple | object.ClassLight,
		ObjSubClass: 3, PrevPos: types.Ptf(17, 9), PosVec: types.Ptf(19, 11)}
	source = missile
	if owner != "self" {
		source = damageMeleeUnitFixture4E17B0(t, owner == "player")
		missile.ObjOwner = source
	}
	return
}

func TestDefaultDamagePlayerSpellMissileImpact4E0B30(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, raw := range []int32{8, 0, -3, 33} {
			for _, enemy := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/raw-%d/enemy-%t", owner, raw, enemy), func(t *testing.T) {
					v, a, w := spellMissileImpactFixture4E17B0(t, owner)
					r := damageMeleeWorldRuntime4E0B30(t)
					var events []string
					r.IsEnemy = func(target, source *Object) bool {
						if target != v || source != a {
							t.Fatal("enemy arguments changed")
						}
						events = append(events, "enemy")
						return enemy
					}
					r.BuffOff = func(target *Object, enchant EnchantID) {
						if target != v || enchant != 0 || v.Pos132 != w.PrevPos || v.Obj130 != nil {
							t.Fatal("position/visibility ordering changed")
						}
						events = append(events, "buff")
					}
					r.MonsterHasHitSound = func(source *Object) bool {
						if source != a || owner != "NPC" {
							t.Fatal("monster sound lookup")
						}
						events = append(events, "lookup")
						return false
					}
					r.DefaultDamageSound = func(target, weapon *Object) {
						if target != v || weapon != w || v.Obj130 != w || v.Field131 != 11 || v.Frame134 != 1400 {
							t.Fatal("attribution/sound identity")
						}
						events = append(events, "sound")
					}
					r.GameBallOnDamage = func(source, target *Object, amount int32) {
						if source != a || target != v || amount != raw {
							t.Fatal("GameBall arguments")
						}
						events = append(events, "ball")
					}
					r.PlayerSetState = func(target *Object, state PlayerState) bool {
						if target != v || state != PlayerState30 {
							t.Fatal("hurt state")
						}
						events = append(events, "hurt")
						return true
					}
					r.DamageClear = func(target *Object, amount int32) {
						if target != v || amount != raw {
							t.Fatal("raw signed damage changed")
						}
						if owner == "NPC" && enemy && a.UpdateDataMonster().Field130 != 1400 {
							t.Fatal("source combat latch not stored before HP")
						}
						events = append(events, "hp")
					}
					r.FireProtection = func(*Object) float64 { t.Fatal("IMPACT used fire protection"); return 0 }
					r.ElectricProtection = func(*Object) float64 { t.Fatal("IMPACT used electric protection"); return 0 }
					if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageImpact, r) {
						t.Fatal("unshielded impact failed")
					}
					want := []string{"enemy", "buff"}
					if owner == "NPC" {
						want = append(want, "lookup")
					}
					want = append(want, "sound", "ball")
					if raw >= 20 {
						want = append(want, "hurt")
					}
					if owner == "NPC" {
						want = append(want, "enemy")
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

func TestDefaultDamagePlayerSpellMissileImpactEarlyGates4E0B30(t *testing.T) {
	for _, gate := range []string{"invulnerable", "dead", "campaign-friendly", "no-update", "enemy-callback-no-update"} {
		t.Run(gate, func(t *testing.T) {
			v, a, w := spellMissileImpactFixture4E17B0(t, "player")
			r := damageMeleeWorldRuntime4E0B30(t)
			queries := 0
			r.IsEnemy = func(*Object, *Object) bool {
				queries++
				if gate == "enemy-callback-no-update" {
					v.ObjFlags |= object.FlagNoUpdate
				}
				return false
			}
			wantQueries := 0
			switch gate {
			case "invulnerable":
				v.Buffs = 1 << 23
			case "dead":
				v.ObjFlags |= object.FlagDead
			case "campaign-friendly":
				r.GameplayFlag1 = func() bool { return false }
				wantQueries = 1
			case "no-update":
				v.ObjFlags |= object.FlagNoUpdate
				wantQueries = 1
			case "enemy-callback-no-update":
				wantQueries = 1
			}
			r.BuffOff = func(*Object, EnchantID) { t.Fatal("early gate reached BuffOff") }
			r.DamageClear = func(*Object, int32) { t.Fatal("early gate reached HP") }
			if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) || queries != wantQueries || v.Obj130 != nil || v.HealthData.Cur != 200 {
				t.Fatalf("queries=%d want=%d", queries, wantQueries)
			}
		})
	}
}

func TestDefaultDamagePlayerSpellMissileImpactShield4E0B30(t *testing.T) {
	for _, absorbed := range []bool{false, true} {
		t.Run(fmt.Sprint(absorbed), func(t *testing.T) {
			v, a, w := spellMissileImpactFixture4E17B0(t, "self")
			v.Buffs = 1<<22 | 1<<26
			r := damageMeleeWorldRuntime4E0B30(t)
			shield, hp := 0, 0
			r.CallDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
				t.Fatal("spell missile retaliated Shock")
				return false
			}
			r.BuffOff = func(v *Object, enchant EnchantID) {
				if enchant != 0 {
					t.Fatal("Shock buff removed")
				}
			}
			r.ShieldReduce = func(target *Object, amount *int32, typ object.DamageType, weapon *Object) {
				if target != v || weapon != w || *amount != 8 || typ != object.DamageImpact || v.Obj130 != w {
					t.Fatal("Shield identity/order")
				}
				shield++
				if absorbed {
					*amount = 0
				} else {
					*amount = 3
				}
			}
			r.DamageClear = func(target *Object, amount int32) {
				if target != v || amount != 3 || shield != 1 {
					t.Fatal("HP before Shield")
				}
				hp++
			}
			result := DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r)
			if result == absorbed || shield != 1 || hp != map[bool]int{false: 1, true: 0}[absorbed] || !v.HasEnchant(22) {
				t.Fatalf("result=%t shield=%d hp=%d", result, shield, hp)
			}
		})
	}
}

func TestDefaultDamagePlayerSpellMissileImpactLateDefend4E0B30(t *testing.T) {
	v, a, w := spellMissileImpactFixture4E17B0(t, "NPC")
	item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}
	m := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	item.InitDataModifier().Modifiers[2] = m
	v.InvFirstItem = item
	r := damageMeleeWorldRuntime4E0B30(t)
	var events []string
	r.CanApplyLateDefend = func(effect *ModifierEff) bool { return effect == m }
	r.ApplyLateDefend = func(effect *ModifierEff, equipped, owner, weapon, source *Object, amount int32, typ object.DamageType) int32 {
		if effect != m || equipped != item || owner != v || weapon != w || source != a || amount != 8 || typ != object.DamageImpact || v.Obj130 != nil {
			t.Fatal("late Defend identity/order")
		}
		events = append(events, "defend")
		return 33
	}
	r.Frame = func() uint32 { events = append(events, "frame"); return 1407 }
	r.GameBallOnDamage = func(source, target *Object, amount int32) {
		if source != a || target != v || amount != 33 || v.Frame134 != 1407 {
			t.Fatal("live GameBall damage")
		}
		events = append(events, "ball")
	}
	r.PlayerSetState = func(*Object, PlayerState) bool { events = append(events, "hurt"); return true }
	r.DamageClear = func(target *Object, amount int32) {
		if target != v || amount != 33 || a.UpdateDataMonster().Field130 != 1407 {
			t.Fatal("live damage/latch")
		}
		events = append(events, "hp")
	}
	if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) || !slices.Equal(events, []string{"defend", "frame", "ball", "hurt", "frame", "hp"}) {
		t.Fatalf("events=%v", events)
	}
}

func TestPlayerDamageSpellMissileImpactShape4E17B0(t *testing.T) {
	for _, owner := range []string{"self", "player", "NPC"} {
		t.Run(owner, func(t *testing.T) {
			_, a, w := spellMissileImpactFixture4E17B0(t, owner)
			if !playerDamageSpellMissileImpactShape4E17B0(a, w, object.DamageImpact) {
				t.Fatal("stock shape rejected")
			}
			w.TypeInd = 0
			if !playerDamageSpellMissileImpactShape4E17B0(a, w, object.DamageImpact) {
				t.Fatal("depends on Pixie type ID")
			}
			for _, class := range []object.Class{object.ClassPlayer, object.ClassMonster, object.ClassWeapon, object.ClassWand} {
				w.ObjClass |= class
				if playerDamageSpellMissileImpactShape4E17B0(a, w, object.DamageImpact) {
					t.Fatal("admitted separate modifier/unit branch")
				}
				w.ObjClass &^= class
			}
			if playerDamageSpellMissileImpactShape4E17B0(nil, w, object.DamageImpact) || playerDamageSpellMissileImpactShape4E17B0(a, nil, object.DamageImpact) || playerDamageSpellMissileImpactShape4E17B0(a, w, object.DamageExplosion) {
				t.Fatal("unrelated shape admitted")
			}
		})
	}
}
