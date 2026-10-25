package server

import (
	"encoding/json"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

type waterBarrelOriginalEvent struct {
	Service string   `json:"service"`
	Rect    []uint32 `json:"rect_bits32"`
	Index   int      `json:"object_index"`
	Sound   int      `json:"sound"`
}

func TestWaterBarrelPrecisionOriginal53CB90Native(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("64-bit full native source/candidate object frames")
	}
	var f struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Name       string   `json:"name"`
			Group      string   `json:"group"`
			Pos        []uint32 `json:"source_bits32"`
			Age        uint32   `json:"age"`
			Frame      uint32   `json:"frame"`
			Candidates []struct {
				Pos    []uint32 `json:"pos_bits32"`
				Radius uint32   `json:"radius_bits32"`
				Class  uint32   `json:"class_bits"`
			} `json:"candidates"`
			Events []waterBarrelOriginalEvent `json:"expected_services"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile("testdata/water_barrel_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 383 || len(f.Cases) != 383 || unsafe.Sizeof(Object{}) != 928 || unsafe.Sizeof(Shape{}) != 52 || unsafe.Sizeof(types.Pointf{}) != 8 {
		t.Fatal("incomplete original inputs or changed full native layout")
	}
	for _, group := range []string{"stock-derived", "finite-edges", "age-controls", "class-controls", "empty-control"} {
		t.Run(group, func(t *testing.T) {
			for _, c := range f.Cases {
				if c.Group != group {
					continue
				}
				t.Run(c.Name, func(t *testing.T) {
					if len(c.Pos) != 2 {
						t.Fatal("incomplete source position words")
					}
					frames := make([][]byte, len(c.Candidates)+1)
					before := make([][]byte, len(frames))
					objects := make([]*Object, len(frames))
					for i := range frames {
						frame, free := alloc.Make([]byte(nil), 928+128)
						defer free()
						for k := range frame {
							frame[k] = 0xa5
						}
						frames[i] = frame
						objects[i] = (*Object)(unsafe.Pointer(&frame[64]))
						if uintptr(unsafe.Pointer(objects[i])) <= 0xffffffff {
							t.Fatal("native frame not above4GiB")
						}
					}
					source := objects[0]
					source.PosVec = types.Ptf(math.Float32frombits(c.Pos[0]), math.Float32frombits(c.Pos[1]))
					source.Field32 = c.Frame - c.Age
					for i, input := range c.Candidates {
						if len(input.Pos) != 2 {
							t.Fatal("incomplete candidate position words")
						}
						o := objects[i+1]
						o.ObjClass = object.Class(input.Class)
						o.PosVec = types.Ptf(math.Float32frombits(input.Pos[0]), math.Float32frombits(input.Pos[1]))
						o.Shape.Circle.R = math.Float32frombits(input.Radius)
					}
					for i := range frames {
						before[i] = append([]byte(nil), frames[i]...)
					}
					label := func(o *Object) int {
						for i, p := range objects {
							if p == o {
								return i - 1
							}
						}
						t.Fatal("unknown native object pointer")
						return 0
					}
					events := []waterBarrelOriginalEvent{}
					waterBarrelUpdateNative53CB90(source, waterBarrelUpdateNativeDeps53CB90{
						frame: func() uint32 { return c.Frame },
						eachInRect: func(rect types.Rectf, callback func(*Object) bool) {
							events = append(events, waterBarrelOriginalEvent{Service: "each_rect", Rect: []uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}})
							for _, o := range objects[1:] {
								if !callback(o) {
									t.Fatal("declared enumeration stopped before next original candidate")
								}
							}
						},
						delayedDelete: func(o *Object) { events = append(events, waterBarrelOriginalEvent{Service: "delete", Index: label(o)}) },
						audio: func(id sound.ID, o *Object) {
							events = append(events, waterBarrelOriginalEvent{Service: "audio", Sound: int(id), Index: label(o)})
						},
					})
					for i := range frames {
						if !reflect.DeepEqual(frames[i], before[i]) {
							t.Fatalf("native full source/candidate/shape/nonoutput/canary frame%d changed", i)
						}
					}
					if !reflect.DeepEqual(events, c.Events) {
						t.Fatalf("actual native WaterBarrel events=%+v original=%+v", events, c.Events)
					}
				})
			}
		})
	}
}
