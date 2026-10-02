package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func damageMeleeUnitFixture4E17B0(t *testing.T, player bool) *Object {
	t.Helper()
	unit := &Object{
		TypeInd: 71, ObjFlags: object.FlagActive | object.FlagEnabled,
		HealthData: &HealthData{Cur: 200, Max: 200, Field2: 200},
		PrevPos:    types.Pointf{X: 32.5, Y: -18.25},
	}
	if player {
		unit.ObjClass = object.ClassPlayer
		unit.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &Player{}, State: PlayerState13})
	} else {
		unit.ObjClass = object.ClassMonster
		unit.ObjSubClass = 0x11012 // Original NPC subclass, not ordinary Spider.
		unit.UpdateData = unsafe.Pointer(&MonsterUpdateData{})
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= uintptr(^uint32(0)) {
		t.Fatal("native-width melee fixture is not above 4 GiB")
	}
	return unit
}

func damageMeleeWorldRuntime4E0B30(t *testing.T) DefaultDamageWorldRuntime4E0B30 {
	t.Helper()
	return DefaultDamageWorldRuntime4E0B30{
		Frame: func() uint32 { return 1400 }, GameplayFlag1: func() bool { return true },
		IsEnemy:            func(*Object, *Object) bool { return true },
		BuffOff:            func(*Object, EnchantID) {},
		MonsterHasHitSound: func(*Object) bool { return false },
		PlayerSetState: func(unit *Object, state PlayerState) bool {
			unit.UpdateDataPlayer().State = state
			return true
		},
		DamageClear: func(unit *Object, damage int32) {
			if damage >= int32(unit.HealthData.Cur) {
				unit.HealthData.Cur = 0
			} else {
				unit.HealthData.Cur -= uint16(damage)
			}
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("ordinary melee rejected: %s", reason)
		},
	}
}

