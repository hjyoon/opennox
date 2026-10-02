package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDefaultDamageWorld4E0B30SimpleCrushMatrix(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, targetKind := range []string{"Player", "NPC", "Monster"} {
			for i, name := range []string{"SmallFist", "MediumFist", "LargeFist"} {
				t.Run(fmt.Sprintf("player-source-%t/%s/%s", fromPlayer, targetKind, name), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, targetKind == "Player")
					if targetKind == "Monster" {
						target.ObjSubClass = 0x202 // Ordinary Spider, not the NPC defense callback.
					}
					source := damageMeleeUnitFixture4E17B0(t, fromPlayer)
					fist := &Object{TypeInd: uint16(777 + i), ObjClass: object.ClassSimple, ObjOwner: source,
						PrevPos: types.Pointf{X: 45.5, Y: 67.25}}
					r := damageMeleeWorldRuntime4E0B30(t)
					reason := ""
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					if !DefaultDamageWorld4E0B30(target, source, fist, 25, object.DamageCrush, r) ||
						target.HealthData.Cur != 175 || reason != "" {
						t.Fatalf("SIMPLE CRUSH tail HP=200->%d reason=%q", target.HealthData.Cur, reason)
					}
					if target.Obj130 != fist || target.Pos132 != fist.PrevPos || target.Field131 != 2 || target.Frame134 != 1400 {
						t.Fatal("Fist attribution/position/type/frame did not reach the default tail")
					}
					if targetKind == "Player" {
						if target.UpdateDataPlayer().State != PlayerState30 {
							t.Fatal("Fist must retain the player hurt-animation tail")
						}
					} else if ud := target.UpdateDataMonster(); ud.Field547 != 1 || ud.Field546 != uint32(fist.TypeInd) || !ud.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("Fist must retain the monster hurt/weapon attribution")
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30SimpleCrushFriendlyAndShock(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, gameplay := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-%t/gameplay-%t", player, gameplay), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, player)
				source := damageMeleeUnitFixture4E17B0(t, true)
				fist := &Object{ObjClass: object.ClassSimple, ObjOwner: source}
				target.Buffs |= 1 << defaultDamageShockEnchant4E0B30
				r := damageMeleeWorldRuntime4E0B30(t)
				r.GameplayFlag1 = func() bool { return gameplay }
				r.IsEnemy = func(*Object, *Object) bool { return false }
				r.CallDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("SIMPLE Fist must not retaliate with melee Shock")
					return false
				}
				DefaultDamageWorld4E0B30(target, source, fist, 10, object.DamageCrush, r)
				want := uint16(200)
				if gameplay {
					want = 190 // 004E1400 is false; only the campaign owner gate applies.
				}
				if target.HealthData.Cur != want || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
					t.Fatalf("Fist campaign/friendly/Shock: HP=%d want=%d", target.HealthData.Cur, want)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30SimpleCrushServices(t *testing.T) {
	for _, missing := range []string{"clear", "buff", "enemy", "sound", "hurt"} {
		t.Run(missing, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			source := damageMeleeUnitFixture4E17B0(t, false)
			fist := &Object{ObjClass: object.ClassSimple}
			r := damageMeleeWorldRuntime4E0B30(t)
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			switch missing {
			case "clear":
				r.DamageClear = nil
			case "buff":
				r.BuffOff = nil
			case "enemy":
				r.IsEnemy = nil
			case "sound":
				r.MonsterHasHitSound = nil
			case "hurt":
				r.PlayerSetState = nil
			}
			DefaultDamageWorld4E0B30(target, source, fist, 25, object.DamageCrush, r)
			if reason == "" || target.HealthData.Cur != 200 || target.Obj130 != nil || target.Frame134 != 0 || target.UpdateDataPlayer().State != PlayerState13 {
				t.Fatalf("missing %s was not rejected before damage stores: HP=%d reason=%q", missing, target.HealthData.Cur, reason)
			}
		})
	}
}

func TestPlayerDamageSimpleCrushShape4E17B0(t *testing.T) {
	player := damageMeleeUnitFixture4E17B0(t, true)
	npc := damageMeleeUnitFixture4E17B0(t, false)
	fist := &Object{ObjClass: object.ClassSimple}
	if !playerDamageSimpleCrushShape4E17B0(player, fist, object.DamageCrush) || !playerDamageSimpleCrushShape4E17B0(npc, fist, object.DamageCrush) {
		t.Fatal("stock SIMPLE CRUSH unit shapes were rejected")
	}
	for _, tc := range []struct {
		source, weapon *Object
		typ            object.DamageType
	}{
		{nil, fist, object.DamageCrush}, {player, nil, object.DamageCrush}, {fist, fist, object.DamageCrush},
		{player, fist, object.DamageBlade}, {player, fist, object.DamageZapRay},
		{&Object{ObjClass: object.ClassMonster}, fist, object.DamageCrush},
		{player, &Object{ObjClass: object.ClassSimple | object.ClassMissile}, object.DamageCrush},
		{player, &Object{ObjClass: object.ClassSimple | object.ClassWeapon}, object.DamageCrush},
		{player, &Object{ObjClass: object.ClassSimple | object.ClassWand}, object.DamageCrush},
	} {
		if playerDamageSimpleCrushShape4E17B0(tc.source, tc.weapon, tc.typ) {
			t.Fatalf("unsupported SIMPLE CRUSH shape was admitted: %+v", tc)
		}
	}
}
