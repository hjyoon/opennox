package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageNative4E17B0NPCReflectElectricFacing(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		for _, selfWeapon := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				for _, front := range []bool{false, true} {
					t.Run(fmt.Sprintf("player-source-%t/self-%t/%s/front-%t", playerSource, selfWeapon, typ, front), func(t *testing.T) {
						target, source := defaultDamageElectricSelfFixture4E0B30(t, false, playerSource)
						target.Buffs = 1 << playerDamageReflectEnchant4E17B0
						source.PosVec = types.Pointf{X: 17, Y: -30}
						weapon := (*Object)(nil)
						if selfWeapon {
							weapon = source
						}
						ud := target.UpdateDataMonster()
						ud.Field1 = math.Float32bits(0.25)
						before := *ud
						var events []string
						r := PlayerDamageRuntime4E17B0{
							BlockDirection: func(got *Object, pos types.Pointf) bool {
								if got != target || pos != source.PosVec || typ != object.DamageAirborneElectric {
									t.Fatal("NPC Reflect must face the original weapon-or-source current position")
								}
								events = append(events, "direction")
								return front
							},
							Audio: func(id int, got *Object) {
								if id != 122 || got != target || ud.Field547 != 0 || ud.Field546 != 77 || ud.Field1 != before.Field1 {
									t.Fatal("NPC Reflect audio/marker/carry order")
								}
								events = append(events, "audio")
							},
							ElectricArmorScale: func(got *Object) float32 {
								if got != target {
									t.Fatal("NPC electric scale target")
								}
								events = append(events, "scale")
								return 1
							},
							DefaultDamage: func(got, attacker, originalWeapon *Object, damage int32, gotType object.DamageType) bool {
								if got != target || attacker != source || originalWeapon != weapon || damage != 8 || gotType != typ || ud.Field547 != 2 || ud.Field546 != uint32(typ) {
									t.Fatal("rear/type-9 hit lost the electric default arguments or attribution")
								}
								events = append(events, "default")
								return true
							},
							ObserveClear:      func(*Object) { t.Fatal("NPC read as a player observer") },
							ProjectileReflect: func(*Object, *Object) { t.Fatal("non-missile reflected a projectile") },
							PointFX:           func(int, types.Pointf) { t.Fatal("type 9/17 emitted ZapRay point FX") },
							Unsupported:       func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
						}
						reflected := typ == object.DamageAirborneElectric && front
						if handled, result := PlayerDamageNative4E17B0(target, source, weapon, 8, typ, r); !handled || result == reflected {
							t.Fatalf("NPC reflection=%t/%t want reflected=%t", handled, result, reflected)
						}
						want := []string{"scale", "default"}
						if typ == object.DamageAirborneElectric {
							want = append([]string{"direction"}, want...)
						}
						if reflected {
							want = []string{"direction", "audio"}
							before.Field547 = 0
							if *ud != before {
								t.Fatal("reflection changed NPC state outside the original hit-marker DWORD")
							}
						}
						if !slices.Equal(events, want) || target.HealthData.Cur != 60 || target.Buffs != 1<<playerDamageReflectEnchant4E17B0 {
							t.Fatalf("events=%v want=%v HP=%d buffs=%#x", events, want, target.HealthData.Cur, target.Buffs)
						}
					})
				}
			}
		}
	}
}

func npcReflectRuntime4E17B0(t *testing.T, target, attack *Object, events *[]string) PlayerDamageRuntime4E17B0 {
	t.Helper()
	return PlayerDamageRuntime4E17B0{
		BlockDirection: func(got *Object, pos types.Pointf) bool {
			if got != target || pos != attack.PosVec {
				t.Fatal("NPC missile reflection facing")
			}
			*events = append(*events, "direction")
			return true
		},
		ProjectileReflect: func(got, reflector *Object) {
			if got != attack || reflector != target || target.UpdateDataMonster().Field547 != 0 {
				t.Fatal("NPC projectile reflection identity/prefix")
			}
			*events = append(*events, "reflect")
		},
		ClearOwner: func(got *Object) {
			if got != attack {
				t.Fatal("NPC reflection clear owner")
			}
			*events = append(*events, "clear")
		},
		SetOwner: func(owner, got *Object) {
			if got != attack || owner != target {
				t.Fatal("NPC reflection set owner argument order")
			}
			*events = append(*events, "set")
		},
		ChangeOwner: func(got, owner *Object) {
			if got != attack || owner != target {
				t.Fatal("NPC reflection change owner argument order")
			}
			*events = append(*events, "change")
		},
		PointFX: func(id int, pos types.Pointf) {
			if id != 132 || pos != target.PosVec {
				t.Fatal("NPC ZapRay point FX")
			}
			*events = append(*events, "point")
		},
		Audio: func(id int, got *Object) {
			if id != 122 || got != target {
				t.Fatal("NPC reflection audio")
			}
			*events = append(*events, "audio")
		},
		ObserveClear: func(*Object) { t.Fatal("NPC reflected through player layout") },
		DefaultDamage: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
			t.Fatal("reflected hit entered default damage")
			return false
		},
	}
}

