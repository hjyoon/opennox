package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func damageMeleeRuntimeFixture4E17B0(t *testing.T) PlayerDamageRuntime4E17B0 {
	t.Helper()
	world := damageMeleeWorldRuntime4E0B30(t)
	return PlayerDamageRuntime4E17B0{
		Frame: world.Frame,
		DefaultDamage: func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
			return DefaultDamageWorld4E0B30(target, source, weapon, damage, typ, world)
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("native ordinary melee rejected: %s", reason)
		},
	}
}

// Keep the scalar view used by carry assertions, but allocate the complete
// eight-byte update record read by the real EquipDamage implementation.
func damageArmorCarryFixture4E17B0(value float32) *float32 {
	update := &WeaponArmorUpdateData{Field0: math.Float32bits(value)}
	return (*float32)(unsafe.Pointer(update))
}

func damageMeleeArmorFixture4E17B0(target *Object, armorValue, carry float32) *Object {
	if target.Class().Has(object.ClassPlayer) {
		target.UpdateDataPlayer().Field57 = math.Float32bits(armorValue)
		target.UpdateDataPlayer().Field21 = math.Float32bits(carry)
	} else {
		target.UpdateDataMonster().Field518 = math.Float32bits(armorValue)
		target.UpdateDataMonster().Field1 = math.Float32bits(carry)
	}
	itemCarry := damageArmorCarryFixture4E17B0(0.4)
	armor := &Object{
		ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		Damage:   unsafe.Pointer(new(byte)),
		InitData: unsafe.Pointer(&ModifierInitData{}), UpdateData: unsafe.Pointer(itemCarry),
		HealthData: &HealthData{Cur: 25, Max: 25},
	}
	target.InvFirstItem = armor
	return armor
}

func damageMeleeArmorRuntime4E17B0(r *PlayerDamageRuntime4E17B0, armor *Object, value float32) {
	r.ItemArmorValue = func(*Object) float32 { return value }
	r.CanDamageArmor = func(item *Object) bool { return item == armor }
	r.DamageArmor = func(item, _, _ *Object, amount int32, _ object.DamageType) bool {
		item.HealthData.Cur -= uint16(amount)
		return true
	}
}

func damageMeleeMarker4E17B0(target *Object) (uint32, uint32, float32) {
	if target.Class().Has(object.ClassPlayer) {
		ud := target.UpdateDataPlayer()
		return ud.Field76, ud.Field75, math.Float32frombits(ud.Field21)
	}
	ud := target.UpdateDataMonster()
	return ud.Field547, ud.Field546, math.Float32frombits(ud.Field1)
}

func TestPlayerDamageMeleeNative4E17B0ArmoredMatrix(t *testing.T) {
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
				{"UnarmedClaw", 0, 0, object.DamageClaw},
				{"UnarmedCrush", 0, 0, object.DamageCrush},
			} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/%s", fromPlayer, toPlayer, attack.name), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, toPlayer)
					source := damageMeleeUnitFixture4E17B0(t, fromPlayer)
					source.TypeInd = 88
					var weapon *Object
					if attack.class != 0 {
						weapon = &Object{TypeInd: 777, ObjClass: attack.class, ObjSubClass: attack.subclass, InitData: unsafe.Pointer(&ModifierInitData{})}
					}
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					r := damageMeleeRuntimeFixture4E17B0(t)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					if handled, result := PlayerDamageMeleeNative4E17B0(target, source, weapon, 9, attack.typ, r); !handled || !result {
						t.Fatalf("handled=%t result=%t", handled, result)
					}
					wantDamage, wantCarry := uint16(5), float32(-0.1)
					if attack.typ == object.DamageCrush {
						wantDamage, wantCarry = 7, 0.15
					}
					marker, markerType, carry := damageMeleeMarker4E17B0(target)
					wantType := uint32(777)
					if weapon == nil {
						wantType = uint32(source.TypeInd)
					}
					if target.HealthData.Cur != 200-wantDamage || armor.HealthData.Cur != 25-(9-wantDamage) ||
						marker != 1 || markerType != wantType || math.Abs(float64(carry-wantCarry)) > 1e-6 {
						t.Fatalf("HP=%d armor=%d marker=%d/%d carry=%g", target.HealthData.Cur, armor.HealthData.Cur, marker, markerType, carry)
					}
					if math.Abs(float64(*(*float32)(armor.UpdateData)-0.4)) > 1e-6 {
						t.Fatal("armor fractional durability was lost")
					}
				})
			}
		}
	}
}

