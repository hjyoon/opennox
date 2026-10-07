package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// C-owned records retain the actual class-zero script caster and separate
// terminal-owner/caster identities. The expected arithmetic is independent of
// the port's rounding helper: GAME.EXE spills absorption and carry to f32 and
// rounds to nearest even before the positive-only minimum and late Quest scale.
func TestPlayerDamageNPCCasterImpact4E17B0ArmorAndOwners(t *testing.T) {
	for _, owner := range []string{"player-self", "player-owned-NPC", "imaginary-self", "imaginary-owned-NPC", "monster-owner", "proxy-monster", "world-self", "nil"} {
		for _, absorption := range []float32{0, 0.25, 1} {
			for _, raw := range []int32{-3, 0, 1, 5, 8, 33} {
				for _, mode := range []string{"normal", "quest", "god"} {
					t.Run(fmt.Sprintf("%s/armor-%g/raw-%d/%s", owner, absorption, raw, mode), func(t *testing.T) {
						v, a, w := defaultDamageCasterImpactFixture4E0B30(t, owner, "NPC")
						ud := v.UpdateDataMonster()
						ud.Field518, ud.Field1 = math.Float32bits(absorption), math.Float32bits(0.125)
						ud.Field547, ud.Field546 = 99, 77
						v.Buffs |= 1 << playerDamageReflectEnchant4E17B0
						var events []string
						r := damageMeleeRuntimeFixture4E17B0(t)
						r.Frame = func() uint32 { t.Fatal("uninvulnerable NPC IMPACT read frame before DefaultDamage"); return 1400 }
						r.CoopMode = func() bool {
							if ud.Field547 != 99 {
								t.Fatal("NPC cached marker cleared before Coop mode")
							}
							events = append(events, "coop")
							return mode != "normal"
						}
						r.BlockSourceExcluded = func(attack *Object) bool {
							if attack != w || ud.Field547 != 0 || ud.Field546 != 77 {
								t.Fatal("caster marker clear/exclusion identity")
							}
							events = append(events, "exclude")
							return false
						}
						r.BlockDirection = func(target *Object, pos types.Pointf) bool {
							marker, markerType := uint32(0), uint32(77)
							if a != w {
								marker, markerType = 1, uint32(w.TypeInd)
							}
							if target != v || pos != w.PrevPos || ud.Field547 != marker || ud.Field546 != markerType {
								t.Fatal("caster PrevPos/attribution before ordinary facing")
							}
							events = append(events, "facing")
							return false
						}
						accumulated := float32((1-float64(absorption))*float64(raw)) + float32(0.125)
						rounded := int32(math.RoundToEven(float64(accumulated)))
						effective := rounded
						if raw > 0 && effective == 0 {
							effective = 1
						}
						if mode == "quest" {
							before := effective
							effective = int32(math.RoundToEven(float64(float32(float64(effective) * 0.5))))
							if before > 0 && effective < 1 {
								effective = 1
							}
						}
						marker, markerType := uint32(2), uint32(11)
						if a != nil && a != w {
							marker, markerType = 1, uint32(w.TypeInd)
						}
						r.GodMode = func() bool {
							if ud.Field547 != marker || ud.Field546 != markerType || ud.Field1 != math.Float32bits(accumulated-float32(rounded)) {
								t.Fatal("NPC God query preceded live carry/cached hit marker")
							}
							events = append(events, "god")
							return mode == "god"
						}
						r.QuestMode = func() bool { events = append(events, "quest"); return mode == "quest" }
						r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
						r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
							if target != v || source != a || weapon != w || amount != effective || typ != object.DamageImpact {
								t.Fatalf("NPC caster tail amount=%d want=%d", amount, effective)
							}
							events = append(events, "default")
							return true
						}
						r.Audio = func(int, *Object) { t.Fatal("unblocked caster IMPACT used block/Reflect audio") }
						if h, result := PlayerDamageNative4E17B0(v, a, w, raw, object.DamageImpact, r); !h || !result {
							t.Fatalf("NPC caster IMPACT handled/result=%t/%t", h, result)
						}
						want := []string{"coop"}
						if a != nil {
							want = append(want, "exclude", "facing")
						}
						want = append(want, "god", "quest")
						if mode == "quest" {
							want = append(want, "scale")
						}
						want = append(want, "default")
						if !slices.Equal(events, want) || v.HealthData.Cur != 20 {
							t.Fatalf("NPC caster events=%v want=%v HP=%d", events, want, v.HealthData.Cur)
						}
					})
				}
			}
		}
	}
}
