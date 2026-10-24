package legacy

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestFloatToIntOriginal419A70Native(t *testing.T) {
	raw, err := os.ReadFile("../server/testdata/player_camper_reset_4d7c60.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Unchanged bool   `json:"original_unchanged"`
		Startup   uint32 `json:"actual_startup_control_word"`
		Rows      []struct {
			Name   string            `json:"name"`
			CW     uint32            `json:"fpcw"`
			Input  map[string]uint32 `json:"input_bits"`
			Result uint32            `json:"start"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil || !fixture.Unchanged || len(fixture.Rows) != 90 || fixture.Startup&0xf3f != 0xe3f {
		t.Fatalf("original converter fixture: %v", err)
	}
	convert := summonRuntime500DA0().FloatToInt // Actual linked production C.nox_float2int binding.
	selected := 0
	for _, row := range fixture.Rows {
		if row.CW != fixture.Startup {
			continue
		}
		selected++
		t.Run(row.Name, func(t *testing.T) {
			bits := row.Input["CamperStartTime"]
			if got := convert(math.Float32frombits(bits)); got != int32(row.Result) {
				t.Fatalf("actual linked C helper bits=%08x got=%08x original=%08x", bits, uint32(got), row.Result)
			}
		})
	}
	if selected != 30 {
		t.Fatalf("selected%d original startup converter observations", selected)
	}
}