func TestPlayerDamageMeleeNative4E17B0ArmorPrefixOrder(t *testing.T) {
	for _, player := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-%t", player), func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, player)
			source := damageMeleeUnitFixture4E17B0(t, !player)
			weapon := &Object{TypeInd: 333, ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(&ModifierInitData{})}
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			armor.InitDataModifier().Modifiers[1] = modifier
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			var events []string
			r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
			r.ApplyArmorDefend = func(m *ModifierEff, item, owner, effective, attacker *Object, amount *float32) bool {
				marker, markerType, carry := damageMeleeMarker4E17B0(target)
				if m != modifier || item != armor || owner != target || effective != weapon || attacker != source ||
					*amount != 4 || marker != 1 || markerType != 333 || math.Abs(float64(carry+0.1)) > 1e-6 {
					t.Fatalf("armor modifier preceded native hit/carry stores: marker=%d/%d carry=%g amount=%g", marker, markerType, carry, *amount)
				}
				events = append(events, "armor-defend")
				*amount *= 0.5
				return true
			}
			r.DamageArmor = func(item, attacker, effective *Object, amount int32, typ object.DamageType) bool {
				if item != armor || attacker != source || effective != weapon || amount != 2 || typ != object.DamageBlade {
					t.Fatal("armor durability lost native arguments")
				}
				events = append(events, "armor-damage")
				item.HealthData.Cur -= uint16(amount)
				return true
			}
			r.ReportArmorHealth = func(owner, item *Object, before, after uint16) {
				if !player || owner != target || item != armor || before != 25 || after != 23 {
					t.Fatal("bad armor report")
				}
				events = append(events, "armor-report")
			}
			r.DefaultDamage = func(owner, attacker, effective *Object, amount int32, typ object.DamageType) bool {
				if owner != target || attacker != source || effective != weapon || amount != 5 || typ != object.DamageBlade {
					t.Fatal("bad default damage")
				}
				events = append(events, "default")
				return true
			}
			if h, result := PlayerDamageMeleeNative4E17B0(target, source, weapon, 9, object.DamageBlade, r); !h || !result {
				t.Fatal("melee prefix failed")
			}
			want := []string{"armor-defend", "armor-damage"}
			if player {
				want = append(want, "armor-report")
			}
			want = append(want, "default")
			if !slices.Equal(events, want) {
				t.Fatalf("events=%v", events)
			}
		})
	}
}

