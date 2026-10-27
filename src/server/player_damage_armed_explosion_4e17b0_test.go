package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageArmedExplosion4E17B0RepeatedHP(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, toPlayer := range []bool{false, true} {
			for _, kind := range []string{"Sword", "Wand", "Hammer"} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/%s", fromPlayer, toPlayer, kind), func(t *testing.T) {
					v, a := damageMeleeUnitFixture4E17B0(t, toPlayer), damageMeleeUnitFixture4E17B0(t, fromPlayer)
					w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, ObjFlags: object.Flags(17301776), InitData: unsafe.Pointer(&ModifierInitData{})}
					if kind == "Wand" {
						w.TypeInd, w.ObjClass, w.ObjSubClass, w.UseData.Ptr = 777, object.ClassWand, 0x10000, unsafe.Pointer(&WandUseData{Flags: 2})
					}
					if kind == "Hammer" {
						w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWeapon, 0x4000
					}
					if toPlayer {
						v.UpdateDataPlayer().Field57 = math.Float32bits(.5)
					} else {
						v.UpdateDataMonster().Field518 = math.Float32bits(.5)
					}
					r := damageMeleeRuntimeFixture4E17B0(t)
					world := damageMeleeWorldRuntime4E0B30(t)
					world.FireProtection = func(*Object) float64 { return 0 }
					r.DefaultDamage = func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
						return DefaultDamageWorld4E0B30(target, source, weapon, damage, typ, world)
					}
					for hit, want := range []uint16{199, 197, 196} {
						if h, result := PlayerDamageNative4E17B0(v, a, w, 3, object.DamageExplosion, r); !h || !result {
							t.Fatalf("hit %d handled/result=%t/%t", hit, h, result)
						}
						mark, kind, carry := damageMeleeMarker4E17B0(v)
						if v.HealthData.Cur != want || mark != 1 || kind != uint32(w.TypeInd) || v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != 7 || carry != []float32{.5, 0, .5}[hit] {
							t.Fatalf("hit %d HP=%d/%d marker=%d/%d carry=%g type=%d", hit, v.HealthData.Cur, want, mark, kind, carry, v.Field131)
						}
					}
				})
			}
		}
	}
}
