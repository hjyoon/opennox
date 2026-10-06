package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func defaultDamageCloudFixture4E0B30(t *testing.T, kind string, subclass object.SubClass) (target, source, cloud *Object) {
	t.Helper()
	target = defaultDamagePoisonFixture4E0B30(t, subclass)
	cloud, freeCloud := alloc.New(Object{})
	t.Cleanup(freeCloud)
	// Stock ToxicCloud is 0x190008, with neither WEAPON nor WAND bits.
	*cloud = Object{TypeInd: 1221, ObjClass: 0x190008, PrevPos: types.Ptf(17, 19)}
	if kind == "nil" {
		return target, nil, cloud
	}
	if kind == "self" {
		return target, cloud, cloud
	}
	source, freeSource := alloc.New(Object{})
	t.Cleanup(freeSource)
	*source = Object{TypeInd: 1399, PrevPos: types.Ptf(101, 103)}
	cloud.ObjOwner = source
	switch kind {
	case "imaginary": // The real position-to-position script caster has class 0.
	case "player":
		update, freeUpdate := alloc.New(PlayerUpdateData{})
		t.Cleanup(freeUpdate)
		*update = PlayerUpdateData{}
		source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(update)
	case "NPC", "proxy-NPC":
		update, freeUpdate := alloc.New(MonsterUpdateData{})
		t.Cleanup(freeUpdate)
		*update = MonsterUpdateData{}
		source.ObjClass, source.UpdateData = object.ClassMonster, unsafe.Pointer(update)
		if kind == "proxy-NPC" {
			proxy, freeProxy := alloc.New(Object{})
			t.Cleanup(freeProxy)
			*proxy = Object{ObjOwner: source, ObjClass: object.ClassSimple}
			source, cloud.ObjOwner = proxy, proxy
		}
	default:
		t.Fatalf("unknown cloud source %q", kind)
	}
	return
}

func TestDefaultDamageCloudPoison4E0B30Tail(t *testing.T) {
	for _, kind := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
		for _, subclass := range []object.SubClass{2, 0x10, 0x400, 0x800, 0x11012} {
			for _, raw := range []int32{-3, 0, 3, 10, 25} {
				for _, enemy := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/subclass-%x/raw-%d/enemy-%t", kind, uint32(subclass), raw, enemy), func(t *testing.T) {
						v, a, w := defaultDamageCloudFixture4E0B30(t, kind, subclass)
						ud := v.UpdateDataMonster()
						var events []string
						var attacker *Object
						if kind == "NPC" {
							attacker = a
						} else if kind == "proxy-NPC" {
							attacker = a.ObjOwner
						}
						marker, markerType := uint32(1), uint32(w.TypeInd)
						prefixMarker, prefixType := marker, markerType
						position := w.PrevPos
						if a == nil || a == w {
							marker, markerType = 2, 5
							prefixMarker, prefixType = 0, 91
						}
						if a == nil {
							position = types.Pointf{}
						}
						r := DefaultDamageWorldRuntime4E0B30{
							GameplayFlag1: func() bool { return true },
							Frame: func() uint32 {
								events = append(events, "frame")
								return 1400
							},
							IsEnemy: func(target, source *Object) bool {
								if target != v || (source != a && source != attacker) || source == nil {
									t.Fatal("enemy identity")
								}
								events = append(events, "enemy")
								return enemy
							},
							BuffOff: func(target *Object, enchant EnchantID) {
								if target != v || enchant != 0 || a == nil || v.Pos132 != position ||
									ud.Field547 != prefixMarker || ud.Field546 != prefixType || v.Frame134 != 77 {
									t.Fatal("visibility/position/type-latch ordering")
								}
								v.Buffs &^= 1 << enchant
								events = append(events, "buff")
							},
							MonsterHasHitSound: func(source *Object) bool {
								if kind != "NPC" || source != a {
									t.Fatal("monster sound identity")
								}
								events = append(events, "lookup")
								return false
							},
							DefaultDamageSound: func(target, weapon *Object) {
								if target != v || weapon != w || v.Obj130 != w || v.Field131 != 5 ||
									v.Frame134 != 1400 || ud.Field547 != marker || ud.Field546 != markerType {
									t.Fatal("sound/attribution ordering")
								}
								events = append(events, "sound")
							},
							AdjustFieldGuide: func(source, target *Object, damage int32) int32 {
								if source != a || target != v || damage != raw {
									t.Fatal("field-guide identity")
								}
								events = append(events, "guide")
								return damage
							},
							DamageClear: func(target *Object, damage int32) {
								if target != v || damage != raw || (attacker != nil && enemy && attacker.UpdateDataMonster().Field130 != 1400) {
									t.Fatal("raw HP/attacker-latch ordering")
								}
								events = append(events, "hp")
								v.HealthData.Cur = uint16(max(20-damage, 0))
							},
							FireProtection:     func(*Object) float64 { t.Fatal("poison used fire protection"); return 0 },
							ElectricProtection: func(*Object) float64 { t.Fatal("poison used electric protection"); return 0 },
							ShieldReduce:       func(*Object, *int32, object.DamageType, *Object) { t.Fatal("Shield reduced poison") },
							CallDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
								t.Fatal("cloud retaliated Shock")
								return false
							},
							Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
								t.Fatalf("cloud poison rejected: %s", reason)
							},
						}
						if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamagePoison, r) {
							t.Fatal("poison returned zero")
						}
						var want []string
						if a != nil {
							want = append(want, "enemy", "buff")
						}
						want = append(want, "frame")
						if kind == "NPC" {
							want = append(want, "lookup")
						}
						want = append(want, "sound", "guide")
						if attacker != nil {
							want = append(want, "enemy")
							if enemy {
								want = append(want, "frame")
							}
						}
						want = append(want, "hp")
						if !slices.Equal(events, want) || v.HealthData.Cur != uint16(max(20-raw, 0)) || v.Pos132 != position ||
							!v.HasEnchant(22) || !v.HasEnchant(26) || v.HasEnchant(0) != (a == nil) ||
							!ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) {
							t.Fatalf("events=%v want=%v HP=%d position=%v", events, want, v.HealthData.Cur, v.Pos132)
						}
					})
				}
			}
		}
	}
}