func TestPlayerDamageMeleeNative4E17B0RoundAndGodQuest(t *testing.T) {
	for _, tc := range []struct {
		name               string
		player, god, quest bool
		want               int32
	}{
		{"player", true, false, false, 7}, {"NPC", false, false, false, 7},
		{"player GodMode", true, true, false, 0}, {"NPC ignores GodMode", false, true, false, 7},
		{"Quest tie rounds even", true, false, true, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, tc.player)
			source := damageMeleeUnitFixture4E17B0(t, !tc.player)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			r.GodMode = func() bool { return tc.god }
			r.QuestMode = func() bool { return tc.quest }
			r.QuestDamageScale = func() float32 { return 0.5 }
			var amount int32
			r.DefaultDamage = func(_, _, _ *Object, damage int32, _ object.DamageType) bool { amount = damage; return true }
			if h, result := PlayerDamageMeleeNative4E17B0(target, source, nil, 9, object.DamageCrush, r); !h || !result {
				t.Fatal("melee failed")
			}
			if amount != tc.want || armor.HealthData.Cur != 23 {
				t.Fatalf("HP damage=%d want=%d armor=%d", amount, tc.want, armor.HealthData.Cur)
			}
		})
	}
	for _, player := range []bool{false, true} {
		target := damageMeleeUnitFixture4E17B0(t, player)
		source := damageMeleeUnitFixture4E17B0(t, !player)
		armor := damageMeleeArmorFixture4E17B0(target, 0.75, 0)
		r := damageMeleeRuntimeFixture4E17B0(t)
		damageMeleeArmorRuntime4E17B0(&r, armor, 0.75)
		r.QuestMode = func() bool { return true }
		r.QuestDamageScale = func() float32 { return 0 }
		var amounts []int32
		r.DefaultDamage = func(_, _, _ *Object, damage int32, _ object.DamageType) bool {
			amounts = append(amounts, damage)
			return true
		}
		for i := 0; i < 2; i++ {
			PlayerDamageMeleeNative4E17B0(target, source, nil, 2, object.DamageClaw, r)
			_, _, carry := damageMeleeMarker4E17B0(target)
			if carry != float32(1-i)*0.5 {
				t.Fatalf("half-integer carry=%g", carry)
			}
		}
		if !slices.Equal(amounts, []int32{1, 1}) || armor.HealthData.Cur != 22 {
			t.Fatalf("minimum damage/carry=%v armor=%d", amounts, armor.HealthData.Cur)
		}
	}
}

func TestPlayerDamageMeleeNative4E17B0BlockDurabilityAndStates(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, kind := range []string{"shield", "greatsword", "staff"} {
			for _, destroyed := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-%t/%s/destroyed-%t", player, kind, destroyed), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, player)
					source := damageMeleeUnitFixture4E17B0(t, !player)
					weapon := &Object{TypeInd: 999, ObjClass: object.ClassWeapon, PrevPos: types.Pointf{X: 33, Y: 44}}
					mask, armorMask, sound := uint32(2), uint32(0x1000000), 878
					if kind == "greatsword" {
						mask, armorMask, sound = 0x400, 0, 890
					}
					if kind == "staff" {
						mask, armorMask, sound = 0x8000, 0, 894
					}
					item := &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped}
					decoy := &Object{ObjSubClass: object.SubClass(mask), InvNextItem: item}
					later := &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped}
					item.InvNextItem, target.InvFirstItem = later, decoy
					weaponMask := mask
					if kind == "shield" {
						weaponMask = 0
					}
					if player {
						ud := target.UpdateDataPlayer()
						ud.Player.WeaponEquip, ud.Player.ArmorEquip = weaponMask, armorMask
						if kind == "shield" {
							ud.State = PlayerState16
						}
					} else {
						ud := target.UpdateDataMonster()
						ud.WeaponEquipFlags, ud.ArmorEquipFlags = weaponMask, armorMask
						ud.AIStack[0].Action = uint32(ai.ACTION_GUARD)
						if kind == "shield" {
							ud.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
						}
					}
					var events []string
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.BlockSourceExcluded = func(got *Object) bool { return false }
					r.BlockDirection = func(got *Object, pos types.Pointf) bool {
						if got != target || pos != weapon.PrevPos {
							t.Fatal("block did not use attack PrevPos")
						}
						return true
					}
					r.Audio = func(got int, unit *Object) {
						marker, markerType, _ := damageMeleeMarker4E17B0(target)
						if got != sound || unit != target || marker != 1 || markerType != 999 {
							t.Fatal("block audio preceded marker")
						}
						events = append(events, "audio")
					}
					r.BlockDamagePercent = func() float64 { return 0.2 }
					r.PlayerSetState = func(unit *Object, state PlayerState) bool {
						if unit != target {
							t.Fatal("wrong state target")
						}
						events = append(events, fmt.Sprintf("state-%d", state))
						unit.UpdateDataPlayer().State = state
						return true
					}
					r.Melee.RandomInt = func(min, max int) int {
						if min != 18 || max != 20 {
							t.Fatal("wrong sword RNG range")
						}
						events = append(events, "rng")
						return 19
					}
					r.Melee.MonsterBlockAction = func(unit *Object) { events = append(events, "push-23") }
					r.Melee.MonsterPopBlockAction = func(unit *Object) { events = append(events, "pop") }
					can := func(got *Object) bool { return got == item }
					apply := func(got, owner, attacker, effective *Object, amount float32, typ object.DamageType) bool {
						if got != item || owner != target || attacker != source || effective != weapon || amount != float32(1.8) || typ != object.DamageBlade {
							t.Fatal("wrong block durability args")
						}
						events = append(events, "durability")
						if destroyed {
							got.ObjFlags |= object.FlagDestroyed
						}
						return true
					}
					r.CanDamageBlockItem, r.Melee.CanDamageBlockWeapon = can, can
					r.DamageBlockItem, r.Melee.DamageBlockWeapon = apply, apply
					r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
						t.Fatal("blocked melee reached HP damage")
						return false
					}
					if h, result := PlayerDamageMeleeNative4E17B0(target, source, weapon, 9, object.DamageBlade, r); !h || result {
						t.Fatalf("blocked result=%t/%t", h, result)
					}
					want := []string{"audio"}
					if kind != "shield" {
						if !player {
							want = append(want, "push-23")
						} else if kind == "staff" {
							want = append(want, "state-21")
						} else {
							want = append(want, "rng", "state-19")
						}
					}
					want = append(want, "durability")
					if destroyed {
						if player {
							want = append(want, "state-13")
						} else {
							want = append(want, "pop")
						}
					}
					if target.HealthData.Cur != 200 || !slices.Equal(events, want) {
						t.Fatalf("HP=%d events=%v want=%v", target.HealthData.Cur, events, want)
					}
				})
			}
		}
	}
}

