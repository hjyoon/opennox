package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// These independently executed original results also apply to the ordinary
// collision dispatcher's unshared search. Its production arithmetic must use
// the original startup FPU mode, while the generic nearest-profile controls
// keep their existing explicit expectations unchanged.
func TestChakramRetargetNativeOriginalStartup4EB250(t *testing.T) {
	raw, err := os.ReadFile("testdata/chakram_retarget_native_original_4eb250_pc53_tozero.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "e03c702b1cbe7aaaec0c71746bc48376bcb8fdfb30ce50c8ffb8d579cbd332eb" {
		t.Fatal("frozen original result changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	type record struct {
		Class  uint32 `json:"cls"`
		Flags  uint32
		Pos    [2]uint32
		Speed  uint32
		Vel    [2]uint32
		Owner  int
		Update int
	}
	var data struct {
		Rows []struct {
			Input struct {
				Name    string
				Objects [6]record
				Updates [3]struct {
					Last  int
					State uint8
				}
				Enums []struct{ Candidates []int }
			}
			Output struct {
				Result int
				State  struct{ Objects [6]record }
			}
		}
	}
	if err := json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, row := range data.Rows {
		name := row.Input.Name
		if name != "driver/route/3" && name != "driver/numeric/2" && name != "driver/numeric/15" && !strings.HasPrefix(name, "driver/finite-rounding/") {
			continue
		}
		checked++
		t.Run(name, func(t *testing.T) {
			var objects [6]Object
			var updates [3]ChakramUpdateData
			at := func(i int) *Object {
				if i == -1 {
					return nil
				}
				return &objects[i]
			}
			for i, value := range row.Input.Objects {
				objects[i] = Object{ObjClass: object.Class(value.Class), ObjFlags: object.Flags(value.Flags), ObjOwner: at(value.Owner), UpdateData: unsafe.Pointer(&updates[value.Update]), PosVec: types.Ptf(math.Float32frombits(value.Pos[0]), math.Float32frombits(value.Pos[1])), VelVec: types.Ptf(math.Float32frombits(value.Vel[0]), math.Float32frombits(value.Vel[1])), SpeedCur: math.Float32frombits(value.Speed)}
			}
			for i, value := range row.Input.Updates {
				updates[i].LastHit, updates[i].ReturnState = at(value.Last), value.State
			}
			result := chakramRetargetNative4EB250(&objects[0], chakramRetargetNativeDeps4EB250{
				eachInRect: func(_ types.Rectf, callback func(*Object)) {
					for _, index := range row.Input.Enums[0].Candidates {
						callback(at(index))
					}
				},
				mapCheck: func(*Object, *Object) bool { return true },
			})
			if result != at(row.Output.Result) {
				t.Fatal("ordinary native target differs from original", result, row.Output.Result)
			}
			got := [2]uint32{math.Float32bits(objects[0].VelVec.X), math.Float32bits(objects[0].VelVec.Y)}
			if got != row.Output.State.Objects[0].Vel {
				t.Fatal("ordinary native velocity differs from original startup arithmetic", got, row.Output.State.Objects[0].Vel)
			}
		})
	}
	if checked != 99 {
		t.Fatal("explicit original startup regression scope", checked)
	}
}
