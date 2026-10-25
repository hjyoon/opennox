package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestShapeBoxOriginal428270Native(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit box geometry context")
	}
	const path = "testdata/shape_box_original.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Group    string   `json:"group"`
			Name     string   `json:"name"`
			W        uint32   `json:"w_bits32"`
			H        uint32   `json:"h_bits32"`
			Expected []uint32 `json:"expected_output_bits32"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 1246 || len(fixture.Cases) != 1246 {
		t.Fatal("incomplete original contexts")
	}
	if unsafe.Sizeof(server.Shape{}) != 52 || unsafe.Offsetof(server.Shape{}.Box) != 12 || unsafe.Offsetof(server.ShapeBox{}.W) != 0 || unsafe.Offsetof(server.ShapeBox{}.H) != 4 || unsafe.Offsetof(server.ShapeBox{}.LeftTop) != 8 || unsafe.Offsetof(server.ShapeBox{}.RightTop2) != 36 {
		t.Fatal("native shape layout changed")
	}
	for _, group := range []string{"stock", "finite_edges"} {
		for _, kind := range []string{"Go", "C"} {
			t.Run(group+"_"+kind, func(t *testing.T) {
				count := 0
				for _, c := range fixture.Cases {
					if c.Group != group {
						continue
					}
					count++
					t.Run(c.Name, func(t *testing.T) {
						if len(c.Expected) != 8 {
							t.Fatal("incomplete eight original outputs")
						}
						frame, free := alloc.Make([]byte(nil), 180)
						defer free()
						if uintptr(unsafe.Pointer(&frame[64])) <= 0xffffffff {
							t.Fatal("native fixture below4GiB")
						}
						for i := range frame {
							frame[i] = 0xa5
						}
						binary.LittleEndian.PutUint32(frame[64:], 3)
						binary.LittleEndian.PutUint32(frame[64+12:], c.W)
						binary.LittleEndian.PutUint32(frame[64+16:], c.H)
						before := append([]byte(nil), frame...)
						shape := (*server.Shape)(unsafe.Pointer(&frame[64]))
						if kind == "Go" {
							shape.Box.Calc()
						} else {
							shapeBoxCalcCallNative428270(shape)
						}
						for i, b := range frame {
							if (i < 84 || i >= 116) && b != before[i] {
								t.Fatalf("native frame canary/nonoutput byte changed at%d", i)
							}
						}
						got := make([]uint32, 8)
						for i := range got {
							got[i] = binary.LittleEndian.Uint32(frame[84+4*i:])
						}
						equal := true
						for i := range got {
							if got[i] != c.Expected[i] {
								equal = false
							}
						}
						if !equal {
							t.Fatalf("actual native shape box=%08x want original=%08x", got, c.Expected)
						}
					})
				}
				want := 1182
				if group == "finite_edges" {
					want = 64
				}
				if count != want {
					t.Fatal("incomplete selected group", count, want)
				}
			})
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, after) || sha256.Sum256(raw) != sha256.Sum256(after) {
		t.Fatal("original fixture changed", err)
	}
}
