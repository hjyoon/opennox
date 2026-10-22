package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// Script Damage/DoDamage supplies the initialized player as both source
// and weapon. Signed nonpositive rows exercise the direct callback ABI;
// public DoDamage rejects those amounts. DefaultDamage/wear are declared
// original boundaries here, and native HP is checked separately.
func TestPlayerDamagePlayerSelfSourceMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_player_self_source_4e17b0.json")
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
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 80 {
		t.Fatal("unexpected original initialized-player source fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("player-%t/type-%d/raw-%d", row.Player, row.Type, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			attacker := damageMeleeUnitFixture4E17B0(t, true)
			attacker.UpdateDataPlayer().Player.PlayerUnit = attacker
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
					t.Fatal("source-less DefaultDamage arguments")
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
					t.Fatal("source-less armor arguments")
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

func TestPlayerDamagePlayerSelfSourceRequiresInfo4E17B0(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageBlade, object.DamageImpale, object.DamageDrain, object.DamageBite, object.DamageClaw} {
			t.Run(fmt.Sprintf("player-%t/type-%d", player, typ), func(t *testing.T) {
				v := damageMeleeUnitFixture4E17B0(t, player)
				if player {
					v.UpdateDataPlayer().Field76, v.UpdateDataPlayer().Field75 = 99, 77
				} else {
					v.UpdateDataMonster().Field547, v.UpdateDataMonster().Field546 = 99, 77
				}
				a := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{})}
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.CoopMode = func() bool { return false }
				r.ObserveClear = func(*Object) { t.Fatal("invalid source entered observer prefix") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("invalid source entered defenses"); return false }
				r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
					t.Fatal("invalid source reached HP")
					return false
				}
				reason := ""
				r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
				h, ok := PlayerDamageNative4E17B0(v, a, a, 9, typ, r)
				marker, kind, carry := damageMeleeMarker4E17B0(v)
				want := "unsupported monster damage shape"
				if player {
					want = "unsupported player damage shape"
				}
				if h || ok || reason != want || v.HealthData.Cur != 200 || marker != 99 || kind != 77 || carry != 0 {
					t.Fatalf("missing player info=%t/%t reason=%q HP=%d marker=%d/%d carry=%g", h, ok, reason, v.HealthData.Cur, marker, kind, carry)
				}
			})
		}
	}
}
