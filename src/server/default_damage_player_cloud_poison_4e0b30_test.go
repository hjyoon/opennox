package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/libs/types"
)

func TestDefaultDamagePlayerCloudPoison4E0B30SignedTail(t *testing.T) {
	for _, owner := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
		for _, raw := range []int32{-3, 0, 3, 10, 25} {
			for _, enemy := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/raw-%d/enemy-%t", owner, raw, enemy), func(t *testing.T) {
					v, a, w := worldFlameFixture4E17B0(t, owner, "player")
					w.ObjClass, w.TypeInd = 0x190008, 1221
					v.ObjSubClass = 0x200 // This is MONSTER poison immunity, not PLAYER immunity.
					v.Buffs |= 1 << 26
					updateBefore := *v.UpdateDataPlayer()
					var events []string
					position := w.PrevPos
					if a == nil {
						position = types.Pointf{}
					}
					r := DefaultDamageWorldRuntime4E0B30{
						Frame:         func() uint32 { events = append(events, "frame"); return 1400 },
						GameplayFlag1: func() bool { return true },
						IsEnemy:       func(*Object, *Object) bool { events = append(events, "enemy"); return enemy },
						BuffOff: func(got *Object, id EnchantID) {
							if got != v || id != 0 || v.Pos132 != position {
								t.Fatal("visibility/position order")
							}
							events = append(events, "buff")
						},
						MonsterHasHitSound: func(*Object) bool { events = append(events, "lookup"); return false },
						DefaultDamageSound: func(got, attack *Object) {
							if got != v || attack != w || v.Obj130 != w || v.Field131 != 5 || v.Frame134 != 1400 {
								t.Fatal("sound/attribution order")
							}
							events = append(events, "sound")
						},
						PlayerSetState: func(got *Object, state PlayerState) bool {
							if got != v || raw < 20 || state != PlayerState30 {
								t.Fatal("hurt state")
							}
							got.UpdateDataPlayer().State = state
							events = append(events, "hurt")
							return true
						},
						DamageClear: func(got *Object, damage int32) {
							if got != v || damage != raw {
								t.Fatal("raw poison HP")
							}
							v.HealthData.Cur = uint16(max(20-damage, 0))
							events = append(events, "hp")
						},
						FireProtection:     func(*Object) float64 { t.Fatal("poison used fire protection"); return 0 },
						ElectricProtection: func(*Object) float64 { t.Fatal("poison used electric protection"); return 0 },
						ShieldReduce:       func(*Object, *int32, object.DamageType, *Object) { t.Fatal("Shield reduced poison") },
						AdjustFieldGuide:   func(*Object, *Object, int32) int32 { t.Fatal("player used monster field guide"); return 0 },
						Unsupported:        func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
					}
					if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamagePoison, r) {
						t.Fatal("cloud default rejected")
					}
					want := []string{}
					if a != nil {
						want = append(want, "enemy", "buff")
					}
					want = append(want, "frame")
					if owner == "NPC" {
						want = append(want, "lookup")
					}
					want = append(want, "sound")
					if raw >= 20 {
						want = append(want, "hurt")
						updateBefore.State = PlayerState30
					}
					if owner == "NPC" || owner == "proxy-NPC" {
						want = append(want, "enemy")
						if enemy {
							want = append(want, "frame")
						}
					}
					want = append(want, "hp")
					if !slices.Equal(events, want) || *v.UpdateDataPlayer() != updateBefore || v.HealthData.Cur != uint16(max(20-raw, 0)) || !v.HasEnchant(26) || math.Float32frombits(v.UpdateDataPlayer().Field21) != 0.25 {
						t.Fatalf("events=%v want=%v HP=%d", events, want, v.HealthData.Cur)
					}
				})
			}
		}
	}
}
