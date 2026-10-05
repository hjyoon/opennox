package server

import (
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDefaultDamageWorld4E0B30WeaponlessExplosionMatrix(t *testing.T) {
	for _, sourceKind := range []string{"none", "player", "npc"} {
		for _, npc := range []bool{false, true} {
			for _, tc := range []struct {
				name       string
				damage     int32
				immune     bool
				protection float64
				want       int32
			}{
				{"ordinary", 45, false, 0, 45}, {"resist-even", 9, false, 0.5, 4},
				{"resist-odd", 11, false, 0.5, 6}, {"immune", 9, true, 0, 4},
				{"immune-resist", 11, true, 0.5, 2}, {"signed", -9, true, 0.5, -2},
				{"minimum", 1, true, 0, 1}, {"zero", 0, false, 0, 1},
			} {
				t.Run(fmt.Sprintf("%s/npc-%t/%s", sourceKind, npc, tc.name), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, false)
					if !npc {
						target.ObjSubClass = 0x202 // Ordinary monster, no equipped NPC defense.
					}
					if tc.immune {
						target.ObjSubClass |= 0x400
					}
					var source *Object
					if sourceKind != "none" {
						source = damageMeleeUnitFixture4E17B0(t, sourceKind == "player")
					}
					r := damageMeleeWorldRuntime4E0B30(t)
					r.FireProtection = func(*Object) float64 { return tc.protection }
					if !DefaultDamageWorld4E0B30(target, source, nil, tc.damage, object.DamageExplosion, r) {
						t.Fatal("weapon-less explosion returned false")
					}
					position := types.Pointf{}
					if source != nil {
						position = source.PrevPos
					}
					ud := target.UpdateDataMonster()
					if target.HealthData.Cur != uint16(200-tc.want) || target.Obj130 != source || target.Pos132 != position ||
						target.Field131 != uint32(object.DamageExplosion) || target.Frame134 != 1400 ||
						ud.Field547 != 2 || ud.Field546 != uint32(object.DamageExplosion) ||
						!ud.StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) {
						t.Fatalf("weapon-less explosion HP=%d want=%d metadata=%p/%v/%d/%d latch=%d/%d flags=%x",
							target.HealthData.Cur, 200-tc.want, target.Obj130, target.Pos132, target.Field131, target.Frame134,
							ud.Field547, ud.Field546, ud.StatusFlags)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessExplosionOwnerGate(t *testing.T) {
	for _, sourcePlayer := range []bool{false, true} {
		for _, gameplay := range []bool{false, true} {
			for _, enemy := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-%t/gameplay-%t/enemy-%t", sourcePlayer, gameplay, enemy), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, false)
					target.ObjSubClass = 0x202
					source := damageMeleeUnitFixture4E17B0(t, sourcePlayer)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.GameplayFlag1 = func() bool { return gameplay }
					r.IsEnemy = func(*Object, *Object) bool { return enemy }
					r.FireProtection = func(*Object) float64 { return 0 }
					DefaultDamageWorld4E0B30(target, source, nil, 9, object.DamageExplosion, r)
					want := uint16(191)
					if !gameplay && !enemy {
						want = 200 // Original owner/campaign gate remains in force.
					}
					if target.HealthData.Cur != want {
						t.Fatalf("HP=%d want=%d", target.HealthData.Cur, want)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessExplosionOrderAndShield(t *testing.T) {
	for _, sourcePresent := range []bool{false, true} {
		t.Run(fmt.Sprintf("source-%t", sourcePresent), func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, false)
			target.ObjSubClass = 0x602 // Fire-immune ordinary monster.
			target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			var source *Object
			if sourcePresent {
				source = damageMeleeUnitFixture4E17B0(t, true)
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.FireProtection = func(*Object) float64 { events = append(events, "fire"); return 0.5 }
			r.Audio = func(id int, unit *Object) {
				if id != 104 || unit != target {
					t.Fatalf("unexpected explosion audio %d/%p", id, unit)
				}
				events = append(events, "fire-sound")
			}
			r.BuffOff = func(*Object, EnchantID) { events = append(events, "buff") }
			r.DefaultDamageSound = func(unit, by *Object) {
				if unit != target || by != source || target.Obj130 != source || target.Frame134 != 1400 {
					t.Fatal("damage sound preceded native attribution/frame")
				}
				events = append(events, "damage-sound")
			}
			r.ShieldReduce = func(unit *Object, damage *int32, typ object.DamageType, by *Object) {
				if unit != target || *damage != 2 || typ != object.DamageExplosion || by != source {
					t.Fatal("Shield did not receive post-immunity/protection damage and original source")
				}
				events = append(events, "shield")
				*damage = 0 // External Shield callback, not a production hit result.
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("zero post-Shield damage must not clear health") }
			if DefaultDamageWorld4E0B30(target, source, nil, 11, object.DamageExplosion, r) || target.HealthData.Cur != 200 {
				t.Fatal("zero post-Shield explosion did not return false with unchanged HP")
			}
			want := []string{"fire", "fire-sound", "damage-sound", "shield"}
			if sourcePresent {
				want = []string{"fire", "fire-sound", "buff", "damage-sound", "shield"}
			}
			if !slices.Equal(events, want) {
				t.Fatalf("events=%v want=%v", events, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessExplosionMissingService(t *testing.T) {
	for _, service := range []string{"damage-clear", "buff-off", "monster-hit-sound", "fire-protection"} {
		t.Run(service, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, false)
			target.ObjSubClass = 0x202
			source := damageMeleeUnitFixture4E17B0(t, false)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 { return 0 }
			switch service {
			case "damage-clear":
				r.DamageClear = nil
			case "buff-off":
				r.BuffOff = nil
			case "monster-hit-sound":
				r.MonsterHasHitSound = nil
			case "fire-protection":
				r.FireProtection = nil
			}
			var unsupported string
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { unsupported = reason }
			DefaultDamageWorld4E0B30(target, source, nil, 45, object.DamageExplosion, r)
			if unsupported == "" || target.HealthData.Cur != 200 || target.Obj130 != nil || target.Frame134 != 0 || target.UpdateDataMonster().Field547 != 0 {
				t.Fatal("missing required service silently reached the damage tail")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessExplosionNoShock(t *testing.T) {
	target := damageMeleeUnitFixture4E17B0(t, false)
	target.ObjSubClass = 0x202
	target.Buffs = 1 << defaultDamageShockEnchant4E0B30
	source := damageMeleeUnitFixture4E17B0(t, true)
	r := damageMeleeWorldRuntime4E0B30(t)
	r.IsEnemy = func(*Object, *Object) bool {
		t.Fatal("weapon-less arena hit must not enter the melee friendly gate")
		return false
	}
	r.FireProtection = func(*Object) float64 { return 0 }
	r.CallDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
		t.Fatal("weapon-less explosion must not trigger Shock retaliation")
		return true
	}
	DefaultDamageWorld4E0B30(target, source, nil, 45, object.DamageExplosion, r)
	if target.HealthData.Cur != 155 || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
		t.Fatal("weapon-less explosion lost its damage or consumed Shock")
	}
}

func TestDefaultDamageWorld4E0B30WeaponlessExplosionKeepsPlayerPrefixSeparate(t *testing.T) {
	for _, sourceKind := range []string{"none", "player", "npc"} {
		t.Run(sourceKind, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			update := target.UpdateDataPlayer()
			update.Field76, update.Field75, update.Field21, update.Field57 = 99, 77, 0x3e800000, 0x3f000000
			var source *Object
			if sourceKind != "none" {
				source = damageMeleeUnitFixture4E17B0(t, sourceKind == "player")
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 { return 0 }
			var unsupported string
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { unsupported = reason }
			DefaultDamageWorld4E0B30(target, source, nil, 45, object.DamageExplosion, r)
			if unsupported != "" || target.HealthData.Cur != 155 || target.Obj130 != source ||
				update.Field76 != 99 || update.Field75 != 77 || update.Field21 != 0x3e800000 || update.Field57 != 0x3f000000 {
				t.Fatal("DefaultDamage re-ran PlayerDamage's dedicated armor/carry/marker prefix")
			}
		})
	}
}