func TestPlayerDamageMonsterBlockReady534340(t *testing.T) {
	target := damageMeleeUnitFixture4E17B0(t, false)
	ud := target.UpdateDataMonster()
	for action := 0; action < 72; action++ {
		ud.AIStack[0].Action = uint32(action)
		want := slices.Contains([]int{0, 1, 4, 23, 25, 26, 27}, action)
		if got := playerDamageMonsterBlockReady534340(target); got != want {
			t.Fatalf("action=%d ready=%t want=%t", action, got, want)
		}
	}
	ud.AIStackInd = -1
	if playerDamageMonsterBlockReady534340(target) {
		t.Fatal("invalid AI stack can block")
	}
}

func TestPlayerDamageMeleeNative4E17B0UnsupportedBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*Object, *PlayerDamageRuntime4E17B0)
	}{
		{"missing armor data", func(target *Object, _ *PlayerDamageRuntime4E17B0) { target.InvFirstItem.InitData = nil }},
		{"unported armor modifier", func(target *Object, r *PlayerDamageRuntime4E17B0) {
			target.InvFirstItem.InitDataModifier().Modifiers[1] = &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			r.ApplyArmorDefend = func(_ *ModifierEff, _, _, _, _ *Object, _ *float32) bool {
				t.Fatal("unvalidated modifier executed")
				return false
			}
		}},
		{"missing default tail", func(_ *Object, r *PlayerDamageRuntime4E17B0) { r.DefaultDamage = nil }},
		{"missing block direction", func(target *Object, _ *PlayerDamageRuntime4E17B0) {
			target.UpdateDataPlayer().Player.WeaponEquip = 0x400
		}},
		{"missing block item", func(target *Object, r *PlayerDamageRuntime4E17B0) {
			target.UpdateDataPlayer().Player.WeaponEquip = 0x400
			r.BlockSourceExcluded = func(*Object) bool { return false }
			r.BlockDirection = func(*Object, types.Pointf) bool { return true }
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			source := damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			tc.setup(target, &r)
			target.UpdateDataPlayer().Field76, target.UpdateDataPlayer().Field75 = 67, 68
			targetBefore, updateBefore, armorBefore := *target, *target.UpdateDataPlayer(), *armor
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageMeleeNative4E17B0(target, source, &Object{ObjClass: object.ClassWeapon}, 9, object.DamageBlade, r)
			if h || result || reason == "" || *target != targetBefore || *target.UpdateDataPlayer() != updateBefore || *armor != armorBefore || armor.HealthData.Cur != 25 {
				t.Fatalf("unsupported branch mutated state: handled=%t result=%t reason=%q", h, result, reason)
			}
		})
	}
}

