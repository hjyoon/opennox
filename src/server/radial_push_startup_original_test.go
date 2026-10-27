package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type radialInputOriginal struct {
	Kind           string      `json:"kind"`
	Name           string      `json:"name"`
	Origin         [2]uint32   `json:"origin"`
	OriginMode     int         `json:"origin_mode"`
	Outer          uint32      `json:"outer"`
	Inner          uint32      `json:"inner"`
	Force          uint32      `json:"force"`
	Objects        [][7]uint32 `json:"objects"`
	Order          []int       `json:"order"`
	Callback       bool        `json:"callback"`
	TraceMask      uint32      `json:"trace_mask"`
	MutationEvent  string      `json:"mutation_event"`
	MutationAction string      `json:"mutation_action"`
}
type radialStateOriginal struct {
	Objects [][7]uint32 `json:"objects"`
	Origin  [2]uint32   `json:"origin"`
	Context []uint32    `json:"context"`
}
type radialEventOriginal struct {
	Service string   `json:"service"`
	Args    []uint32 `json:"args"`
	radialStateOriginal
}
type radialRowOriginal struct {
	Mode      string                `json:"mode"`
	Input     radialInputOriginal   `json:"input"`
	Events    []radialEventOriginal `json:"events"`
	Mutated   bool                  `json:"mutated"`
	Preserved bool                  `json:"nonoutput_bytes_preserved"`
	radialStateOriginal
}
type radialOpsOriginal struct {
	origin               func() *types.Pointf
	radius, inner, force func() float32
	each                 func(types.Rectf, func(*Object) bool)
	movable              func(*Object) bool
	position             func(*Object) types.Pointf
	mass                 func(*Object) float32
	missile              func(*Object) bool
	loadForce            func(*Object) types.Pointf
	storeForce           func(*Object, types.Pointf, types.Pointf)
	trace                func(types.Pointf, types.Pointf) bool
	before               func(*Object, float32)
	wake                 func(*Object)
	first                *Object
}

