package opennox

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/opennox/libs/object"
)

func rootPhysicalOriginal4E17B0(t *testing.T, typ object.DamageType, raw int32, armor, carry float32) (int32, uint32) {
	t.Helper()
	blob, err := os.ReadFile("server/testdata/physical_tail_compatibility_original.json")
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
	key := [7]uint32{uint32(typ), 1, uint32(raw), math.Float32bits(armor), math.Float32bits(carry), 0, math.Float32bits(1)}
	for _, row := range data.Rows {
		if len(row) == 10 && [7]uint32(row[:7]) == key {
			return int32(row[7]), row[8]
		}
	}
	t.Fatalf("missing original physical input type=%d raw=%d armor=%g carry=%g", typ, raw, armor, carry)
	return 0, 0
}