func TestPlayerDamageMeleeNative4E17B0EarlyGatesAndReflect(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		setup                   func(*Object, *Object, *PlayerDamageRuntime4E17B0)
		wantHandled, wantResult bool
		wantAudio               int
	}{
		{"no update", func(o, _ *Object, _ *PlayerDamageRuntime4E17B0) { o.ObjFlags |= object.FlagNoUpdate }, true, false, 0},
		{"dead", func(o, _ *Object, _ *PlayerDamageRuntime4E17B0) { o.ObjFlags |= object.FlagDead }, true, false, 0},
		{"invulnerable", func(o, _ *Object, _ *PlayerDamageRuntime4E17B0) {
			o.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
		}, true, true, 71},
		{"observer", func(o, _ *Object, _ *PlayerDamageRuntime4E17B0) { o.UpdateDataPlayer().Player.Field3680 |= 1 }, true, false, 0},
		{"Coop owned source", func(o, s *Object, r *PlayerDamageRuntime4E17B0) {
			s.ObjOwner = o
			r.CoopMode = func() bool { return true }
		}, true, false, 0},
		{"possession unported", func(o, _ *Object, _ *PlayerDamageRuntime4E17B0) {
			p := o.UpdateDataPlayer().Player
			p.Field3680 |= 2
			p.CameraFollowObj = &Object{}
		}, false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			source := damageMeleeUnitFixture4E17B0(t, false)
			r := damageMeleeRuntimeFixture4E17B0(t)
			tc.setup(target, source, &r)
			before, playerBefore := *target.UpdateDataPlayer(), *target.UpdateDataPlayer().Player
			var audio int
			r.Audio = func(id int, _ *Object) { audio = id }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("early gate reached DefaultDamage")
				return false
			}
			r.Unsupported = func(string, *Object, *Object, *Object, int32, object.DamageType) {}
			h, result := PlayerDamageMeleeNative4E17B0(target, source, nil, 9, object.DamageClaw, r)
			if h != tc.wantHandled || result != tc.wantResult || audio != tc.wantAudio || target.HealthData.Cur != 200 ||
				*target.UpdateDataPlayer() != before || *target.UpdateDataPlayer().Player != playerBefore {
				t.Fatalf("early gate=%t/%t audio=%d HP=%d", h, result, audio, target.HealthData.Cur)
			}
		})
	}
	for _, player := range []bool{false, true} {
		target := damageMeleeUnitFixture4E17B0(t, player)
		source := damageMeleeUnitFixture4E17B0(t, !player)
		target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
		r := damageMeleeRuntimeFixture4E17B0(t)
		if h, result := PlayerDamageMeleeNative4E17B0(target, source, nil, 9, object.DamageClaw, r); !h || !result || target.HealthData.Cur != 191 {
			t.Fatal("Reflect Shield incorrectly intercepted non-missile ordinary melee")
		}
	}
}