func radialRowsOriginal(t *testing.T) []radialRowOriginal {
	t.Helper()
	blob, err := os.ReadFile("testdata/radial_push_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string              `json:"original_sha256"`
		Unchanged  bool                `json:"original_unchanged"`
		CW         uint16              `json:"actual_startup_control_word"`
		Contexts   int                 `json:"contexts"`
		Executions int                 `json:"root_executions"`
		Faults     []json.RawMessage   `json:"memory_faults"`
		Rows       []radialRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !f.Unchanged || f.CW != 0xe3f || f.Contexts != 454 || f.Executions != 908 || len(f.Rows) != 908 || len(f.Faults) != 0 {
		t.Fatal("incomplete original radial push fixture")
	}
	return f.Rows
}
func TestRadialPushStartupOriginal(t *testing.T) {
	count := 0
	for _, row := range radialRowsOriginal(t) {
		if row.Mode != "actual-startup" || row.Input.Kind != "around" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			radialFixtureOriginal(t, row, func(o radialOpsOriginal) {
				if !mapPushUnitsAround52E040(o.origin(), o.radius(), o.inner(), o.force(), radialHooksOriginal(o)) {
					t.Fatal("radial root rejected native dependencies")
				}
			})
		})
	}
	if count != 350 {
		t.Fatalf("around conditions=%d", count)
	}
}
func radialFixtureOriginal(t *testing.T, row radialRowOriginal, run func(radialOpsOriginal)) {
	t.Helper()
	i := row.Input
	var objs [5]*Object
	var frames [5][]byte
	var expected [5][]byte
	frame := func(size uintptr) []byte {
		p, free := alloc.Make([]byte(nil), int(size)+64)
		t.Cleanup(free)
		for n := range p {
			p[n] = 0xa5
		}
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&p[32])) <= math.MaxUint32 {
			t.Fatal("native C frame below 4 GiB")
		}
		return p
	}
	for n := range objs {
		p := frame(unsafe.Sizeof(Object{}))
		frames[n] = p
		obj := (*Object)(unsafe.Pointer(&p[32]))
		objs[n] = obj
		d := i.Objects[n]
		*obj = Object{PosVec: rigidPointOriginal([2]uint32{d[0], d[1]}), ForceVec: rigidPointOriginal([2]uint32{d[2], d[3]}), Mass: math.Float32frombits(d[4]), ObjClass: object.Class(d[5]), ObjFlags: object.Flags(d[6]), Field37: 0x55667788, Field38: 0x11223344}
	}
	litFrame, altFrame := frame(unsafe.Sizeof(types.Pointf{})), frame(unsafe.Sizeof(types.Pointf{}))
	literal, alternate := (*types.Pointf)(unsafe.Pointer(&litFrame[32])), (*types.Pointf)(unsafe.Pointer(&altFrame[32]))
	*literal = types.Pointf{}
	*alternate = types.Ptf(299.9, 299.8)
	refs := []*types.Pointf{literal, &objs[0].PosVec, &objs[0].ForceVec, &objs[1].PosVec, alternate}
	origin := refs[i.OriginMode]
	*origin = rigidPointOriginal(i.Origin)
	radius, inner, scaled := math.Float32frombits(i.Outer), math.Float32frombits(i.Inner), math.Float32frombits(i.Force)
	enabled := i.Callback
	for n, p := range frames {
		expected[n] = bytes.Clone(p)
		obj := objs[n]
		d := row.Objects[n]
		for _, v := range []struct {
			offset uintptr
			value  uint32
		}{{unsafe.Offsetof(obj.PosVec), d[0]}, {unsafe.Offsetof(obj.PosVec) + 4, d[1]}, {unsafe.Offsetof(obj.ForceVec), d[2]}, {unsafe.Offsetof(obj.ForceVec) + 4, d[3]}, {unsafe.Offsetof(obj.Mass), d[4]}, {unsafe.Offsetof(obj.ObjClass), d[5]}, {unsafe.Offsetof(obj.ObjFlags), d[6]}} {
			binary.LittleEndian.PutUint32(expected[n][32+v.offset:], v.value)
		}
	}
	state := func() radialStateOriginal {
		s := radialStateOriginal{Objects: make([][7]uint32, 5), Origin: [2]uint32{math.Float32bits(origin.X), math.Float32bits(origin.Y)}, Context: make([]uint32, 0)}
		for n, p := range objs {
			s.Objects[n] = [7]uint32{math.Float32bits(p.PosVec.X), math.Float32bits(p.PosVec.Y), math.Float32bits(p.ForceVec.X), math.Float32bits(p.ForceVec.Y), math.Float32bits(p.Mass), uint32(p.ObjClass), uint32(p.ObjFlags)}
		}
		if i.Kind == "candidate" {
			index := -1
			for n, p := range refs {
				if p == origin {
					index = n
				}
			}
			if index < 0 {
				t.Fatal("foreign origin reference")
			}
			flag := uint32(0)
			if enabled {
				flag = 1
			}
			s.Context = []uint32{math.Float32bits(radius), math.Float32bits(inner), math.Float32bits(scaled), flag, uint32(index)}
		}
		return s
	}
	events := make([]radialEventOriginal, 0)
	visit, current := -1, 0
	mutated := false
	index := func(p *Object) uint32 {
		for n, o := range objs {
			if o == p {
				return uint32(n)
			}
		}
		t.Fatal("foreign native object")
		return 0
	}
	observe := func(stage string, args ...uint32) {
		events = append(events, radialEventOriginal{Service: stage, Args: args, radialStateOriginal: state()})
	}
	mutate := func(stage string) {
		if mutated || i.MutationEvent != stage {
			return
		}
		mutated = true
		p := objs[current]
		switch i.MutationAction {
		case "position":
			p.PosVec = types.Ptf(600.1, -300.2)
		case "origin":
			*origin = types.Ptf(302.25, 299.75)
		case "force":
			p.ForceVec = types.Ptf(5.3, -8.9)
		case "mass":
			p.Mass = 5.5
		case "class":
			p.ObjClass = 1
		case "flags":
			p.ObjFlags |= 8
		case "radius":
			radius = 10.5
		case "inner":
			inner = -10.5
		case "scaled":
			scaled = 123.75
		case "origin-ref":
			origin = alternate
		case "callback":
			enabled = false
		default:
			t.Fatal("unknown original callback mutation")
		}
	}
	traceRay := func(from, to types.Pointf) bool {
		stage := fmt.Sprintf("trace:%d", visit)
		observe(stage, uint32(current), math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y))
		mutate(stage)
		return i.TraceMask>>visit&1 != 0
	}
	wake := func(p *Object) {
		if int(index(p)) != current {
			t.Fatal("wake target changed")
		}
		stage := fmt.Sprintf("wake:%d", visit)
		observe(stage, uint32(current))
		mutate(stage)
	}
	before := func(p *Object, distance float32) {
		if !enabled {
			return
		}
		if int(index(p)) != current {
			t.Fatal("callback target changed")
		}
		stage := fmt.Sprintf("before:%d", visit)
		observe(stage, uint32(current), math.Float32bits(distance), 0x12345678)
		mutate(stage)
	}
	if i.Kind == "candidate" {
		visit = 0
	}
	run(radialOpsOriginal{origin: func() *types.Pointf { return origin }, radius: func() float32 { return radius }, inner: func() float32 { return inner }, force: func() float32 { return scaled },
		each: func(rect types.Rectf, fn func(*Object) bool) {
			observe("enumerate", math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y))
			for n, v := range i.Order {
				visit, current = n, v
				if !fn(objs[v]) {
					t.Fatal("enumeration stopped early")
				}
			}
		},
		movable: (*Object).IsMovable, position: (*Object).Pos, mass: func(p *Object) float32 { return p.Mass }, missile: func(p *Object) bool { return p.ObjClass.Has(object.ClassMissile) }, loadForce: func(p *Object) types.Pointf { return p.ForceVec }, storeForce: func(p *Object, _ types.Pointf, v types.Pointf) { p.ForceVec = v }, trace: traceRay, before: before, wake: wake, first: objs[0]})
	if !row.Preserved || mutated != row.Mutated || !reflect.DeepEqual(state(), row.radialStateOriginal) || !reflect.DeepEqual(events, row.Events) {
		t.Fatalf("mutated=%t/%t state=%+v original=%+v events=%+v original=%+v", mutated, row.Mutated, state(), row.radialStateOriginal, events, row.Events)
	}
	for n, p := range frames {
		if !bytes.Equal(p, expected[n]) {
			t.Fatalf("native object frame %d differs from original output", n)
		}
	}
	// Literal point writes are permitted by declared origin callbacks. Guard
	// bytes around both points remain exact, including on alias cases.
	for _, p := range [][]byte{litFrame, altFrame} {
		for _, b := range append(bytes.Clone(p[:32]), p[40:]...) {
			if b != 0xa5 {
				t.Fatal("origin point canary changed")
			}
		}
	}
}
