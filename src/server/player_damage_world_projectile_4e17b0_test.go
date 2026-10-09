package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageNative4E17B0WorldProjectiles(t *testing.T) {
	for _, tc := range worldProjectileCases4E0B30() {
		for _, player := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/player-%t", tc.name, player), func(t *testing.T) {
				v := damageMeleeUnitFixture4E17B0(t, player)
				a, w, m := worldProjectileFixture4E0B30(tc)
				armor := damageMeleeArmorFixture4E17B0(v, 0.5, 0.4)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				world := damageMeleeWorldRuntime4E0B30(t)
				world.FireProtection = func(*Object) float64 { return 0 }
				reason := ""
				world.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				r.Unsupported = world.Unsupported
				r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
					return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
				}
				h, ok := PlayerDamageNative4E17B0(v, a, w, 9, tc.typ, r)
				wantHP, wantArmor := uint16(195), uint16(21)
				if tc.typ == object.DamageCrush {
					wantHP, wantArmor = 193, 23
				}
				if tc.typ == object.DamageFlame {
					wantHP, wantArmor = 191, 16
				}
				if !h || !ok || v.HealthData.Cur != wantHP || armor.HealthData.Cur != wantArmor || reason != "" {
					t.Fatalf("world entry %t/%t HP=%d/%d armor=%d/%d unsupported=%q", h, ok, v.HealthData.Cur, wantHP, armor.HealthData.Cur, wantArmor, reason)
				}
				marker, kind, carry := damageMeleeMarker4E17B0(v)
				if marker != 1 || kind != uint32(m.TypeInd) || v.Obj130 != m {
					t.Fatalf("world marker=%d/%d", marker, kind)
				}
				wantCarry := float32(-0.1)
				if tc.typ == object.DamageCrush {
					wantCarry = float32(0.15)
				}
				if tc.typ == object.DamageFlame {
					wantCarry = 0.4
				}
				if math.Abs(float64(carry-wantCarry)) > 0.000001 {
					t.Fatalf("carry=%g want=%g", carry, wantCarry)
				}
			})
		}
	}
}