func TestDefaultDamageWorld4E0B30UnitMeleeMatrix(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, toPlayer := range []bool{false, true} {
			for _, attack := range []struct {
				name     string
				class    object.Class
				subclass object.SubClass
				typ      object.DamageType
			}{
				{"Sword", object.ClassWeapon, object.SubClass(object.WeaponSword), object.DamageBlade},
				{"MorningStar", object.ClassWeapon, object.SubClass(object.WeaponMace), object.DamageCrush},
				{"WarHammer", object.ClassWeapon, object.SubClass(object.WeaponHammer), object.DamageCrush},
				{"WoodenStaff", object.ClassWand, 0, object.DamageBlade},
				{"Unarmed", 0, 0, object.DamageClaw},
				{"UnarmedCrush", 0, 0, object.DamageCrush},
			} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/%s", fromPlayer, toPlayer, attack.name), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, toPlayer)
					source := damageMeleeUnitFixture4E17B0(t, fromPlayer)
					var weapon *Object
					if attack.class != 0 {
						weapon = &Object{TypeInd: 777, ObjClass: attack.class, ObjSubClass: attack.subclass}
					}
					r := damageMeleeWorldRuntime4E0B30(t)
					if !DefaultDamageWorld4E0B30(target, source, weapon, 25, attack.typ, r) {
						t.Fatal("ordinary melee tail returned false")
					}
					attribution := weapon
					if attribution == nil {
						attribution = source
					}
					if target.HealthData.Cur != 175 || target.Obj130 != attribution || target.Pos132 != source.PrevPos ||
						target.Field131 != uint32(attack.typ) || target.Frame134 != 1400 {
						t.Fatalf("melee result: HP=%d object=%p pos=%v type=%d frame=%d", target.HealthData.Cur,
							target.Obj130, target.Pos132, target.Field131, target.Frame134)
					}
					if toPlayer {
						if target.UpdateDataPlayer().State != PlayerState30 {
							t.Fatal("player hurt animation state was not set")
						}
					} else {
						ud := target.UpdateDataMonster()
						if ud.Field547 != 1 || ud.Field546 != uint32(attribution.TypeInd) || !ud.StatusFlags.Has(object.MonStatusInjured) {
							t.Fatalf("NPC injured attribution: %d/%d flags=%x", ud.Field547, ud.Field546, ud.StatusFlags)
						}
					}
					if !fromPlayer && source.UpdateDataMonster().Field130 != 1400 {
						t.Fatal("NPC successful-hit timestamp was not updated")
					}
					t.Logf("native melee pointers target=%p source=%p weapon=%p HP=200->175", target, source, weapon)
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30UnitMeleeFriendlyRules(t *testing.T) {
	for _, toPlayer := range []bool{false, true} {
		for _, tc := range []struct {
			name       string
			weapon     object.SubClass
			unarmed    bool
			gameplay   bool
			wantDamage bool
		}{
			{"friendly Sword", object.SubClass(object.WeaponSword), false, true, false},
			{"friendly Mace", object.SubClass(object.WeaponMace), false, true, false},
			{"War Hammer melee exception", object.SubClass(object.WeaponHammer), false, true, true},
			{"unarmed has no weapon gate", 0, true, true, true},
			{"campaign still rejects Hammer owner", object.SubClass(object.WeaponHammer), false, false, false},
			{"campaign still rejects unarmed owner", 0, true, false, false},
		} {
			t.Run(fmt.Sprintf("player-target-%t/%s", toPlayer, tc.name), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, toPlayer)
				source := damageMeleeUnitFixture4E17B0(t, true)
				weapon := &Object{ObjClass: object.ClassWeapon, ObjSubClass: tc.weapon}
				typ := object.DamageCrush
				if tc.unarmed {
					weapon, typ = nil, object.DamageClaw
				}
				r := damageMeleeWorldRuntime4E0B30(t)
				r.GameplayFlag1 = func() bool { return tc.gameplay }
				r.IsEnemy = func(*Object, *Object) bool { return false }
				DefaultDamageWorld4E0B30(target, source, weapon, 10, typ, r)
				want := uint16(200)
				if tc.wantDamage {
					want = 190
				}
				if target.HealthData.Cur != want {
					t.Fatalf("friendly HP=%d, want %d", target.HealthData.Cur, want)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerMeleeTailOrder(t *testing.T) {
	target := damageMeleeUnitFixture4E17B0(t, true)
	source := damageMeleeUnitFixture4E17B0(t, false)
	weapon := &Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(object.WeaponSword)}
	pre := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	weapon.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{pre}})
	late := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped}
	armor.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, late}})
	target.InvFirstItem = armor
	target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
	target.DamageSound = unsafe.Pointer(new(byte))
	var events []string
	r := damageMeleeWorldRuntime4E0B30(t)
	r.BuffOff = func(*Object, EnchantID) { events = append(events, "invisible-off") }
	r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == late }
	r.ApplyLateDefend = func(_ *ModifierEff, item, got, gotWeapon, gotSource *Object, damage int32, typ object.DamageType) int32 {
		if item != armor || got != target || gotWeapon != weapon || gotSource != source || damage != 30 || typ != object.DamageBlade {
			t.Fatal("late defend lost native melee arguments")
		}
		events = append(events, "late-defend")
		return damage - 2
	}
	r.CanApplyPreDamage = func(m *ModifierEff) bool { return m == pre }
	r.ApplyPreDamage = func(_ *ModifierEff, gotWeapon, gotSource, gotTarget *Object, damage *int32) {
		if gotWeapon != weapon || gotSource != source || gotTarget != target || *damage != 28 || target.Obj130 != weapon {
			t.Fatal("pre-damage lost arguments or ran before attribution")
		}
		events = append(events, "pre-damage")
		*damage += 3
	}
	r.MonsterHasHitSound = func(*Object) bool { events = append(events, "hit-sound"); return false }
	r.PlayerDamageSoundC = target.DamageSound
	r.PlayerDamageSound = func(*Object, *Object) { events = append(events, "player-sound") }
	r.GameBallOnDamage = func(*Object, *Object, int32) { events = append(events, "ball") }
	r.PlayerSetState = func(_ *Object, state PlayerState) bool {
		if state != PlayerState30 {
			t.Fatal("incorrect hurt state")
		}
		events = append(events, "hurt")
		return true
	}
	r.ShieldReduce = func(got *Object, damage *int32, typ object.DamageType, gotSource *Object) {
		if got != target || gotSource != weapon || *damage != 31 || typ != object.DamageBlade || source.UpdateDataMonster().Field130 != 1400 {
			t.Fatal("Shield/timestamp order is incorrect")
		}
		events = append(events, "shield")
		*damage = 7
	}
	r.DamageClear = func(_ *Object, damage int32) {
		if damage != 7 {
			t.Fatal("melee ignored Shield reduction")
		}
		events = append(events, "clear")
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 30, object.DamageBlade, r) {
		t.Fatal("player melee tail returned false")
	}
	want := []string{"invisible-off", "late-defend", "pre-damage", "hit-sound", "player-sound", "ball", "hurt", "shield", "clear"}
	if !slices.Equal(events, want) {
		t.Fatalf("tail events=%v, want %v", events, want)
	}
}