func TestPlayerDamageNative4E17B0NPCReflectMissileAndZapRay(t *testing.T) {
	for _, flags := range []uint32{0, 2, 0x40, 0x42} {
		for _, typ := range []object.DamageType{object.DamageImpact, object.DamageZapRay} {
			for _, damage := range []int32{0, 8, -5} {
				for _, sourceLess := range []bool{false, true} {
					t.Run(fmt.Sprintf("subclass-%x/%s/damage-%d/source-less-%t", flags, typ, damage, sourceLess), func(t *testing.T) {
						target, source := defaultDamageElectricSelfFixture4E0B30(t, false, true)
						target.Buffs = 1 << playerDamageReflectEnchant4E17B0
						target.PosVec = types.Pointf{X: 10, Y: 20}
						attack := &Object{ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(flags), PosVec: types.Pointf{X: 37, Y: 14}, PrevPos: types.Pointf{X: -99, Y: -14}}
						if sourceLess {
							source = nil
						}
						before := *target.UpdateDataMonster()
						var events []string
						r := npcReflectRuntime4E17B0(t, target, attack, &events)
						r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }
						if handled, result := PlayerDamageNative4E17B0(target, source, attack, damage, typ, r); !handled || result {
							t.Fatalf("NPC missile reflect=%t/%t", handled, result)
						}
						want := []string{"direction", "reflect"}
						if flags&0x40 == 0 {
							want = append(want, "clear", "set")
						}
						if flags&2 != 0 {
							want = append(want, "change")
						}
						if typ == object.DamageZapRay {
							want = append(want, "point")
						}
						want = append(want, "audio")
						before.Field547 = 0
						if !slices.Equal(events, want) || *target.UpdateDataMonster() != before || target.HealthData.Cur != 60 {
							t.Fatalf("NPC reflection events=%v want=%v HP=%d", events, want, target.HealthData.Cur)
						}
					})
				}
			}
		}
	}
	// Non-missile ZapRay takes only direction, point-FX and audio, even if its
	// source is absent and the original weapon pointer supplies the position.
	target, _ := defaultDamageElectricSelfFixture4E0B30(t, false, true)
	target.Buffs = 1 << playerDamageReflectEnchant4E17B0
	attack := &Object{PosVec: types.Pointf{X: 17, Y: 5}}
	var events []string
	r := npcReflectRuntime4E17B0(t, target, attack, &events)
	if handled, result := PlayerDamageNative4E17B0(target, nil, attack, 0, object.DamageZapRay, r); !handled || result || !slices.Equal(events, []string{"direction", "point", "audio"}) {
		t.Fatalf("NPC non-missile ZapRay=%t/%t events=%v", handled, result, events)
	}
}

func TestPlayerDamageNative4E17B0NPCReflectMissingServiceDoesNotMutate(t *testing.T) {
	for _, missing := range []string{"direction", "reflect", "clear", "set", "change", "point", "audio"} {
		t.Run(missing, func(t *testing.T) {
			target, source := defaultDamageElectricSelfFixture4E0B30(t, false, true)
			target.Buffs = 1 << playerDamageReflectEnchant4E17B0
			attack := &Object{ObjClass: object.ClassMissile, ObjSubClass: 2}
			beforeTarget, beforeUpdate, beforeAttack := *target, *target.UpdateDataMonster(), *attack
			var events []string
			r := npcReflectRuntime4E17B0(t, target, attack, &events)
			var reason string
			r.Unsupported = func(why string, got, attacker, weapon *Object, amount int32, typ object.DamageType) {
				if got != target || attacker != source || weapon != attack || amount != 8 || typ != object.DamageZapRay {
					t.Fatal("unsupported NPC reflection identity")
				}
				reason = why
			}
			switch missing {
			case "direction":
				r.BlockDirection = nil
			case "reflect":
				r.ProjectileReflect = nil
			case "clear":
				r.ClearOwner = nil
			case "set":
				r.SetOwner = nil
			case "change":
				r.ChangeOwner = nil
			case "point":
				r.PointFX = nil
			case "audio":
				r.Audio = nil
			}
			if handled, result := PlayerDamageNative4E17B0(target, source, attack, 8, object.DamageZapRay, r); handled || result || reason == "" {
				t.Fatalf("missing service=%t/%t reason=%q", handled, result, reason)
			}
			if *target != beforeTarget || *target.UpdateDataMonster() != beforeUpdate || *attack != beforeAttack || (len(events) != 0 && !slices.Equal(events, []string{"direction"})) {
				t.Fatal("unsupported NPC reflection committed state or an effect")
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCReflectSkipsUnusedOwnerServices(t *testing.T) {
	target, source := defaultDamageElectricSelfFixture4E0B30(t, false, true)
	target.Buffs = 1 << playerDamageReflectEnchant4E17B0
	attack := &Object{ObjClass: object.ClassMissile, ObjSubClass: 0x40}
	var events []string
	r := npcReflectRuntime4E17B0(t, target, attack, &events)
	r.ClearOwner, r.SetOwner, r.ChangeOwner, r.PointFX = nil, nil, nil, nil
	r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }
	if handled, result := PlayerDamageNative4E17B0(target, source, attack, 8, object.DamageImpact, r); !handled || result || !slices.Equal(events, []string{"direction", "reflect", "audio"}) {
		t.Fatalf("subclass 0x40 required an unused service: %t/%t events=%v", handled, result, events)
	}
}
