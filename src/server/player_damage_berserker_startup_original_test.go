package server

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type berserkerStartupOriginalRow4E17B0 struct {
	Input struct {
		Name       string            `json:"name"`
		Player     bool              `json:"player"`
		Raw        int32             `json:"raw"`
		Type       object.DamageType `json:"damage_type"`
		Armor      uint32            `json:"absorption_bits32"`
		Carry      uint32            `json:"carry_bits32"`
		Quest      bool              `json:"quest"`
		QuestScale uint32            `json:"quest_scale_bits32"`
	} `json:"input"`
	Events []struct {
		Service string `json:"service"`
		Damage  uint32 `json:"damage_bits32"`
	} `json:"events"`
	Carry  uint32 `json:"carry_bits32"`
	Marker uint32 `json:"marker"`
	Kind   uint32 `json:"marker_type"`
}

func berserkerStartupOriginalRows4E17B0(t *testing.T) []berserkerStartupOriginalRow4E17B0 {
	return physicalStartupOriginalRows4E17B0(t, "testdata/berserker_crush_startup_original.json")
}

func physicalStartupOriginalRows4E17B0(t *testing.T, path string) []berserkerStartupOriginalRow4E17B0 {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		SHA       string                              `json:"original_sha256"`
		Unchanged bool                                `json:"original_unchanged"`
		CW        uint16                              `json:"original_startup_control_word"`
		Faults    []json.RawMessage                   `json:"memory_faults"`
		Rows      []berserkerStartupOriginalRow4E17B0 `json:"rows"`
	}
	if err := json.Unmarshal(blob, &data); err != nil {
		t.Fatal(err)
	}
	if data.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !data.Unchanged || data.CW != 0xe3f || len(data.Faults) != 0 || len(data.Rows) != 60 {
		t.Fatal("incomplete original Berserker startup observations")
	}
	return data.Rows
}

func TestPlayerDamageBerserkerStartupOriginal4E17B0(t *testing.T) {
	testPhysicalStartupOriginalRows4E17B0(t, berserkerStartupOriginalRows4E17B0(t))
}

func TestPlayerDamageTrollImpactStartupOriginal4E17B0(t *testing.T) {
	testPhysicalStartupOriginalRows4E17B0(t, physicalStartupOriginalRows4E17B0(t, "testdata/troll_impact_startup_original.json"))
}

func testPhysicalStartupOriginalRows4E17B0(t *testing.T, rows []berserkerStartupOriginalRow4E17B0) {
	for _, row := range rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			makePlayer := func() *Object {
				u := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 713, ObjClass: object.ClassPlayer})
				info := selfWeaponElectricAllocOriginal(t, Player{PlayerUnit: u})
				ud := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: info, State: PlayerState13})
				u.UpdateData = unsafe.Pointer(ud)
				return u
			}
			source := makePlayer()
			if in.Type == object.DamageImpact {
				source.TypeInd, source.ObjClass, source.ObjSubClass = 1343, object.ClassMonster, 0x11012
				source.UpdateData = unsafe.Pointer(selfWeaponElectricAllocOriginal(t, MonsterUpdateData{}))
			}
			source.PrevPos = types.Ptf(44, 19)
			target := makePlayer()
			var carry, marker, kind *uint32
			if in.Player {
				u := target.UpdateDataPlayer()
				u.Field21, u.Field57, u.Field76, u.Field75 = in.Carry, in.Armor, 99, 77
				carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
			} else {
				u := selfWeaponElectricAllocOriginal(t, MonsterUpdateData{Field1: in.Carry, Field518: in.Armor, Field547: 99, Field546: 77})
				target.ObjClass, target.ObjSubClass, target.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(u)
				carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
			}
			target.HealthData = selfWeaponElectricAllocOriginal(t, HealthData{Cur: 200, Max: 200})
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(source.CObj()) <= math.MaxUint32) {
				t.Fatal("expected C-owned pointers above 4 GiB")
			}
			if len(row.Events) != 2 || row.Events[0].Service != "armor" || row.Events[1].Service != "default" {
				t.Fatal("original boundaries missing")
			}
			calls := 0
			check := func(v, a, w *Object, amount int32, typ object.DamageType) bool {
				if v != target || a != source || w != source || typ != in.Type {
					t.Fatal("original charge identity changed")
				}
				if uint32(amount) != row.Events[1].Damage {
					t.Errorf("charge damage=%d want=%d", amount, int32(row.Events[1].Damage))
				}
				calls++
				return true
			}
			r := PlayerDamageRuntime4E17B0{
				Frame: func() uint32 { return 1400 }, CoopMode: func() bool { return false },
				GameplayFlag1: func() bool { return true }, GodMode: func() bool { return false },
				QuestMode: func() bool { return in.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(in.QuestScale) },
				IsEnemy:             func(*Object, *Object) bool { return true },
				BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
				BlockDirection: func(*Object, types.Pointf) bool { return false },
				BuffOff:        func(*Object, EnchantID) {}, Audio: func(int, *Object) {}, PlayerDamageSound: func(*Object, *Object) {},
				PlayerSetState: func(v *Object, state PlayerState) bool { v.UpdateDataPlayer().State = state; return true },
				DefaultDamage:  check,
				DamageClear:    func(v *Object, amount int32) { check(v, source, source, amount, in.Type) },
				Unsupported:    func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
			}
			h, result := PlayerDamageNative4E17B0(target, source, source, in.Raw, in.Type, r)
			if !h || !result || calls != 1 || *carry != row.Carry || *marker != row.Marker || *kind != row.Kind || target.HealthData.Cur != 200 {
				t.Fatalf("charge h/result=%t/%t calls=%d carry=%08x/%08x marker=%d/%d want=%d/%d", h, result, calls, *carry, row.Carry, *marker, *kind, row.Marker, row.Kind)
			}
		})
	}
}
