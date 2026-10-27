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

func selfStrikeStartupOriginalRows4E17B0(t *testing.T, path string, count int) [][14]uint32 {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"original_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      [][]uint32        `json:"rows"`
	}
	if err := json.Unmarshal(blob, &data); err != nil {
		t.Fatal(err)
	}
	if data.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !data.Unchanged || data.CW != 0xe3f || len(data.Faults) != 0 || len(data.Rows) != count {
		t.Fatal("incomplete unchanged-original self-strike observations")
	}
	rows := make([][14]uint32, len(data.Rows))
	for i, row := range data.Rows {
		if len(row) != 14 {
			t.Fatal("invalid original self-strike observation width")
		}
		copy(rows[i][:], row)
	}
	return rows
}

func TestPlayerDamageSelfStrikeStartupOriginal4E17B0(t *testing.T) {
	rows := selfStrikeStartupOriginalRows4E17B0(t, "testdata/self_strike_startup_original.json", 660)
	for i, row := range rows {
		t.Run(fmt.Sprintf("source-player-%t/type-%d/player-%t/numeric-%d", row[0] != 0, row[2], row[1] != 0, i%30), func(t *testing.T) {
			makeUnit := func(player bool) *Object {
				u := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 1343, ObjClass: object.ClassMonster, ObjSubClass: 0x11012})
				if player {
					info := selfWeaponElectricAllocOriginal(t, Player{PlayerUnit: u})
					ud := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: info, State: PlayerState13})
					u.TypeInd, u.ObjClass, u.ObjSubClass, u.UpdateData = 713, object.ClassPlayer, 0, unsafe.Pointer(ud)
				} else {
					u.UpdateData = unsafe.Pointer(selfWeaponElectricAllocOriginal(t, MonsterUpdateData{}))
				}
				return u
			}
			source, target := makeUnit(row[0] != 0), makeUnit(row[1] != 0)
			source.PrevPos = types.Ptf(44, 19)
			target.HealthData = selfWeaponElectricAllocOriginal(t, HealthData{Cur: 200, Max: 200})
			var carry, marker, kind *uint32
			if row[1] != 0 {
				ud := target.UpdateDataPlayer()
				ud.Field21, ud.Field57, ud.Field76, ud.Field75 = row[5], row[4], 99, 77
				carry, marker, kind = &ud.Field21, &ud.Field76, &ud.Field75
			} else {
				ud := target.UpdateDataMonster()
				ud.Field1, ud.Field518, ud.Field547, ud.Field546 = row[5], row[4], 99, 77
				carry, marker, kind = &ud.Field1, &ud.Field547, &ud.Field546
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(source.CObj()) <= math.MaxUint32) {
				t.Fatal("expected native C-owned records above 4 GiB")
			}
			calls := 0
			check := func(v, a, w *Object, amount int32, typ object.DamageType) bool {
				if v != target || a != source || w != source || uint32(typ) != row[2] {
					t.Fatal("self-strike attribution changed")
				}
				if uint32(amount) != row[10] {
					t.Errorf("self-strike damage=%d want=%d", amount, int32(row[10]))
				}
				calls++
				return true
			}
			r := PlayerDamageRuntime4E17B0{
				Frame: func() uint32 { return 1400 }, CoopMode: func() bool { return false },
				GameplayFlag1: func() bool { return true }, GodMode: func() bool { return false },
				QuestMode: func() bool { return row[6] != 0 }, QuestDamageScale: func() float32 { return math.Float32frombits(row[7]) },
				IsEnemy:             func(*Object, *Object) bool { return true },
				BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
				BlockDirection: func(*Object, types.Pointf) bool { return false },
				BuffOff:        func(*Object, EnchantID) {}, Audio: func(int, *Object) {}, PlayerDamageSound: func(*Object, *Object) {},
				PlayerSetState: func(v *Object, state PlayerState) bool { v.UpdateDataPlayer().State = state; return true },
				DefaultDamage:  check,
				DamageClear:    func(v *Object, amount int32) { check(v, source, source, amount, object.DamageType(row[2])) },
				Unsupported:    func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
			}
			h, result := PlayerDamageNative4E17B0(target, source, source, int32(row[3]), object.DamageType(row[2]), r)
			if !h || !result || calls != 1 || *carry != row[11] || *marker != row[12] || *kind != row[13] || target.HealthData.Cur != 200 {
				t.Fatalf("self-strike h/result=%t/%t calls=%d carry=%08x/%08x marker=%d/%d want=%d/%d", h, result, calls, *carry, row[11], *marker, *kind, row[12], row[13])
			}
		})
	}
}
