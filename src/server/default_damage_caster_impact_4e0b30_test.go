package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func defaultDamageCasterImpactFixture4E0B30(t *testing.T, owner, victim string) (v, a, w *Object) {
	t.Helper()
	baseOwner, baseVictim := "imaginary", victim
	switch owner {
	case "player-self", "player-owned-NPC":
		baseOwner = "player"
	case "monster-owner", "proxy-monster", "nil":
		baseOwner = map[string]string{"monster-owner": "NPC", "proxy-monster": "proxy-NPC", "nil": "nil"}[owner]
	}
	if victim == "world" {
		baseVictim = "monster"
	}
	v, a, w = worldFlameFixture4E17B0(t, baseOwner, baseVictim)
	w.ObjClass = 0 // The stock script-only ImaginaryCaster has no class bits.
	switch owner {
	case "player-self", "imaginary-self":
		w = a
		w.ObjOwner = nil
	case "player-owned-NPC", "imaginary-owned-NPC":
		update, freeUpdate := alloc.New(MonsterUpdateData{})
		t.Cleanup(freeUpdate)
		*update = MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.375)}
		w.ObjClass, w.ObjSubClass, w.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(update)
	case "world-self":
		w.ObjClass, w.ObjOwner, a = object.ClassSimple, nil, w
	case "monster-owner", "proxy-monster", "nil":
	default:
		t.Fatalf("unknown caster %q", owner)
	}
	if victim == "world" {
		v.ObjClass, v.ObjSubClass = object.ClassObstacle, 0
	}
	v.Buffs = 1 << 0
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(w), v.UpdateData, unsafe.Pointer(v.HealthData)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("caster IMPACT fixture pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	return
}

