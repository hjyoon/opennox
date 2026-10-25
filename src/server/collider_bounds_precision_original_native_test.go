package server

import (
	"encoding/json"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

func TestColliderBoundsPrecisionOriginalNative(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("64-bit native full collider object frames")
	}
	var fixture struct {
		Contexts int      `json:"contexts"`
		Initial  []uint32 `json:"initial_output_bits32"`
		Cases    []struct {
			Name    string   `json:"name"`
			Group   string   `json:"group"`
			Shape   []uint32 `json:"shape_bits32"`
			Pos     []uint32 `json:"position_bits32"`
			Current []uint32 `json:"expected_current_bits32"`
			Next    []uint32 `json:"expected_next_bits32"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile("testdata/collider_bounds_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 2910 || len(fixture.Cases) != 2910 || len(fixture.Initial) != 4 || unsafe.Sizeof(Object{}) != 928 || unsafe.Sizeof(Shape{}) != 52 || unsafe.Sizeof(types.Pointf{}) != 8 {
		t.Fatal("incomplete original inputs or changed native layout")
	}
	for _, which := range []string{"PosVec", "NewPos"} {
		t.Run(which, func(t *testing.T) {
			for _, c := range fixture.Cases {
				t.Run(c.Name, func(t *testing.T) {
					if len(c.Shape) != 13 || len(c.Pos) != 4 || len(c.Current) != 4 || len(c.Next) != 4 {
						t.Fatal("incomplete input words")
					}
					frame, free := alloc.Make([]byte(nil), 928+128)
					defer free()
					for i := range frame {
						frame[i] = 0xa5
					}
					obj := (*Object)(unsafe.Pointer(&frame[64]))
					if uintptr(unsafe.Pointer(obj)) <= 0xffffffff {
						t.Fatal("native object not above4GiB")
					}
					shape := (*[13]uint32)(unsafe.Pointer(&obj.Shape))
					copy(shape[:], c.Shape)
					obj.PosVec = types.Ptf(math.Float32frombits(c.Pos[0]), math.Float32frombits(c.Pos[1]))
					obj.NewPos = types.Ptf(math.Float32frombits(c.Pos[2]), math.Float32frombits(c.Pos[3]))
					obj.CollideP1 = types.Ptf(math.Float32frombits(fixture.Initial[0]), math.Float32frombits(fixture.Initial[1]))
					obj.CollideP2 = types.Ptf(math.Float32frombits(fixture.Initial[2]), math.Float32frombits(fixture.Initial[3]))
					before := append([]byte(nil), frame...)
					want := c.Current
					var result *Object
					if which == "PosVec" {
						result = obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
					} else {
						result = obj.Sub_4E7350()
						want = c.Next
					}
					if result != obj {
						t.Fatal("native returned object identity changed")
					}
					got := []uint32{math.Float32bits(obj.CollideP1.X), math.Float32bits(obj.CollideP1.Y), math.Float32bits(obj.CollideP2.X), math.Float32bits(obj.CollideP2.Y)}
					offsets := []int{int(unsafe.Offsetof(Object{}.CollideP1)), int(unsafe.Offsetof(Object{}.CollideP1)) + 4, int(unsafe.Offsetof(Object{}.CollideP2)), int(unsafe.Offsetof(Object{}.CollideP2)) + 4}
					allowed := make(map[int]bool)
					for _, off := range offsets {
						for i := 0; i < 4; i++ {
							allowed[64+off+i] = true
						}
					}
					for i := range frame {
						if !allowed[i] && frame[i] != before[i] {
							t.Fatalf("native full object/shape/Z/canary byte%d changed", i)
						}
					}
					if !reflect.DeepEqual(shape[:], c.Shape) {
						t.Fatal("original shape/canary words changed")
					}
					if c.Shape[0] != 1 && c.Shape[0] != 2 && c.Shape[0] != 3 {
						if !reflect.DeepEqual(frame, before) {
							t.Fatal("unknown kind changed original frame")
						}
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("actual native collider bounds=%08x original=%08x", got, want)
					}
				})
			}
		})
	}
}
