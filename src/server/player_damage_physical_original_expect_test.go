package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

// npcMissileCarryOriginal4E17B0 reads the existing startup-CW original numeric
// observations for the common full-armor case-11/4E20F0 path. The source-less
// oracle does not supply Harpoon or Stone caller/attribution equivalence.
// Its non-Quest scale=-1 is a recorded, unused service input; it is not queried
// by either live non-Quest missile test.
func npcMissileCarryOriginal4E17B0(t *testing.T, carry float32) physicalOriginalExpected4E17B0 {
	t.Helper()
	blob, err := os.ReadFile("testdata/source_less_physical_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(blob)) != "03bc22764197f9b9d805eba4050c2ec1a04c0defa6d7a36127f7e66560dbbc01" {
		t.Fatal("changed independently recorded physical numeric observations")
	}
	var data struct {
		SHA   string `json:"original_sha256"`
		Cases []struct {
			sourceLessPhysicalOriginalInput
			Wear uint32 `json:"expected_wear_bits32"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(blob, &data); err != nil {
		t.Fatal(err)
	}
	if data.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(data.Cases) != 6622 {
		t.Fatal("incomplete original physical numeric contexts")
	}
	var selected physicalOriginalExpected4E17B0
	matches := 0
	for _, row := range data.Cases {
		if !row.Player && row.Type == object.DamageImpact && row.Raw == 3 && row.Armor == math.Float32bits(.25) && row.Carry == math.Float32bits(carry) && !row.Quest && row.Scale == math.Float32bits(-1) && !row.God {
			if len(row.ExpectedDamage) != 1 {
				t.Fatal("original numeric row must have one actual forwarded amount")
			}
			selected = physicalOriginalExpected4E17B0{Damage: int32(row.ExpectedDamage[0]), Carry: row.ExpectedCarry, Armor: int32(row.Wear)}
			matches++
		}
	}
	if matches != 1 {
		t.Fatalf("missing or ambiguous original case11 input carry=%08x rows=%d", math.Float32bits(carry), matches)
	}
	return selected
}