func TestDefaultDamageCasterImpact4E0B30Tail(t *testing.T) {
	for _, victim := range []string{"monster", "NPC", "player", "world"} {
		for _, owner := range []string{"player-self", "player-owned-NPC", "imaginary-self", "imaginary-owned-NPC", "monster-owner", "proxy-monster", "world-self", "nil"} {
			for _, raw := range []int32{-3, 0, 1, 8, 25, 33} {
				for _, enemy := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/raw-%d/enemy-%t", victim, owner, raw, enemy), func(t *testing.T) {
						v, a, w := defaultDamageCasterImpactFixture4E0B30(t, owner, victim)
						var events []string
						var attacker *Object
						if owner == "monster-owner" {
							attacker = a
						} else if owner == "proxy-monster" {
							attacker = a.ObjOwner
						}
						blocked := !enemy && victim != "world" && w.Class().Has(object.ClassMonster)
						position := w.PrevPos
						if a == nil {
							position = types.Pointf{}
						}
						r := DefaultDamageWorldRuntime4E0B30{
							GameplayFlag1: func() bool { return true },
							Frame:         func() uint32 { events = append(events, "frame"); return 1400 },
							IsEnemy: func(target, source *Object) bool {
								if target != v || source == nil || source != a && source != attacker {
									t.Fatal("caster enemy identity")
								}
								events = append(events, "enemy")
								return enemy
							},
							BuffOff: func(target *Object, enchant EnchantID) {
								if blocked || target != v || enchant != 0 || a == nil || v.Pos132 != position || v.Frame134 != 77 || v.Obj130 != nil {
									t.Fatal("caster visibility prefix")
								}
								v.Buffs &^= 1 << enchant
								events = append(events, "buff")
							},
							MonsterHasHitSound: func(source *Object) bool {
								if owner != "monster-owner" || source != a {
									t.Fatal("caster hit-sound source must be the terminal owner")
								}
								events = append(events, "lookup")
								return false
							},
							DefaultDamageSound: func(target, weapon *Object) {
								if target != v || weapon != w || v.Obj130 != w || v.Field131 != 11 || v.Frame134 != 1400 {
									t.Fatal("caster sound/attribution")
								}
								events = append(events, "sound")
							},
							GameBallOnDamage: func(source, target *Object, damage int32) {
								if source != a || target != v || damage != raw || victim != "player" {
									t.Fatal("caster GameBall tail")
								}
								events = append(events, "ball")
							},
							PlayerSetState: func(target *Object, state PlayerState) bool {
								if target != v || state != PlayerState30 || victim != "player" || raw < 20 {
									t.Fatal("caster hurt-state tail")
								}
								events = append(events, "hurt")
								return true
							},
							AdjustFieldGuide: func(source, target *Object, damage int32) int32 {
								if source != a || target != v || damage != raw || victim == "world" || victim == "player" {
									t.Fatal("caster field-guide tail")
								}
								events = append(events, "guide")
								return damage
							},
							DamageClear: func(target *Object, damage int32) {
								if blocked || target != v || damage != raw || (attacker != nil && enemy && attacker.UpdateDataMonster().Field130 != 1400) {
									t.Fatal("caster HP amount/attacker latch")
								}
								events = append(events, "hp")
								v.HealthData.Cur = uint16(max(20-damage, 0))
							},
							FireProtection:     func(*Object) float64 { t.Fatal("IMPACT used fire protection"); return 0 },
							ElectricProtection: func(*Object) float64 { t.Fatal("IMPACT used electric protection"); return 0 },
							Unsupported:        func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
						}
						if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageImpact, r) {
							t.Fatal("unshielded caster IMPACT returned false")
						}
						var want []string
						if a != nil {
							want = append(want, "enemy")
						}
						if blocked {
							if !slices.Equal(events, want) || v.HealthData.Cur != 20 || v.Frame134 != 77 || v.Obj130 != nil {
								t.Fatalf("friendly caster events=%v HP=%d", events, v.HealthData.Cur)
							}
							return
						}
						if a != nil {
							want = append(want, "buff")
						}
						want = append(want, "frame")
						if owner == "monster-owner" {
							want = append(want, "lookup")
						}
						want = append(want, "sound")
						if victim == "player" {
							want = append(want, "ball")
							if raw >= 20 {
								want = append(want, "hurt")
							}
						} else if victim != "world" {
							want = append(want, "guide")
						}
						if attacker != nil {
							want = append(want, "enemy")
							if enemy {
								want = append(want, "frame")
							}
						}
						want = append(want, "hp")
						if !slices.Equal(events, want) || v.HealthData.Cur != uint16(max(20-raw, 0)) || v.Pos132 != position || v.HasEnchant(0) != (a == nil) {
							t.Fatalf("caster events=%v want=%v HP=%d position=%v", events, want, v.HealthData.Cur, v.Pos132)
						}
						if victim != "player" && victim != "world" {
							marker, markerType := uint32(1), uint32(w.TypeInd)
							if a == nil || a == w {
								marker, markerType = 2, 11
							}
							ud := v.UpdateDataMonster()
							if ud.Field547 != marker || ud.Field546 != markerType || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) || ud.Field1 != math.Float32bits(0.25) {
								t.Fatal("caster hit marker/status/carry")
							}
						}
					})
				}
			}
		}
	}
}

func TestDefaultDamageCasterImpact4E0B30EarlyOrder(t *testing.T) {
	for _, victim := range []string{"monster", "NPC", "player", "world"} {
		for _, gate := range []string{"invulnerable", "dead", "campaign-friendly", "no-update"} {
			t.Run(victim+"/"+gate, func(t *testing.T) {
				v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-self", victim)
				switch gate {
				case "invulnerable":
					v.Buffs |= 1 << 23
				case "dead":
					v.ObjFlags |= object.FlagDead
				case "no-update":
					v.ObjFlags |= object.FlagNoUpdate
				}
				var events []string
				r := DefaultDamageWorldRuntime4E0B30{
					Frame:         func() uint32 { events = append(events, "frame"); return 1401 },
					GameplayFlag1: func() bool { events = append(events, "gameplay"); return gate != "campaign-friendly" },
					IsZombie:      func(*Object) bool { events = append(events, "zombie"); return false },
					IsEnemy: func(target, source *Object) bool {
						if target != v || source != a {
							t.Fatal("early enemy identity")
						}
						events = append(events, "enemy")
						return false
					},
					QuestMode:   func() bool { t.Fatal("different friendly target must not query Quest"); return true },
					BuffOff:     func(*Object, EnchantID) { t.Fatal("early BuffOff") },
					DamageClear: func(*Object, int32) { t.Fatal("early HP") },
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
				}
				if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) {
					t.Fatal("early caster result")
				}
				want := []string{"gameplay", "enemy"}
				if gate == "invulnerable" {
					want = []string{"frame"}
				}
				if gate == "dead" {
					want = []string{"zombie"}
				}
				if !slices.Equal(events, want) || v.HealthData.Cur != 20 || v.Obj130 != nil || v.Frame134 != 77 {
					t.Fatalf("early caster events=%v want=%v", events, want)
				}
				if victim != "player" && victim != "world" {
					marker := uint32(0)
					if gate == "invulnerable" {
						marker = 99
					}
					if v.UpdateDataMonster().Field547 != marker {
						t.Fatal("early caster monster latch")
					}
				}
			})
		}
	}
}

