package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type pointShapeRow547DB0 struct {
	Input struct {
		Name            string
		Kind            uint32
		Position, Point [2]uint32
		Radius2         uint32
		Box             [6]uint32
		NilUnit         bool `json:"nil_unit"`
		NilPoint        bool `json:"nil_point"`
	}
	Output struct {
		Result *uint32
		Fault  []json.RawMessage
	}
}

func pointShapeOriginalRows547DB0(t *testing.T) []pointShapeRow547DB0 {
	t.Helper()
	path := os.Getenv("OPENNOX_POINT_SHAPE_ORIGINAL_GOLDEN")
	if path == "" {
		path = "../server/testdata/point_shape_original_547db0.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "c2a156668feeb44f178eecf9db25c44496cf45a0a1e22c79ec4068dcb3c87cb8" {
		t.Fatal("frozen original gzip changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "df1f713432dbdd2c97c5e635ae1d6e3b8ce94b7d9aa97037e4b34c4867963059" {
		t.Fatal("frozen original decoded bytes changed")
	}
	var data struct{ Rows []pointShapeRow547DB0 }
	if err := json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 987 {
		t.Fatal("original context count", len(data.Rows))
	}
	return data.Rows
}

func TestPointShapeOriginal547DB0(t *testing.T) {
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			if path == "C" && pointShapeOriginalABI547DB0() != 64 {
				t.Fatal("native public C entry not implemented; exact legacy fault is qualified separately")
			}
			for _, row := range pointShapeOriginalRows547DB0(t) {
				t.Run(row.Input.Name, func(t *testing.T) {
					w := newDodgeWorld547C50(t, dodgeRow547C50{})
					w.Unit.PosVec = types.Ptf(-73, 129) // Guard unrelated current position; original reads NewPos.
					w.Unit.NewPos = types.Ptf(math.Float32frombits(row.Input.Position[0]), math.Float32frombits(row.Input.Position[1]))
					w.Unit.Shape.Kind = server.ShapeKind(row.Input.Kind)
					w.Unit.Shape.Circle.R2 = math.Float32frombits(row.Input.Radius2)
					fields := []*float32{&w.Unit.Shape.Box.LeftTop, &w.Unit.Shape.Box.LeftBottom, &w.Unit.Shape.Box.LeftBottom2, &w.Unit.Shape.Box.LeftTop2, &w.Unit.Shape.Box.RightTop, &w.Unit.Shape.Box.RightBottom}
					for i, target := range fields {
						*target = math.Float32frombits(row.Input.Box[i])
					}
					point := (*types.Pointf)(unsafe.Pointer(&w.Heap[24000]))
					*point = types.Ptf(math.Float32frombits(row.Input.Point[0]), math.Float32frombits(row.Input.Point[1]))
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(point)) <= math.MaxUint32 {
						t.Fatal("native point must exceed4GiB")
					}
					unit := w.Unit
					if row.Input.NilUnit {
						unit = nil
					}
					if row.Input.NilPoint {
						point = nil
					}
					before := bytes.Clone(w.Heap)
					var fault any
					result := 0
					func() {
						defer func() { fault = recover() }()
						if path == "C" {
							result = pointShapeOriginalC547DB0(unit, point)
						} else if unit.Sub547DB0(point) {
							result = 1
						}
					}()
					if (fault != nil) != (len(row.Output.Fault) != 0) {
						t.Fatal("original raw fault boundary", fault, row.Output.Fault)
					}
					if !bytes.Equal(before, w.Heap) {
						t.Fatal("complete native object/MUD/point allocation or guards modified")
					}
					if fault == nil && (row.Output.Result == nil || uint32(result) != *row.Output.Result) {
						t.Fatalf("original point/shape result: got%d want%v", result, row.Output.Result)
					}
				})
			}
		})
	}
}
