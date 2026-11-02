package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type occupancyUnit54EF00 struct {
	Kind     uint32
	Position [2]uint32
	Radius2  uint32
	Box      [6]uint32
	NilUnit  bool   `json:"nil_unit"`
	Class    uint32 `json:"unit_class"`
	Flags    uint32
}
type occupancyStep54EF00 struct {
	Unit           int
	Point          *[2]uint32
	Class          *uint32 `json:"unit_class"`
	Flags, Scratch *uint32
}
type occupancyRow54EF00 struct {
	Input struct {
		Name, Entry, Alias string
		Point              [2]uint32
		NilPoint           bool `json:"nil_point"`
		Units              []occupancyUnit54EF00
		Scratch            uint32
		Steps              []occupancyStep54EF00
		PostScratch        *uint32 `json:"post_scratch"`
	}
	Output struct {
		Events  []json.RawMessage
		Scratch uint32
		Point   []uint32
		Result  *uint32
		Fault   []json.RawMessage
	}
}

func occupancyOriginalRows54EF00(t *testing.T) []occupancyRow54EF00 {
	t.Helper()
	path := os.Getenv("OPENNOX_POINT_OCCUPANCY_ORIGINAL_GOLDEN")
	if path == "" {
		path = "../server/testdata/point_occupancy_original_54ef00.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "75fed3afdf90b7083cdf42403fd6a7545f97b97b31867363ae038e9b3fef7384" {
		t.Fatal("frozen original closure gzip changed")
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
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "9d4ba9fc453e876795a34a8e1280ccf30109bf713ef121934b376aaf3992f093" {
		t.Fatal("frozen original closure decoded bytes changed")
	}
	var data struct{ Rows []occupancyRow54EF00 }
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 1224 {
		t.Fatal("original closure row count", len(data.Rows))
	}
	return data.Rows
}

type occupancyWorld54EF00 struct {
	Heap    []byte
	Units   []*server.Object
	Point   *types.Pointf
	Scratch *uint32
	Stage   string
	Events  []any
	Row     occupancyRow54EF00
}

func newOccupancyWorld54EF00(t *testing.T, row occupancyRow54EF00) *occupancyWorld54EF00 {
	t.Helper()
	base := newDodgeWorld547C50(t, dodgeRow547C50{})
	w := &occupancyWorld54EF00{Heap: base.Heap, Scratch: memmap.PtrUint32(0x5D4594, 2491708), Row: row, Events: make([]any, 0), Stage: "candidate"}
	neighbor := memmap.PtrUint32(0x5D4594, 2491712)
	old, oldNeighbor := *w.Scratch, *neighbor
	t.Cleanup(func() { *w.Scratch, *neighbor = old, oldNeighbor })
	*w.Scratch, *neighbor = row.Input.Scratch, 0xa5a5a5a5
	for i, input := range row.Input.Units {
		ptr := base.Unit
		if i > 0 {
			ptr = (*server.Object)(unsafe.Pointer(&w.Heap[4096+4096*i]))
		}
		ptr.ObjClass, ptr.Field5 = object.Class(input.Class), input.Flags
		ptr.NewPos = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
		ptr.Shape.Kind = server.ShapeKind(input.Kind)
		ptr.Shape.Circle.R2 = math.Float32frombits(input.Radius2)
		fields := []*float32{&ptr.Shape.Box.LeftTop, &ptr.Shape.Box.LeftBottom, &ptr.Shape.Box.LeftBottom2, &ptr.Shape.Box.LeftTop2, &ptr.Shape.Box.RightTop, &ptr.Shape.Box.RightBottom}
		for j, field := range fields {
			*field = math.Float32frombits(input.Box[j])
		}
		w.Units = append(w.Units, ptr)
	}
	switch row.Input.Alias {
	case "heap":
		w.Point = (*types.Pointf)(unsafe.Pointer(&w.Heap[24000]))
	case "unit-newpos":
		w.Point = &w.Units[0].NewPos
	case "scratch":
		w.Point = (*types.Pointf)(unsafe.Pointer(w.Scratch))
	default:
		t.Fatal("unknown original point role", row.Input.Alias)
	}
	if row.Input.NilPoint {
		w.Point = nil
	} else {
		*w.Point = types.Ptf(math.Float32frombits(row.Input.Point[0]), math.Float32frombits(row.Input.Point[1]))
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && w.Point != nil && uintptr(unsafe.Pointer(w.Point)) <= math.MaxUint32 {
		t.Fatal("native point must exceed4GiB")
	}
	return w
}
func (w *occupancyWorld54EF00) pointBits() []uint32 {
	if w.Point == nil {
		return nil
	}
	return []uint32{math.Float32bits(w.Point.X), math.Float32bits(w.Point.Y)}
}
func (w *occupancyWorld54EF00) state(kind string, index int) map[string]any {
	u := w.Units[index]
	return map[string]any{"kind": kind, "unit": index, "nil_unit": w.Row.Input.Units[index].NilUnit, "unit_class": uint32(u.ObjClass), "flags": u.Field5, "point": w.pointBits(), "scratch": *w.Scratch}
}
func (w *occupancyWorld54EF00) store(value uint32, source string) {
	*w.Scratch = value
	w.Events = append(w.Events, map[string]any{"kind": "store", "source": source, "value": value})
}
func (w *occupancyWorld54EF00) runtime(t *testing.T) pointOccupancyOriginalRuntime54EF00 {
	return pointOccupancyOriginalRuntime54EF00{
		Store: func(v uint32) { w.store(v, w.Stage) },
		Load: func() uint32 {
			w.Events = append(w.Events, map[string]any{"kind": "load", "value": *w.Scratch})
			return *w.Scratch
		},
		Each: func(rect types.Rectf, fn func(*server.Object) bool) {
			w.Events = append(w.Events, map[string]any{"kind": "iterator", "rect": [4]uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}, "point": w.pointBits(), "scratch": *w.Scratch, "same_context": true})
			w.Stage = "candidate"
			for _, step := range w.Row.Input.Steps {
				unit := w.Units[step.Unit]
				if step.Point != nil {
					*w.Point = types.Ptf(math.Float32frombits(step.Point[0]), math.Float32frombits(step.Point[1]))
					w.Events = append(w.Events, map[string]any{"kind": "mutate-point", "value": *step.Point})
				}
				if step.Class != nil {
					unit.ObjClass = object.Class(*step.Class)
					w.Events = append(w.Events, map[string]any{"kind": "mutate-unit_class", "unit": step.Unit, "value": *step.Class})
				}
				if step.Flags != nil {
					unit.Field5 = *step.Flags
					w.Events = append(w.Events, map[string]any{"kind": "mutate-flags", "unit": step.Unit, "value": *step.Flags})
				}
				if step.Scratch != nil {
					w.store(*step.Scratch, "service")
				}
				w.Events = append(w.Events, w.state("before", step.Unit))
				arg := unit
				if w.Row.Input.Units[step.Unit].NilUnit {
					arg = nil
				}
				if !fn(arg) {
					t.Fatal("original iterator must not stop after a hit")
				}
				w.Events = append(w.Events, w.state("after", step.Unit))
			}
			if w.Row.Input.PostScratch != nil {
				w.store(*w.Row.Input.PostScratch, "service")
			}
		},
	}
}
func (w *occupancyWorld54EF00) expectedHeap(t *testing.T, before []byte) []byte {
	t.Helper()
	out := bytes.Clone(before)
	base := uintptr(unsafe.Pointer(&w.Heap[0]))
	put := func(ptr unsafe.Pointer, v uint32) {
		at := uintptr(ptr)
		if at < base || at+4 > base+uintptr(len(out)) {
			return
		}
		binary.LittleEndian.PutUint32(out[at-base:], v)
	}
	for _, raw := range w.Row.Output.Events {
		var e struct {
			Kind  string
			Unit  int
			Value json.RawMessage
		}
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		switch e.Kind {
		case "mutate-point":
			var value [2]uint32
			if err := json.Unmarshal(e.Value, &value); err != nil {
				t.Fatal(err)
			}
			put(unsafe.Pointer(w.Point), value[0])
			put(unsafe.Add(unsafe.Pointer(w.Point), 4), value[1])
		case "mutate-unit_class", "mutate-flags":
			var value uint32
			if err := json.Unmarshal(e.Value, &value); err != nil {
				t.Fatal(err)
			}
			if e.Kind == "mutate-flags" {
				put(unsafe.Pointer(&w.Units[e.Unit].Field5), value)
			} else {
				put(unsafe.Pointer(&w.Units[e.Unit].ObjClass), value)
			}
		}
	}
	return out
}

func TestPointOccupancyOriginal54EF00(t *testing.T) {
	if occupancyOriginalParentABI54EF00() != 64 || occupancyOriginalCandidateABI54EF60() != 64 {
		t.Fatal("native parent/context callback not implemented; old fault retained separately")
	}
	for _, row := range occupancyOriginalRows54EF00(t) {
		t.Run(row.Input.Name, func(t *testing.T) {
			w := newOccupancyWorld54EF00(t, row)
			before := bytes.Clone(w.Heap)
			expected := w.expectedHeap(t, before)
			r := w.runtime(t)
			old := pointOccupancyRuntimeOverride54EF00
			pointOccupancyRuntimeOverride54EF00 = &r
			defer func() { pointOccupancyRuntimeOverride54EF00 = old }()
			var fault any
			var result uint32
			func() {
				defer func() { fault = recover() }()
				if row.Input.Entry == "54EF00" {
					w.Stage = "root"
					result = occupancyOriginalParentC54EF00(w.Point)
				} else {
					w.Events = append(w.Events, w.state("before", 0))
					unit := w.Units[0]
					if row.Input.Units[0].NilUnit {
						unit = nil
					}
					occupancyOriginalCandidateC54EF60(unit, w.Point)
					w.Events = append(w.Events, w.state("after", 0))
				}
			}()
			if (fault != nil) != (len(row.Output.Fault) != 0) {
				t.Fatal("original raw closure fault prefix", fault, row.Output.Fault)
			}
			if fault == nil && row.Input.Entry == "54EF00" && (row.Output.Result == nil || result != *row.Output.Result) {
				t.Fatal("original parent DWORD result", result, row.Output.Result)
			}
			if *w.Scratch != row.Output.Scratch || !reflect.DeepEqual(w.pointBits(), row.Output.Point) {
				t.Fatal("original final scratch/point", *w.Scratch, row.Output.Scratch, w.pointBits(), row.Output.Point)
			}
			if !reflect.DeepEqual(dodgeCanonical547C50(t, w.Events), dodgeCanonical547C50(t, row.Output.Events)) {
				t.Fatalf("original ordered stores/iteration/callback state differs\ngot%s\nwant%s", mustDodgeJSON547C50(w.Events), mustDodgeJSON547C50(row.Output.Events))
			}
			if !bytes.Equal(w.Heap, expected) {
				t.Fatal("whole native object/MUD/point allocation or guard changed beyond original declared input mutations")
			}
		})
	}
}
