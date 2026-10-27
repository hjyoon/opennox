package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The independent x87 reference replays all 26 original FPU bytes, including
// both mass loads, with an adapted x86-64 integer/stack scaffold under Rosetta.
// The complete PE32 root is independently exercised by the finite fixture.
func TestBoxWallImpulse548600SpecialValuesOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/impulse_x87_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA      string `json:"original_sha256"`
		Contexts int    `json:"contexts"`
		Bytes    int    `json:"verified_original_FPU_bytes"`
		Whole    bool   `json:"whole_original_root_native_execution_claim"`
		Rows     []struct {
			Input struct {
				Mass        uint32    `json:"mass_bits32"`
				Impulse     [2]uint32 `json:"impulse_bits32"`
				Accumulator [2]uint32 `json:"initial_accumulator_bits32"`
			} `json:"input"`
			Accumulator [2]uint32 `json:"accumulator_bits32"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Contexts != 256 || len(fixture.Rows) != 256 || fixture.Bytes != 26 || fixture.Whole {
		t.Fatal("incomplete independent x87 reference")
	}
	for index, row := range fixture.Rows {
		t.Run(fmt.Sprintf("%03d", index), func(t *testing.T) {
			obj, free := alloc.New(Object{})
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
				t.Fatal("native impulse object below 4 GiB")
			}
			*obj = Object{Mass: math.Float32frombits(row.Input.Mass), Pos24: types.Ptf(math.Float32frombits(row.Input.Accumulator[0]), math.Float32frombits(row.Input.Accumulator[1])), PosVec: types.Ptf(10, 20), VelVec: types.Ptf(3, 4), Field29: 0x13572468}
			raw := func() []byte { return unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj))) }
			before := bytes.Clone(raw())
			obj.Sub548600(types.Ptf(math.Float32frombits(row.Input.Impulse[0]), math.Float32frombits(row.Input.Impulse[1])))
			got := [2]uint32{math.Float32bits(obj.Pos24.X), math.Float32bits(obj.Pos24.Y)}
			if got != row.Accumulator {
				t.Fatalf("native accumulator=%08x original x87=%08x", got, row.Accumulator)
			}
			obj.Pos24 = types.Ptf(math.Float32frombits(row.Input.Accumulator[0]), math.Float32frombits(row.Input.Accumulator[1]))
			if !bytes.Equal(raw(), before) {
				t.Fatal("impulse changed nonoutput bytes")
			}
		})
	}
}