func TestDefaultDamageCasterImpact4E0B30CampaignOwnTarget(t *testing.T) {
	for _, enemy := range []bool{false, true} {
		for _, quest := range []bool{false, true} {
			t.Run(fmt.Sprintf("enemy-%t/quest-%t", enemy, quest), func(t *testing.T) {
				v, _, w := defaultDamageCasterImpactFixture4E0B30(t, "player-owned-NPC", "player")
				w.ObjOwner = v
				var events []string
				r := damageMeleeWorldRuntime4E0B30(t)
				r.GameplayFlag1 = func() bool { return false }
				r.IsEnemy = func(target, source *Object) bool {
					if target != v || source != v {
						t.Fatal("own-target enemy identity")
					}
					events = append(events, "enemy")
					return enemy
				}
				r.QuestMode = func() bool { events = append(events, "quest"); return quest }
				r.BuffOff = func(*Object, EnchantID) { events = append(events, "buff") }
				r.DamageClear = func(*Object, int32) { events = append(events, "hp") }
				if !DefaultDamageWorld4E0B30(v, v, w, 8, object.DamageImpact, r) {
					t.Fatal("own-target result")
				}
				want := []string{"enemy"}
				if !enemy {
					want = append(want, "quest")
				}
				if enemy {
					want = append(want, "enemy", "buff", "hp")
				} else if !quest {
					want = append(want, "enemy")
				}
				if !slices.Equal(events, want) {
					t.Fatalf("campaign events=%v want=%v", events, want)
				}
			})
		}
	}
}

func TestDefaultDamageCasterImpact4E0B30ShockShield(t *testing.T) {
	for _, victim := range []string{"monster", "NPC", "player", "world"} {
		for _, owner := range []string{"player-self", "player-owned-NPC", "imaginary-self", "imaginary-owned-NPC"} {
			for _, absorbed := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/absorbed-%t", victim, owner, absorbed), func(t *testing.T) {
					v, a, w := defaultDamageCasterImpactFixture4E0B30(t, owner, victim)
					v.Buffs |= 1<<22 | 1<<26
					retaliates := owner == "player-owned-NPC"
					var events []string
					r := damageMeleeWorldRuntime4E0B30(t)
					r.IsEnemy = func(*Object, *Object) bool { events = append(events, "enemy"); return true }
					r.Audio = func(id int, source *Object) {
						if !retaliates || id != 135 || source != a || v.Obj130 != nil || v.Pos132 != types.Ptf(31, 47) {
							t.Fatal("caster Shock voice/order")
						}
						events = append(events, "audio")
					}
					r.BuffOff = func(target *Object, enchant EnchantID) {
						if target != v || enchant != 0 && (!retaliates || enchant != 22) {
							t.Fatal("caster Shock/visibility")
						}
						v.Buffs &^= 1 << enchant
						events = append(events, fmt.Sprintf("buff-%d", enchant))
					}
					r.BalanceFloatInd = func(key string, index int) float64 {
						if !retaliates || key != "ShockDamage" || index != 4 {
							t.Fatal("caster Shock balance")
						}
						events = append(events, "balance")
						return 8.5 // The stored binary32 Shock value is rounded by FISTP.
					}
					r.CallDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
						if !retaliates || target != a || source != v || weapon != nil || amount != 8 || typ != object.DamageElectric {
							t.Fatal("caster Shock damage identity")
						}
						events = append(events, "shock")
						return true
					}
					r.PlayerSetState = func(target *Object, state PlayerState) bool {
						if !retaliates || target != a || state != PlayerState23 {
							t.Fatal("caster Shock state")
						}
						events = append(events, "hurt")
						return true
					}
					r.ShieldReduce = func(target *Object, damage *int32, typ object.DamageType, weapon *Object) {
						if target != v || *damage != 8 || typ != object.DamageImpact || weapon != w || v.Obj130 != w || v.Field131 != 11 {
							t.Fatal("caster Shield identity/order")
						}
						events = append(events, "shield")
						if absorbed {
							*damage = 0
						}
					}
					r.DamageClear = func(*Object, int32) { events = append(events, "hp") }
					result := DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r)
					want := []string{"enemy"}
					if retaliates {
						want = append(want, "audio", "buff-22", "balance", "shock", "hurt")
					}
					want = append(want, "buff-0", "shield")
					if !absorbed {
						want = append(want, "hp")
					}
					if result == absorbed || !slices.Equal(events, want) || v.HasEnchant(22) == retaliates {
						t.Fatalf("caster Shock/Shield result=%t events=%v want=%v", result, events, want)
					}
				})
			}
		}
	}
}

