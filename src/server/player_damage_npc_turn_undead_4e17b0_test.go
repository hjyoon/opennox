package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageNPCTurnUndeadMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_npc_turn_undead_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Owner   string `json:"owner"`
			Front   bool   `json:"front"`
			Raw     int32  `json:"raw"`
			Forward uint32 `json:"forward_damage"`
			Marker  uint32 `json:"marker"`
			Kind    uint32 `json:"marker_type"`
			Carry   uint32 `json:"carry_bits"`
			Return  int    `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 80 {
		t.Fatal("unexpected original NPC UndeadKiller fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("owner-%s/front-%t/raw-%d", row.Owner, row.Front, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			v.ObjSubClass = 0x11052
			ud := v.UpdateDataMonster()
			ud.Field547, ud.Field546 = 99, 77
			ud.Field518, ud.Field1 = math.Float32bits(.5), math.Float32bits(.25)
			w := &Object{ObjClass: object.Class(0x40080000), TypeInd: 777, PrevPos: types.Ptf(44, 7)}
			var a *Object
			switch row.Owner {
			case "player":
				a = damageMeleeUnitFixture4E17B0(t, true)
			case "monster":
				a = damageMeleeUnitFixture4E17B0(t, false)
			case "none":
			case "self":
				a = w
			case "world":
				a = &Object{}
			default:
				t.Fatal("unknown owner")
			}
			// Raw type6 must never admit durability or inspect this item.
			v.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.CoopMode = func() bool { return false }
			r.QuestMode = func() bool { return false }
			r.GodMode = func() bool { return true }
			r.CanDamageArmor = func(*Object) bool { t.Fatal("type6 admitted armor wear"); return false }
			r.ItemArmorValue = func(*Object) float32 { t.Fatal("type6 read armor value"); return 0 }
			exclusions, directions, calls := 0, 0, 0
			r.BlockSourceExcluded = func(it *Object) bool {
				if a == nil || it != w || ud.Field547 != 0 {
					t.Fatal("exclusion arguments or cached marker order")
				}
				exclusions++
				return false
			}
			r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("collider used weapon-less exclusions"); return false }
			r.BlockDirection = func(it *Object, pos types.Pointf) bool {
				if a == nil || it != v || pos != w.PrevPos {
					t.Fatal("direction arguments/source-less gate")
				}
				directions++
				return row.Front
			}
			forward := int32(123456)
			r.DefaultDamage = func(it, source, weapon *Object, d int32, k object.DamageType) bool {
				if it != v || source != a || weapon != w || k != object.DamageDispelUndead {
					t.Fatal("type6 DefaultDamage arguments")
				}
				calls++
				forward = d
				return true
			}
			h, ok := PlayerDamageNative4E17B0(v, a, w, row.Raw, object.DamageDispelUndead, r)
			wantDefense := 1
			if a == nil {
				wantDefense = 0
			}
			if !h || ok != (row.Return != 0) || calls != 1 || uint32(forward) != row.Forward || ud.Field547 != row.Marker || ud.Field546 != row.Kind || ud.Field1 != row.Carry || exclusions != wantDefense || directions != wantDefense {
				t.Fatalf("NPC type6=%t/%t forward=%d/%d marker=%d/%d carry=%x/%x defenses=%d/%d", h, ok, forward, int32(row.Forward), ud.Field547, ud.Field546, ud.Field1, row.Carry, exclusions, directions)
			}
		})
	}
}