func TestDefaultDamageCloudPoison4E0B30EarlyGates(t *testing.T) {
	for _, gate := range []string{"invulnerable", "dead", "zombie", "campaign-friendly", "no-update", "enemy-no-update", "poison-immune", "enemy-poison-immune", "enemy-melee"} {
		t.Run(gate, func(t *testing.T) {
			v, a, w := defaultDamageCloudFixture4E0B30(t, "player", 0x10)
			ud := v.UpdateDataMonster()
			wantQueries, wantLatch := 1, uint32(0)
			switch gate {
			case "invulnerable":
				v.Buffs |= 1 << 23
				wantQueries, wantLatch = 0, 99
			case "dead", "zombie":
				v.ObjFlags |= object.FlagDead
				wantQueries = 0
			case "no-update":
				v.ObjFlags |= object.FlagNoUpdate
			case "poison-immune":
				v.ObjSubClass = 0x10202 // Stock Troll: immunity must not change HP.
				v.HealthData = nil      // The original exits before health/late defense.
			}
			queries, audio := 0, 0
			r := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 1400 },
				GameplayFlag1: func() bool { return gate != "campaign-friendly" },
				IsZombie: func(target *Object) bool {
					if target != v {
						t.Fatal("zombie identity")
					}
					return gate == "zombie"
				},
				IsEnemy: func(target, source *Object) bool {
					if target != v || source != a {
						t.Fatal("gate enemy identity")
					}
					queries++
					switch gate {
					case "enemy-no-update":
						v.ObjFlags |= object.FlagNoUpdate
					case "enemy-poison-immune":
						v.ObjSubClass |= 0x200
						v.HealthData = nil
					case "enemy-melee":
						w.ObjClass = object.ClassWeapon
					}
					return false
				},
				Audio: func(id int, target *Object) {
					if gate != "invulnerable" || id != 71 || target != v {
						t.Fatal("early audio")
					}
					audio++
				},
				BuffOff:     func(*Object, EnchantID) { t.Fatal("early gate removed invisibility") },
				DamageClear: func(*Object, int32) { t.Fatal("early gate reached HP") },
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("early gate rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamagePoison, r) || queries != wantQueries || ud.Field547 != wantLatch || ud.Field546 != 91 || ud.StatusFlags != 0 {
				t.Fatalf("queries=%d/%d latch=%d/%d status=%x", queries, wantQueries, ud.Field547, ud.Field546, ud.StatusFlags)
			}
			if gate == "zombie" {
				if v.Obj130 != w || v.Field131 != 5 || v.Frame134 != 1400 {
					t.Fatal("zombie attribution")
				}
			} else if v.Obj130 != nil || v.Field131 != 0 || v.Frame134 != 77 {
				t.Fatal("early gate attribution")
			}
			if audio != map[bool]int{false: 0, true: 1}[gate == "invulnerable"] || v.Pos132 != types.Ptf(31, 47) {
				t.Fatal("early gate state")
			}
		})
	}
}

