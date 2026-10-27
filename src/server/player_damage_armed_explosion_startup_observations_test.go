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
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func originalArmedExplosionStartupSeal4E17B0(t *testing.T, nearestControl string) (string, int) {
	t.Helper()
	return originalArmedStartupSeal4E17B0(t, "testdata/armed_explosion_startup_prefix_original.json", nearestControl, 7194)
}

func originalArmedStartupSeal4E17B0(t *testing.T, path, nearestControl string, expectedRows int) (string, int) {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		Faults    []json.RawMessage `json:"faults"`
		Rows      []struct {
			Mode  string `json:"mode"`
			CW    uint16 `json:"control_word"`
			Rows  int    `json:"rows"`
			Bytes int    `json:"bytes"`
			SHA   string `json:"sha256"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Faults) != 0 || len(fixture.Rows) != 2 {
		t.Fatal("incomplete original armed prefix observations")
	}
	var control, startup string
	var size int
	for _, row := range fixture.Rows {
		if row.Rows != expectedRows {
			t.Fatal("original armed prefix input count changed")
		}
		if row.Mode == "nearest" && row.CW == 0x37f {
			control = row.SHA
		} else if row.Mode == "startup" && row.CW == 0xe3f {
			startup, size = row.SHA, row.Bytes
		}
	}
	if control != nearestControl || startup == "" || size == 0 {
		t.Fatal("original nearest control or startup prefix seal did not match")
	}
	return startup, size
}

// Original x87 multiplication/spill supplies the wear bits. Item damage and
// final HP are declared boundaries here; native C HP is exercised separately.
func TestPlayerDamageArmedExplosionShieldStartupOriginal4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/armed_explosion_shield_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		Percent   float64           `json:"block_percent"`
		Faults    []json.RawMessage `json:"faults"`
		Rows      [][]uint64        `json:"startup_rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Faults) != 0 || fixture.Percent != .3 || len(fixture.Rows) != 12 {
		t.Fatal("incomplete original armed shield observations")
	}
	for _, row := range fixture.Rows {
		if len(row) < 24 || row[8] != 7 || row[5] != 4 {
			t.Fatal("not the original shield raw9 fixture")
		}
		t.Run(fmt.Sprintf("source-%d/target-%d/weapon-%d", row[0], row[1], row[2]), func(t *testing.T) {
			v, a := damageMeleeUnitFixture4E17B0(t, row[1] == 1), damageMeleeUnitFixture4E17B0(t, row[0] == 1)
			w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, InitData: unsafe.Pointer(&ModifierInitData{}), PrevPos: types.Ptf(17, 9)}
			if row[2] == 1 {
				w.TypeInd, w.ObjClass, w.ObjSubClass, w.UseData.Ptr = 777, object.ClassWand, 0x10000, unsafe.Pointer(&WandUseData{Flags: 2})
			} else if row[2] == 2 {
				w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWeapon, 0x4000
			}
			item := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{}}
			v.InvFirstItem = item
			if row[1] == 1 {
				u := v.UpdateDataPlayer()
				u.Player.ArmorEquip, u.State, u.Field57, u.Field21 = 0x1000000, PlayerState16, math.Float32bits(.5), math.Float32bits(.25)
			} else {
				u := v.UpdateDataMonster()
				u.ArmorEquipFlags, u.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
				u.Field518, u.Field1 = math.Float32bits(.5), math.Float32bits(.25)
			}
			if !playerDamageArmedExplosionShape4E17B0(a, w, object.DamageExplosion) {
				t.Fatal("fixture did not select armed EXPLOSION")
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.BlockDirection = func(*Object, types.Pointf) bool { return true }
			r.BlockDamagePercent = func() float64 { return fixture.Percent }
			r.Audio = func(sound int, target *Object) {
				if sound != 878 || target != v {
					t.Fatal("original shield sound")
				}
			}
			r.CanDamageBlockItem = func(got *Object) bool { return got == item }
			calls := 0
			r.DamageBlockItem = func(got, owner, source, weapon *Object, amount float32, typ object.DamageType) bool {
				if got != item || owner != v || source != a || weapon != w || typ != object.DamageExplosion || math.Float32bits(amount) != uint32(row[20]) {
					t.Fatalf("original shield wear bits=%08x want=%08x", math.Float32bits(amount), row[20])
				}
				calls++
				return true
			}
			hp := v.HealthData.Cur
			handled, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			if !handled || result || calls != 1 || v.HealthData.Cur != hp || marker != uint32(row[10]) || kind != uint32(row[11]) || math.Float32bits(carry) != uint32(row[12]) {
				t.Fatal("original shield prefix or blocked-HP boundary changed")
			}
		})
	}
}
