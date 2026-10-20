package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/opennox/libs/types"
)

// The fixture executes GAME.EXE's unchanged 0050BA00 source-coordinate
// prefix and 00419A70, with the gameplay CW established by 0043E2C1.
// It records both modes so a default-nearest oracle cannot conceal the bug.
func TestAIPathGridGameplay50BA00OriginalInstructions(t *testing.T) {
	data, err := os.ReadFile("testdata/ai_path_grid_gameplay_50ba00.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "d981c02efaef59852cc008ed6048c1cf44d40f7a5340a308f8fb75f1dde1bcf1" {
		t.Fatal("original coordinate transcript changed")
	}
	var fixture struct {
		Original string `json:"original_sha256"`
		Control  uint16 `json:"gameplay_control_word"`
		Inverse  uint32 `json:"inverse_bits"`
		Rows     []struct {
			Name     string    `json:"name"`
			Input    [2]uint32 `json:"input_bits"`
			Gameplay [2]int32  `json:"gameplay"`
			Nearest  [2]int32  `json:"nearest"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0x0e3f || fixture.Inverse != math.Float32bits(aiPathGridInverse50AC20) || len(fixture.Rows) != 788 {
		t.Fatal("original gameplay scope changed")
	}
	var paths serverAIPaths
	for _, row := range fixture.Rows {
		t.Run(row.Name, func(t *testing.T) {
			p := types.Ptf(math.Float32frombits(row.Input[0]), math.Float32frombits(row.Input[1]))
			x, y := paths.GridCell50B810(&p)
			if got := [2]int32{int32(x), int32(y)}; got != row.Gameplay {
				t.Errorf("grid(%v) = %v, original gameplay = %v (default-nearest = %v)", p, got, row.Gameplay, row.Nearest)
			}
			for i, value := range []float32{p.X, p.Y} {
				if got := aiPathTargetCell50AC20(value); got != uint16(row.Gameplay[i]) {
					t.Errorf("destination(%v) = %#x, original gameplay low word = %#x", value, got, uint16(row.Gameplay[i]))
				}
			}
		})
	}
}
