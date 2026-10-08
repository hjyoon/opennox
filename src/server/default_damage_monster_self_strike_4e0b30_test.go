package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
)

func TestDefaultDamageMonsterSelfStrike4E0B30(t *testing.T) {
	for _, kind := range []string{"player", "monster"} {
		for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageImpale, object.DamageDrain, object.DamageClaw} {
			for _, enemy := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/type-%d/enemy-%t", kind, typ, enemy), func(t *testing.T) {
					v, a := monsterImpactTarget4E17B0(t, kind), monsterImpactTarget4E17B0(t, "monster")
					r := damageMeleeWorldRuntime4E0B30(t)
					r.IsEnemy = func(target, source *Object) bool {
						if target != v || source != a {
							t.Fatal("lost self-weapon enemy identity")
						}
						return enemy
					}
					if !DefaultDamageWorld4E0B30(v, a, a, 10, typ, r) {
						t.Fatal("native self-strike tail rejected")
					}
					want := uint16(200)
					if enemy {
						want = 190
					}
					if v.HealthData.Cur != want || (enemy && (v.Obj130 != a || v.Field131 != uint32(typ) || a.UpdateDataMonster().Field130 != 1400)) {
						t.Fatalf("HP=%d want=%d", v.HealthData.Cur, want)
					}
				})
			}
		}
	}
}