func TestDefaultDamageCloudPoison4E0B30Services(t *testing.T) {
	for _, missing := range []string{"enemy", "visibility", "HP", "monster-sound"} {
		t.Run(missing, func(t *testing.T) {
			v, a, w := defaultDamageCloudFixture4E0B30(t, "NPC", 0x10)
			reported := ""
			r := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1:      func() bool { return true },
				IsEnemy:            func(*Object, *Object) bool { return true },
				BuffOff:            func(*Object, EnchantID) { t.Fatal("missing-service hit partially applied") },
				MonsterHasHitSound: func(*Object) bool { t.Fatal("missing-service sound executed"); return false },
				DamageClear:        func(*Object, int32) { t.Fatal("missing-service hit reached HP") },
				Unsupported: func(reason string, target, source, weapon *Object, raw int32, typ object.DamageType) {
					if target != v || source != a || weapon != w || raw != 3 || typ != object.DamagePoison {
						t.Fatal("missing-service identity")
					}
					reported = reason
				},
			}
			switch missing {
			case "enemy":
				r.IsEnemy = nil
			case "visibility":
				r.BuffOff = nil
			case "HP":
				r.DamageClear = nil
			case "monster-sound":
				r.MonsterHasHitSound = nil
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamagePoison, r) || reported == "" || v.HealthData.Cur != 20 || v.Obj130 != nil || v.Frame134 != 77 || v.UpdateDataMonster().Field547 != 0 {
				t.Fatalf("reported=%q", reported)
			}
		})
	}
}

