package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
)

// Exercise the production entry, not only its newly restored melee helper.
// The previous entry silently refused these hits on 64-bit hosts.
func TestPlayerDamageMeleeEntry4E17B0BidirectionalHP(t *testing.T) {
	for _, toPlayer := range []bool{false, true} {
		for _, attack := range []struct {
			name     string
			typ      object.DamageType
			subclass object.SubClass
		}{
			{"Sword", object.DamageBlade, object.SubClass(object.WeaponSword)},
			{"MorningStar", object.DamageCrush, object.SubClass(object.WeaponMace)},
			{"Unarmed", object.DamageClaw, 0},
		} {
			t.Run(fmt.Sprintf("player-target-%t/%s", toPlayer, attack.name), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, toPlayer)
				source := damageMeleeUnitFixture4E17B0(t, !toPlayer)
				var weapon *Object
				if attack.subclass != 0 {
					weapon = &Object{TypeInd: 444, ObjClass: object.ClassWeapon, ObjSubClass: attack.subclass}
				}
				r := damageMeleeRuntimeFixture4E17B0(t)
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				handled, result := PlayerDamageNative4E17B0(target, source, weapon, 10, attack.typ, r)
				if !handled || !result || target.HealthData.Cur != 190 || reason != "" {
					t.Fatalf("ordinary melee entry: handled=%t result=%t HP=200->%d reason=%q", handled, result, target.HealthData.Cur, reason)
				}
				t.Logf("production melee entry: target=%p source=%p weapon=%p HP=200->190", target, source, weapon)
			})
		}
	}
}

func TestPlayerDamageMeleeEntry4E17B0FriendlyAndArmored(t *testing.T) {
	for _, toPlayer := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-target-%t", toPlayer), func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, toPlayer)
			source := damageMeleeUnitFixture4E17B0(t, !toPlayer)
			weapon := &Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(object.WeaponSword)}
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0)
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			world := damageMeleeWorldRuntime4E0B30(t)
			enemy := false
			world.IsEnemy = func(*Object, *Object) bool { return enemy }
			r.DefaultDamage = func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
				return DefaultDamageWorld4E0B30(target, source, weapon, damage, typ, world)
			}
			if handled, result := PlayerDamageNative4E17B0(target, source, weapon, 10, object.DamageBlade, r); !handled || !result || target.HealthData.Cur != 200 {
				t.Fatal("friendly Sword must not reduce unit HP")
			}
			enemy = true
			if handled, result := PlayerDamageNative4E17B0(target, source, weapon, 10, object.DamageBlade, r); !handled || !result || target.HealthData.Cur != 195 {
				t.Fatalf("armored hostile Sword HP=%d, want 195", target.HealthData.Cur)
			}
		})
	}
}