func TestPlayerDamageMeleeNative4E17B0BlockAdmission(t *testing.T) {
	for _, tc := range []struct {
		name                              string
		state                             PlayerState
		weaponFlags, armorFlags           uint32
		typ                               object.DamageType
		front, excluded, walking, berserk bool
		wantSound                         int
	}{
		{"staff idle", PlayerState13, 0x8000, 0, object.DamageBlade, true, false, false, false, 894},
		{"staff repeated block", PlayerState21, 0x8000, 0, object.DamageBlade, true, false, false, false, 894},
		{"staff sword block 18", PlayerState18, 0x8000, 0, object.DamageBlade, true, false, false, false, 0},
		{"staff sword block 19", PlayerState19, 0x8000, 0, object.DamageBlade, true, false, false, false, 0},
		{"staff sword block 20", PlayerState20, 0x8000, 0, object.DamageBlade, true, false, false, false, 0},
		{"staff walking option off", PlayerState0, 0x8000, 0, object.DamageBlade, true, false, false, false, 0},
		{"staff walking option on", PlayerState0, 0x8000, 0, object.DamageBlade, true, false, true, false, 894},
		{"sword attacking", PlayerState1, 0x400, 0, object.DamageBlade, true, false, false, false, 0},
		{"sword crush", PlayerState13, 0x400, 0, object.DamageCrush, true, false, false, false, 0},
		{"rear shield", PlayerState16, 0, 0x1000000, object.DamageBlade, false, false, false, false, 0},
		{"excluded shield source", PlayerState16, 0, 0x1000000, object.DamageBlade, true, true, false, false, 0},
		{"berserk option off", PlayerState1, 0, 0x1000000, object.DamageCrush, true, false, false, false, 0},
		{"berserk option on", PlayerState1, 0, 0x1000000, object.DamageCrush, true, false, false, true, 878},
		{"shield precedence", PlayerState16, 0x400, 0x1000000, object.DamageBlade, true, false, false, false, 878},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			source := damageMeleeUnitFixture4E17B0(t, false)
			ud := target.UpdateDataPlayer()
			ud.State, ud.Player.WeaponEquip, ud.Player.ArmorEquip = tc.state, tc.weaponFlags, tc.armorFlags
			// Non-blocking cases enter DefaultDamage's flags-only 004E1320
			// traversal too, so the equipped fixture needs a valid slot base.
			target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, ObjSubClass: object.SubClass(tc.weaponFlags | 2),
				InitData: unsafe.Pointer(&ModifierInitData{})}
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.BlockDirection = func(*Object, types.Pointf) bool { return tc.front }
			r.BlockSourceExcluded = func(*Object) bool { return tc.excluded }
			r.BerserkShieldBlock = func(*Object) bool { return tc.berserk }
			r.Melee.StaffWalkingBlock = func() bool { return tc.walking }
			r.BlockDamagePercent = func() float64 { return 0.2 }
			r.PlayerSetState = func(o *Object, state PlayerState) bool { o.UpdateDataPlayer().State = state; return true }
			r.Melee.RandomInt = func(int, int) int { return 18 }
			r.CanDamageBlockItem = func(*Object) bool { return true }
			r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
			r.DamageBlockItem = func(_, _, _, _ *Object, _ float32, _ object.DamageType) bool { return true }
			r.Melee.DamageBlockWeapon = r.DamageBlockItem
			var sound int
			r.Audio = func(id int, _ *Object) { sound = id }
			h, result := PlayerDamageMeleeNative4E17B0(target, source, &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(&ModifierInitData{})}, 9, tc.typ, r)
			wantHP := uint16(200)
			if tc.wantSound == 0 {
				wantHP = 191
			}
			if !h || result != (tc.wantSound == 0) || sound != tc.wantSound || target.HealthData.Cur != wantHP {
				t.Fatalf("block=%t/%t sound=%d HP=%d", h, result, sound, target.HealthData.Cur)
			}
		})
	}
}
