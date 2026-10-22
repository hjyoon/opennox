package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// Script Damage/DoDamage supplies the initialized monster as both source
// and weapon. Signed nonpositive rows exercise the direct callback ABI;
// public DoDamage rejects those amounts. DefaultDamage/wear are declared
// original boundaries here, and native HP is checked separately.
func TestPlayerDamageMonsterSelfRawFireMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_monster_self_raw_fire_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Player  bool              `json:"player"`
			Type    object.DamageType `json:"damage_type"`
			Raw     int32             `json:"raw"`
			Forward uint32            `json:"forward_damage"`
			Marker  uint32            `json:"marker"`
			Kind    uint32            `json:"marker_type"`
			Carry   uint32            `json:"carry_bits"`
			Wear    []uint32          `json:"wear"`
			Return  int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 144 {
		t.Fatal("unexpected original initialized-monster source fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("player-%t/type-%d/raw-%d", row.Player, row.Type, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			attacker := damageMeleeUnitFixture4E17B0(t, false)
			if row.Player {
				v.UpdateDataPlayer().Field76, v.UpdateDataPlayer().Field75 = 99, 77
			} else {
				v.UpdateDataMonster().Field547, v.UpdateDataMonster().Field546 = 99, 77
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			r.CoopMode = func() bool { return false }
			r.IsEnemy = func(*Object, *Object) bool { return true }
			r.QuestMode = func() bool { return false }
			r.GodMode = func() bool { return false }
			r.GameplayFlag1 = func() bool { return true }
			r.FireProtection = func(*Object) float64 { return 0 }
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			r.BuffOff = func(*Object, EnchantID) {}
			r.PlayerSetState = func(*Object, PlayerState) bool { return true }
			calls, forward, wear := 0, int32(123456), int32(0)
			r.DefaultDamage = func(target, a, w *Object, d int32, k object.DamageType) bool {
				if target != v || a != attacker || w != attacker || k != row.Type {
					t.Fatal("monster-source DefaultDamage arguments")
				}
				calls++
				forward = d
				return true
			}
			r.DamageClear = func(target *Object, d int32) {
				if target != v {
					t.Fatal("wrong HP target")
				}
				calls++
				forward = d
			}
			r.DamageArmor = func(item, a, w *Object, d int32, k object.DamageType) bool {
				if item != armor || a != attacker || w != attacker || k != row.Type {
					t.Fatal("monster-source armor arguments")
				}
				wear += d
				return true
			}
			reason := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			h, ok := PlayerDamageNative4E17B0(v, attacker, attacker, row.Raw, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			wantWear := int32(0)
			if len(row.Wear) != 0 {
				wantWear = max(int32(row.Wear[0]), 0)
			}
			if !h || ok != (row.Return != 0) || calls != 1 || uint32(forward) != row.Forward || marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || wear != wantWear || reason != "" {
				t.Fatalf("handled/result=%t/%t forward=%d/%d calls=%d marker=%d/%d carry=%g/%g wear=%d/%d reason=%s", h, ok, forward, int32(row.Forward), calls, marker, kind, carry, math.Float32frombits(row.Carry), wear, wantWear, reason)
			}
		})
	}
}

func TestPlayerDamageMonsterSelfRawFireFrontReflectMatchesGAMEEXE4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/player_damage_monster_self_raw_fire_front_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type    object.DamageType `json:"damage_type"`
			Player  bool              `json:"player"`
			Mode    string            `json:"owner"`
			Forward []uint32          `json:"forward_damage"`
			Marker  uint32            `json:"marker"`
			Kind    uint32            `json:"marker_type"`
			Carry   uint32            `json:"carry_bits"`
			Events  [][]any           `json:"events"`
			Return  int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 72 {
		t.Fatal("unexpected original raw front/Reflect fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("player-%t/type-%d/%s", row.Player, row.Type, row.Mode), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			a := damageMeleeUnitFixture4E17B0(t, false)
			a.PosVec, a.PrevPos = types.Ptf(20, 0), types.Ptf(-20, 7)
			if row.Player {
				u := v.UpdateDataPlayer()
				u.Field76, u.Field75, u.Field21, u.Field57 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
			} else {
				u := v.UpdateDataMonster()
				u.Field547, u.Field546, u.Field1, u.Field518 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
			}
			reflect, front := strings.Contains(row.Mode, "reflect"), strings.HasPrefix(row.Mode, "front")
			if reflect {
				v.Buffs |= 1 << playerDamageReflectEnchant4E17B0
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.CoopMode = func() bool { return false }
			r.QuestMode = func() bool { return false }
			var trace, wantTrace []string
			for _, e := range row.Events {
				switch e[0] {
				case "facing":
					wantTrace = append(wantTrace, "facing")
				case "point-fx":
					wantTrace = append(wantTrace, fmt.Sprintf("fx:%d", int(e[1].(float64))))
				case "audio":
					wantTrace = append(wantTrace, fmt.Sprintf("audio:%d", int(e[1].(float64))))
				}
			}
			directions, exclusions := 0, 0
			r.BlockDirection = func(target *Object, pos types.Pointf) bool {
				want := a.PrevPos
				if reflect && row.Type == object.DamageZapRay && directions == 0 {
					want = a.PosVec
				}
				if target != v || pos != want {
					t.Fatal("raw Reflect/ordinary direction snapshot")
				}
				directions++
				trace = append(trace, "facing")
				return front
			}
			r.BlockSourceExcluded = func(source *Object) bool {
				if source != a {
					t.Fatal("raw weapon exclusion source")
				}
				exclusions++
				return false
			}
			r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("nonnull actor weapon used source-only exclusions"); return false }
			r.PointFX = func(id int, pos types.Pointf) {
				if pos != v.PosVec {
					t.Fatal("raw Reflect FX position")
				}
				trace = append(trace, fmt.Sprintf("fx:%d", id))
			}
			r.Audio = func(id int, target *Object) {
				if target != v {
					t.Fatal("raw Reflect audio target")
				}
				trace = append(trace, fmt.Sprintf("audio:%d", id))
			}
			var forward []uint32
			r.DefaultDamage = func(target, source, weapon *Object, d int32, k object.DamageType) bool {
				if target != v || source != a || weapon != a || k != row.Type {
					t.Fatal("raw forward identity")
				}
				forward = append(forward, uint32(d))
				return true
			}
			h, ok := PlayerDamageNative4E17B0(v, a, a, 9, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			wantExclusions := 1
			if reflect && front && row.Type == object.DamageZapRay {
				wantExclusions = 0
			}
			if !h || ok != (row.Return != 0) || fmt.Sprint(forward) != fmt.Sprint(row.Forward) || marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || fmt.Sprint(trace) != fmt.Sprint(wantTrace) || exclusions != wantExclusions {
				t.Fatalf("raw front/Reflect=%t/%t forward=%v/%v marker=%d/%d trace=%v/%v exclusions=%d/%d", h, ok, forward, row.Forward, marker, kind, trace, wantTrace, exclusions, wantExclusions)
			}
		})
	}
}