func TestDefaultDamageCloudPoison4E0B30LiveDefendAndFrame(t *testing.T) {
	v, a, w := defaultDamageCloudFixture4E0B30(t, "NPC", 0x10)
	cachedTarget, cachedSource := v.UpdateDataMonster(), a.UpdateDataMonster()
	liveTarget, freeTarget := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeTarget)
	*liveTarget = MonsterUpdateData{Field546: 91}
	liveSource, freeSource := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeSource)
	*liveSource = MonsterUpdateData{}
	m, freeModifier := alloc.New(ModifierEff{})
	t.Cleanup(freeModifier)
	*m = ModifierEff{}
	f, freeFn := alloc.New(byte(0))
	t.Cleanup(freeFn)
	m.Defend76.Fnc = unsafe.Pointer(f)
	data, freeData := alloc.New(ModifierInitData{})
	t.Cleanup(freeData)
	*data = ModifierInitData{}
	data.Modifiers[2] = m
	item, freeItem := alloc.New(Object{})
	t.Cleanup(freeItem)
	*item = Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(data)}
	v.InvFirstItem = item
	frame := uint32(1400)
	var events []string
	r := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		Frame:         func() uint32 { events = append(events, "frame"); return frame },
		IsEnemy: func(target, source *Object) bool {
			if target != v || source != a {
				t.Fatal("live enemy identity")
			}
			events = append(events, "enemy")
			return true
		},
		BuffOff: func(target *Object, enchant EnchantID) {
			if target != v || enchant != 0 || cachedTarget.Field547 != 1 || cachedTarget.Field546 != 1221 || v.Pos132 != types.Ptf(17, 19) {
				t.Fatal("live visibility prefix")
			}
			events = append(events, "buff")
		},
		CanApplyLateDefend: func(effect *ModifierEff) bool { return effect == m },
		ApplyLateDefend: func(effect *ModifierEff, equipped, owner, weapon, source *Object, raw int32, typ object.DamageType) int32 {
			if effect != m || equipped != item || owner != v || weapon != w || source != a || raw != 8 || typ != object.DamagePoison || v.Obj130 != nil {
				t.Fatal("live late-defense identity")
			}
			events = append(events, "defend")
			v.UpdateData, frame = unsafe.Pointer(liveTarget), 1407
			return 9
		},
		MonsterHasHitSound: func(source *Object) bool {
			if source != a {
				t.Fatal("live source sound")
			}
			events = append(events, "lookup")
			return false
		},
		DefaultDamageSound: func(target, weapon *Object) {
			if target != v || weapon != w || v.Obj130 != w || v.Field131 != 5 || v.Frame134 != 1407 || liveTarget.Field547 != 2 || liveTarget.Field546 != 5 {
				t.Fatal("live attribution after defense")
			}
			events = append(events, "sound")
			a.UpdateData = unsafe.Pointer(liveSource)
		},
		AdjustFieldGuide: func(source, target *Object, raw int32) int32 {
			if source != a || target != v || raw != 9 {
				t.Fatal("live guide")
			}
			events = append(events, "guide")
			frame = 1411
			return 12
		},
		DamageClear: func(target *Object, raw int32) {
			if target != v || raw != 12 || liveSource.Field130 != 1411 || cachedSource.Field130 != 0 {
				t.Fatal("live HP/source latch")
			}
			events = append(events, "hp")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("live poison rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(v, a, w, 8, object.DamagePoison, r) || !slices.Equal(events, []string{"enemy", "buff", "defend", "frame", "lookup", "sound", "guide", "enemy", "frame", "hp"}) ||
		cachedTarget.Field547 != 1 || cachedTarget.Field546 != 1221 || cachedTarget.StatusFlags != 0 || !liveTarget.StatusFlags.Has(object.MonStatusInjured) {
		t.Fatalf("events=%v", events)
	}
}

func TestDefaultDamageCloudPoison4E0B30ShockBeforeLiveImmunity(t *testing.T) {
	for _, change := range []string{"retain", "add", "remove"} {
		t.Run(change, func(t *testing.T) {
			v, a, w := defaultDamageCloudFixture4E0B30(t, "player", 0x10)
			init, freeInit := alloc.New(ModifierInitData{})
			t.Cleanup(freeInit)
			*init = ModifierInitData{}
			w.InitData = unsafe.Pointer(init) // The live WEAPON class has a real slot base.
			if change != "add" {
				v.ObjSubClass |= 0x200
			}
			if change != "remove" {
				v.HealthData = nil // Immunity exits after Shock but before HP.
			}
			var events []string
			hits := 0
			r := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return true },
				Frame:         func() uint32 { return 1400 },
				IsEnemy: func(target, source *Object) bool {
					if target != v || source != a {
						t.Fatal("live Shock enemy identity")
					}
					events = append(events, "enemy")
					w.ObjClass = object.ClassWeapon
					return true
				},
				Audio: func(id int, source *Object) {
					if id != 135 || source != a {
						t.Fatal("Shock audio")
					}
					events = append(events, "audio")
				},
				BuffOff: func(target *Object, enchant EnchantID) {
					if target != v || enchant != 22 && (change != "remove" || enchant != 0) {
						t.Fatal("Shock/visibility removal")
					}
					v.Buffs &^= 1 << enchant
					events = append(events, fmt.Sprintf("buff-%d", enchant))
				},
				BalanceFloatInd: func(key string, index int) float64 {
					if key != "ShockDamage" || index != 4 {
						t.Fatal("Shock balance")
					}
					events = append(events, "balance")
					return 8
				},
				CallDamage: func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
					if target != a || source != v || weapon != nil || damage != 8 || typ != object.DamageElectric {
						t.Fatal("Shock damage identity")
					}
					events = append(events, "shock")
					if change == "remove" {
						v.ObjSubClass &^= 0x200
					} else {
						v.ObjSubClass |= 0x200
					}
					return true
				},
				PlayerSetState: func(source *Object, state PlayerState) bool {
					if source != a || state != PlayerState23 {
						t.Fatal("Shock hurt state")
					}
					events = append(events, "hurt")
					return true
				},
				DamageClear: func(target *Object, damage int32) {
					if change != "remove" || target != v || damage != 3 || v.Obj130 != w || v.Field131 != 5 || v.Pos132 != a.PrevPos {
						t.Fatal("live immunity/weapon HP tail")
					}
					hits++
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("live immunity rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamagePoison, r) {
				t.Fatal("live immunity return")
			}
			want := []string{"enemy", "audio", "buff-22", "balance", "shock", "hurt"}
			if change == "remove" {
				want = append(want, "buff-0")
			}
			if !slices.Equal(events, want) || hits != map[bool]int{false: 0, true: 1}[change == "remove"] {
				t.Fatalf("events=%v want=%v HP-calls=%d", events, want, hits)
			}
		})
	}
}

func TestDefaultDamageCloudPoison4E0B30AdmissionBoundary(t *testing.T) {
	for _, class := range []object.Class{0x190008 | object.ClassMonster, 0x190008 | object.ClassWeapon, 0x190008 | object.ClassWand, 0x190008 | object.ClassMissile, object.ClassSimple, object.ClassDangerous} {
		t.Run(fmt.Sprintf("class-%x", uint32(class)), func(t *testing.T) {
			v, _, w := defaultDamageCloudFixture4E0B30(t, "nil", 2)
			w.ObjClass = class
			reported := ""
			r := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return true },
				DamageClear:   func(*Object, int32) { t.Fatal("unsupported world shape reached HP") },
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					reported = reason
				},
			}
			if !DefaultDamageWorld4E0B30(v, nil, w, 3, object.DamagePoison, r) || reported != "unsupported monster damage shape" || v.HealthData.Cur != 20 || v.Obj130 != nil || v.Frame134 != 77 {
				t.Fatalf("boundary reported=%q", reported)
			}
		})
	}
}