func TestDefaultDamageCasterImpact4E0B30LiveDefense(t *testing.T) {
	for _, victim := range []string{"NPC", "player"} {
		t.Run(victim, func(t *testing.T) {
			v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-owned-NPC", victim)
			item, freeItem := alloc.New(Object{})
			m, freeModifier := alloc.New(ModifierEff{})
			fn, freeFn := alloc.New(byte(0))
			init, freeInit := alloc.New(ModifierInitData{})
			liveTarget, freeTarget := alloc.New(MonsterUpdateData{})
			liveSource, freeSource := alloc.New(MonsterUpdateData{})
			for _, free := range []func(){freeItem, freeModifier, freeFn, freeInit, freeTarget, freeSource} {
				t.Cleanup(free)
			}
			*item, *m, *init = Object{ObjFlags: object.FlagEquipped}, ModifierEff{}, ModifierInitData{}
			*liveTarget, *liveSource = MonsterUpdateData{}, MonsterUpdateData{}
			m.Defend76.Fnc, init.Modifiers[2] = unsafe.Pointer(fn), m
			item.InitData, v.InvFirstItem = unsafe.Pointer(init), item
			cachedSource := w.UpdateDataMonster()
			var cachedTarget *MonsterUpdateData
			if victim == "NPC" {
				cachedTarget = v.UpdateDataMonster()
			}
			var events []string
			frame := uint32(1400)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.Frame = func() uint32 { events = append(events, "frame"); return frame }
			r.BuffOff = func(target *Object, enchant EnchantID) {
				if target != v || enchant != 0 || v.Obj130 != nil || v.Pos132 != w.PrevPos {
					t.Fatal("caster defense visibility")
				}
				events = append(events, "buff")
				if victim == "NPC" {
					v.UpdateData = unsafe.Pointer(liveTarget)
				}
				w.UpdateData = unsafe.Pointer(liveSource)
			}
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(effect *ModifierEff, equipped, owner, weapon, source *Object, damage int32, typ object.DamageType) int32 {
				if effect != m || equipped != item || owner != v || weapon != w || source != a || damage != 8 || typ != object.DamageImpact || v.Frame134 != 77 {
					t.Fatal("caster late defense identity/order")
				}
				frame = 1417
				events = append(events, "defend")
				return 3
			}
			r.DefaultDamageSound = func(target, attack *Object) {
				if target != v || attack != w || v.Obj130 != w || v.Frame134 != 1417 {
					t.Fatal("caster live sound/attribution")
				}
				events = append(events, "sound")
			}
			r.DamageClear = func(target *Object, damage int32) {
				if target != v || damage != 3 || w.UpdateDataMonster() != liveSource || cachedSource.Field523_2 != 66 {
					t.Fatal("caster live HP")
				}
				events = append(events, "hp")
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) || !slices.Equal(events, []string{"buff", "defend", "frame", "sound", "hp"}) {
				t.Fatalf("caster live events=%v", events)
			}
			if victim == "NPC" && (cachedTarget.Field547 != 1 || cachedTarget.Field546 != uint32(w.TypeInd) || cachedTarget.StatusFlags != 0 || liveTarget.Field547 != 2 || liveTarget.Field546 != 11 || !liveTarget.StatusFlags.Has(object.MonStatusInjured)) {
				t.Fatal("caster retained stale target update")
			}
		})
	}
}

