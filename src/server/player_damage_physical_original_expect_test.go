package server

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/opennox/libs/object"
)

type physicalOriginalExpected4E17B0 struct {
	Damage int32
	Carry  uint32
	Armor  int32
}

// Expectations come from unchanged original instructions, including old test
// input combinations. No candidate arithmetic supplies these expected values.
func physicalOriginal4E17B0(t *testing.T, typ object.DamageType, player bool, raw int32, armor, carry float32, quest bool, scale float32) physicalOriginalExpected4E17B0 {
	t.Helper()
	blob, err := os.ReadFile("testdata/physical_tail_compatibility_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		SHA  string     `json:"original_sha256"`
		CW   uint16     `json:"original_startup_control_word"`
		Rows [][]uint32 `json:"rows"`
	}
	if err := json.Unmarshal(blob, &data); err != nil {
		t.Fatal(err)
	}
	if data.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || data.CW != 0xe3f {
		t.Fatal("not original startup observations")
	}
	key := [7]uint32{uint32(typ), 0, uint32(raw), math.Float32bits(armor), math.Float32bits(carry), 0, math.Float32bits(scale)}
	if player {
		key[1] = 1
	}
	if quest {
		key[5] = 1
	}
	for _, row := range data.Rows {
		if len(row) != 10 {
			t.Fatal("incomplete original physical row")
		}
		if [7]uint32(row[:7]) == key {
			return physicalOriginalExpected4E17B0{Damage: int32(row[7]), Carry: row[8], Armor: int32(row[9])}
		}
	}
	t.Fatalf("missing original physical input type=%d player=%t raw=%d armor=%g carry=%g quest=%t scale=%g", typ, player, raw, armor, carry, quest, scale)
	return physicalOriginalExpected4E17B0{}
}
