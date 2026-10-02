package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageNative4E17B0SimpleCrushEntry(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, toPlayer := range []bool{false, true} {
			for i, name := range []string{"SmallFist", "MediumFist", "LargeFist"} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/%s", fromPlayer, toPlayer, name), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, toPlayer)
					source := damageMeleeUnitFixture4E17B0(t, fromPlayer)
					fist := &Object{TypeInd: uint16(777 + i), ObjClass: object.ClassSimple, ObjOwner: source}
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					r := damageMeleeRuntimeFixture4E17B0(t)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					reason := ""
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, fist, 9, object.DamageCrush, r)
					if !h || !result || target.HealthData.Cur != 193 || armor.HealthData.Cur != 23 || reason != "" {
						t.Fatalf("Fist entry: handled=%t/%t HP=200->%d armor=%d reason=%q", h, result, target.HealthData.Cur, armor.HealthData.Cur, reason)
					}
					if target.Obj130 != fist || target.Field131 != uint32(object.DamageCrush) || target.Frame134 != 1400 {
						t.Fatal("Fist entry lost default-tail attribution")
					}
				})
			}
		}
	}
}