func TestDefaultDamageCasterImpact4E0B30LiveQualifier(t *testing.T) {
	for _, victim := range []string{"monster", "player", "world"} {
		for _, live := range []string{"player", "class-zero", "monster", "melee", "hammer", "wand"} {
			t.Run(victim+"/"+live, func(t *testing.T) {
				v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-owned-NPC", victim)
				init, freeInit := alloc.New(ModifierInitData{})
				t.Cleanup(freeInit)
				*init = ModifierInitData{}
				w.InitData = unsafe.Pointer(init)
				r, queries, hp := damageMeleeWorldRuntime4E0B30(t), 0, false
				r.IsEnemy = func(target, source *Object) bool {
					if target != v || source != a {
						t.Fatal("live qualifier enemy identity")
					}
					queries++
					switch live {
					case "player":
						w.ObjClass = object.ClassPlayer
					case "class-zero":
						w.ObjClass = 0
					case "melee":
						w.ObjClass, w.ObjSubClass = object.ClassWeapon, 0
					case "hammer":
						w.ObjClass, w.ObjSubClass = object.ClassWeapon, 0x4000
					case "wand":
						w.ObjClass, w.ObjSubClass = object.ClassWand, 0
					}
					return false
				}
				r.DamageClear = func(target *Object, damage int32) {
					if target != v || damage != 8 || v.Obj130 != w || v.Field131 != 11 {
						t.Fatal("live qualifier attribution")
					}
					hp = true
				}
				allowed := victim == "world" || live == "player" || live == "class-zero" || live == "hammer"
				if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) || queries != 1 || hp != allowed {
					t.Fatalf("live qualifier queries=%d HP-called=%t want=%t", queries, hp, allowed)
				}
				if !allowed && (v.Obj130 != nil || v.Frame134 != 77 || v.Pos132 != types.Ptf(31, 47)) {
					t.Fatal("live friendly gate committed a hit prefix")
				}
			})
		}
	}
	t.Run("wand-nil-use-before-NoUpdate", func(t *testing.T) {
		v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-owned-NPC", "monster")
		v.ObjFlags |= object.FlagNoUpdate
		r, queries := damageMeleeWorldRuntime4E0B30(t), 0
		r.IsEnemy = func(*Object, *Object) bool {
			queries++
			w.ObjClass, w.ObjSubClass, w.UseData = object.ClassWand, 0x20000, UseDataPtr{}
			return false
		}
		r.DamageClear = func(*Object, int32) { t.Error("HP after qualifier fault") }
		defer func() {
			if recover() == nil || queries != 1 || v.UpdateDataMonster().Field547 != 0 || v.Obj130 != nil || v.Frame134 != 77 || v.HealthData.Cur != 20 {
				t.Fatal("live caster qualifier fault boundary")
			}
		}()
		DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r)
	})
}

func TestDefaultDamageCasterImpact4E0B30MissingServices(t *testing.T) {
	for _, service := range []string{"enemy", "buff", "HP", "monster-sound"} {
		t.Run(service, func(t *testing.T) {
			owner := "player-self"
			if service == "monster-sound" {
				owner = "monster-owner"
			}
			v, a, w := defaultDamageCasterImpactFixture4E0B30(t, owner, "monster")
			r := damageMeleeWorldRuntime4E0B30(t)
			switch service {
			case "enemy":
				r.IsEnemy = nil
			case "buff":
				r.BuffOff = nil
			case "HP":
				r.DamageClear = nil
			case "monster-sound":
				r.MonsterHasHitSound = nil
			}
			rejections := 0
			r.Unsupported = func(_ string, target, source, weapon *Object, damage int32, typ object.DamageType) {
				if target != v || source != a || weapon != w || damage != 8 || typ != object.DamageImpact {
					t.Fatal("caster rejection identity")
				}
				rejections++
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) || rejections != 1 || v.Obj130 != nil || v.Frame134 != 77 || v.HealthData.Cur != 20 || v.UpdateDataMonster().Field547 != 0 {
				t.Fatal("missing caster service crossed tail stores")
			}
		})
	}
}

