package opennox

import (
	"encoding/json"
	"github.com/opennox/libs/object"
	"math"
	"os"
	"testing"
)

// The exact original input supplies damage, carry and absorbed armor damage.
func nativeSelfStrikeOriginal4E17B0(t *testing.T, sourcePlayer, player bool, typ object.DamageType, raw int32, armor, carry float32) (int32, uint32, int32) {
	t.Helper()
	blob, err := os.ReadFile("server/testdata/self_strike_compatibility_original.json")
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
	if data.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || data.CW != 0xe3f || len(data.Rows) != 764 {
		t.Fatal("incomplete original self-strike observations")
	}
	key := [8]uint32{0, 0, uint32(typ), uint32(raw), math.Float32bits(armor), math.Float32bits(carry), 0, math.Float32bits(1)}
	if sourcePlayer {
		key[0] = 1
	}
	if player {
		key[1] = 1
	}
	for _, row := range data.Rows {
		if len(row) == 14 && [8]uint32(row[:8]) == key {
			return int32(row[10]), row[11], int32(row[9])
		}
	}
	t.Fatalf("missing original self-strike input sourcePlayer=%t player=%t type=%d raw=%d armor=%g carry=%g", sourcePlayer, player, typ, raw, armor, carry)
	return 0, 0, 0
}
