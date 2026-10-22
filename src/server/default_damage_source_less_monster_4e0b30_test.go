package server

import (
	"encoding/json"
	"fmt"
	"github.com/opennox/libs/object"
	"os"
	"testing"
	"unsafe"
)

func TestSourceLessMonsterDefaultMatchesGAMEEXE4E0B30(t *testing.T) {
	data, err := os.ReadFile("testdata/default_damage_source_less_monster_4e0b30.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type     object.DamageType    `json:"damage_type"`
			Subclass object.SubClass      `json:"subclass"`
			Forward  []uint32             `json:"forward_damage"`
			Marker   uint32               `json:"marker"`
			Kind     uint32               `json:"marker_type"`
			Status   object.MonsterStatus `json:"status_flags"`
			Electric []byte               `json:"electric_state_bytes"`
			Return   int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 120 {
		t.Fatal("unexpected original ordinary-monster fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("type-%d/subclass-%x", int32(row.Type), uint32(row.Subclass)), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			v.ObjSubClass = row.Subclass
			ud := v.UpdateDataMonster()
			ud.Field547, ud.Field546 = 99, 77
			ud.Field523_2, ud.Field523_3 = 0x11, 0xab
			mod := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			init := &ModifierInitData{}
			init.Modifiers[2] = mod
			v.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 { return 0 }
			r.ElectricProtection = func(*Object) float64 { return 0 }
			r.CanApplyLateDefend = func(*ModifierEff) bool { t.Fatal("ordinary monster entered NPC late defense"); return false }
			got := []uint32{}
			r.DamageClear = func(target *Object, d int32) {
				if target != v {
					t.Fatal("wrong HP boundary")
				}
				got = append(got, uint32(d))
			}
			result := DefaultDamageWorld4E0B30(v, nil, nil, 9, row.Type, r)
			if result != (row.Return != 0) || fmt.Sprint(got) != fmt.Sprint(row.Forward) || ud.Field547 != row.Marker || ud.Field546 != row.Kind || ud.StatusFlags != row.Status || len(row.Electric) != 2 || ud.Field523_2 != row.Electric[0] || ud.Field523_3 != row.Electric[1] {
				t.Fatalf("ordinary monster result=%t HP-boundary=%v/%v marker=%d/%d status=%x/%x", result, got, row.Forward, ud.Field547, ud.Field546, ud.StatusFlags, row.Status)
			}
		})
	}
}