func TestDefaultDamageCasterImpact4E0B30NilDataAtUse(t *testing.T) {
	for _, victim := range []string{"monster", "NPC", "player", "world"} {
		t.Run(victim+"/nil-health", func(t *testing.T) {
			v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-self", victim)
			v.HealthData = nil // 004EE5E0, not the attack prefix, guards nil health.
			r, hp := damageMeleeWorldRuntime4E0B30(t), false
			r.DamageClear = func(target *Object, damage int32) {
				hp = target == v && damage == 8 && v.Obj130 == w && v.Frame134 == 1400
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r) || !hp {
				t.Fatal("nil health removed the original attribution prefix")
			}
		})
	}
	for _, raw := range []int32{8, 25} {
		t.Run(fmt.Sprintf("player/nil-update/raw-%d", raw), func(t *testing.T) {
			v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-self", "player")
			v.UpdateData = nil
			r, hp := damageMeleeWorldRuntime4E0B30(t), false
			r.DamageClear = func(*Object, int32) { hp = true }
			if raw >= 20 {
				defer func() {
					if recover() == nil || hp || v.Obj130 != w || v.Frame134 != 1400 || v.Field131 != 11 {
						t.Fatal("caster player nil-update fault prefix")
					}
				}()
			}
			if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageImpact, r) || !hp {
				t.Fatal("low IMPACT must not read player update")
			}
		})
	}
	t.Run("monster-source/nil-update", func(t *testing.T) {
		v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "monster-owner", "monster")
		a.UpdateData = nil // Hit sound guards nil; 00532880's post-sound latch does not.
		r, buff, sound, hp, queries := damageMeleeWorldRuntime4E0B30(t), false, false, false, 0
		r.IsEnemy = func(*Object, *Object) bool { queries++; return true }
		r.BuffOff = func(*Object, EnchantID) { buff = true }
		r.MonsterHasHitSound = func(*Object) bool { t.Fatal("nil update was queried for hit sound"); return false }
		r.DefaultDamageSound = func(*Object, *Object) { sound = true }
		r.DamageClear = func(*Object, int32) { hp = true }
		defer func() {
			if recover() == nil || !buff || !sound || hp || queries != 2 || v.Obj130 != w || v.Frame134 != 1400 {
				t.Fatal("caster monster nil-update fault prefix")
			}
		}()
		DefaultDamageWorld4E0B30(v, a, w, 8, object.DamageImpact, r)
	})
}

func TestDefaultDamageCasterImpactShape4E0B30(t *testing.T) {
	for _, class := range []object.Class{0, object.ClassPlayer, object.ClassMonster, object.ClassSimple, object.ClassLight | object.ClassObstacle} {
		t.Run(fmt.Sprintf("caster-%x", uint32(class)), func(t *testing.T) {
			w := &Object{ObjClass: class}
			if !defaultDamageCasterImpactShape4E0B30(w, object.DamageImpact) {
				t.Fatal("caster admission depends on update data")
			}
			for _, typ := range []object.DamageType{object.DamageCrush, object.DamageElectric, object.DamageExplosion, object.DamagePoison} {
				if defaultDamageCasterImpactShape4E0B30(w, typ) {
					t.Fatal("caster IMPACT broadened another damage type")
				}
			}
			for _, item := range []object.Class{object.ClassWeapon, object.ClassWand, object.ClassMissile} {
				w.ObjClass = class | item
				if defaultDamageCasterImpactShape4E0B30(w, object.DamageImpact) {
					t.Fatal("caster IMPACT broadened a weapon/projectile tail")
				}
			}
		})
	}
	if defaultDamageCasterImpactShape4E0B30(nil, object.DamageImpact) {
		t.Fatal("weapon-less IMPACT has a separate branch")
	}
}
