package server

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type blackPowderInputOriginal struct {
	Kind           string    `json:"kind"`
	Name           string    `json:"name"`
	Creation       uint32    `json:"creation"`
	Fuse           uint32    `json:"fuse"`
	Frame          uint32    `json:"frame"`
	FPS            uint32    `json:"fps"`
	Position       [2]uint32 `json:"position"`
	RNG            int       `json:"rng"`
	Cache          [2]uint32 `json:"cache"`
	Lookup         [2]uint32 `json:"lookup"`
	TraceMask      uint32    `json:"trace_mask"`
	NewMask        uint32    `json:"new_mask"`
	MutationEvent  string    `json:"mutation_event"`
	MutationAction string    `json:"mutation_action"`
}
type blackPowderStateOriginal struct {
	Objects [][6]uint32 `json:"objects"`
	Cache   [2]uint32   `json:"cache"`
	Frame   uint32      `json:"frame"`
	FPS     uint32      `json:"fps"`
	RNG     int         `json:"rng"`
}
type blackPowderEventOriginal struct {
	Service string   `json:"service"`
	Args    []uint32 `json:"args"`
	blackPowderStateOriginal
}
type blackPowderRowOriginal struct {
	Mode      string                     `json:"mode"`
	Input     blackPowderInputOriginal   `json:"input"`
	Events    []blackPowderEventOriginal `json:"events"`
	Mutated   bool                       `json:"mutated"`
	Preserved bool                       `json:"nonoutput_bytes_preserved"`
	blackPowderStateOriginal
}

func TestBlackPowderStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/black_powder_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string            `json:"original_sha256"`
		Unchanged  bool              `json:"original_unchanged"`
		CW         uint16            `json:"actual_startup_control_word"`
		Contexts   int               `json:"contexts"`
		Executions int               `json:"root_executions"`
		Faults     []json.RawMessage `json:"memory_faults"`
		Directions bool              `json:"all_256_first_flame_directions"`
		Tables     []struct {
			Name string `json:"name"`
			Hex  string `json:"hex"`
		} `json:"tables"`
		Rows []blackPowderRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !f.Unchanged || f.CW != 0xe3f || f.Contexts != 471 || f.Executions != 942 || len(f.Rows) != 942 || len(f.Faults) != 0 || !f.Directions {
		t.Fatal("incomplete original BlackPowder fixture")
	}
	directions := 0
	for _, table := range f.Tables {
		if table.Name != "directions" {
			continue
		}
		b, err := hex.DecodeString(table.Hex)
		if err != nil || len(b) != 2048 {
			t.Fatal("original direction table")
		}
		for n := range 256 {
			cos, sin := SinCosDir(byte(n))
			if math.Float32bits(cos) != binary.LittleEndian.Uint32(b[8*n:]) || math.Float32bits(sin) != binary.LittleEndian.Uint32(b[8*n+4:]) {
				t.Fatalf("direction %d differs from the original table", n)
			}
		}
		directions++
	}
	if directions != 1 {
		t.Fatal("missing original direction table")
	}
	count := 0
	for _, row := range f.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) { testBlackPowderOriginal(t, row) })
	}
	if count != 471 {
		t.Fatalf("startup conditions=%d", count)
	}
}

func testBlackPowderOriginal(t *testing.T, row blackPowderRowOriginal) {
	i := row.Input
	var objs [5]*Object
	var frames [5][]byte
	var want [5][]byte
	for n := range objs {
		p, free := alloc.Make([]byte(nil), int(unsafe.Sizeof(Object{}))+64)
		t.Cleanup(free)
		for k := range p {
			p[k] = 0xa5
		}
		objs[n] = (*Object)(unsafe.Pointer(&p[32]))
		frames[n] = p
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(objs[n])) <= math.MaxUint32 {
			t.Fatal("native object frame below 4 GiB")
		}
		*objs[n] = Object{PosVec: types.Ptf(float32(900+n), float32(950+n)), Field38: 0x11223344, ObjFlags: 0x40}
	}
	source := objs[0]
	source.PosVec = rigidPointOriginal(i.Position)
	source.Field32 = i.Creation
	source.Field34 = i.Fuse
	for n, p := range frames {
		want[n] = bytes.Clone(p)
		o := objs[n]
		binary.LittleEndian.PutUint32(want[n][32+unsafe.Offsetof(o.PosVec):], row.Objects[n][0])
		binary.LittleEndian.PutUint32(want[n][36+unsafe.Offsetof(o.PosVec):], row.Objects[n][1])
		binary.LittleEndian.PutUint32(want[n][32+unsafe.Offsetof(o.Field34):], row.Objects[n][3])
	}
	frame, fps, cache := i.Frame, i.FPS, i.Cache
	random := prand.New(i.RNG)
	state := func() blackPowderStateOriginal {
		s := blackPowderStateOriginal{Objects: make([][6]uint32, 5), Cache: cache, Frame: frame, FPS: fps, RNG: random.Index()}
		for n, o := range objs {
			s.Objects[n] = [6]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), o.Field32, o.Field34, o.Field38, uint32(o.ObjFlags)}
		}
		return s
	}
	events := make([]blackPowderEventOriginal, 0)
	observe := func(service string, args ...uint32) {
		events = append(events, blackPowderEventOriginal{Service: service, Args: args, blackPowderStateOriginal: state()})
	}
	mutated := false
	mutate := func(stage string) {
		if mutated || i.MutationEvent != stage {
			return
		}
		mutated = true
		switch i.MutationAction {
		case "position":
			source.PosVec = types.Ptf(600.1, -300.2)
		case "frame":
			frame = 200
		case "fps":
			fps = 60
		case "cache":
			cache = [2]uint32{201, 202}
		default:
			t.Fatal("unknown live mutation")
		}
	}
	round := -1
	objIndex := func(o *Object) uint32 {
		if o == nil {
			return math.MaxUint32
		}
		for n, p := range objs {
			if o == p {
				return uint32(n)
			}
		}
		t.Fatal("foreign native object")
		return 0
	}
	newObject := func(index int) *Object {
		stage := fmt.Sprintf("new:%d", round)
		observe(stage, uint32(index))
		mutate(stage)
		if i.NewMask>>round&1 == 0 {
			return nil
		}
		return objs[round+1]
	}
	damage := func(p types.Pointf, outer, inner float32, amount int, kind object.DamageType, o *Object, excluded Obj) {
		if o != source || excluded != nil {
			t.Fatal("damage source/exclusion changed")
		}
		observe("damage", math.Float32bits(p.X), math.Float32bits(p.Y), math.Float32bits(outer), math.Float32bits(inner), uint32(amount), uint32(kind), 0, math.MaxUint32)
		mutate("damage")
	}
	deleteObject := func(o *Object) { observe("delete", objIndex(o)) }
	deps := blackPowderBarrelUpdateDeps53C9A0{
		frame: func() uint32 { return frame }, fps: func() uint32 { return fps },
		damageUnitsAround: damage,
		pushUnitsAround: func(p types.Pointf, outer, inner, force float32) {
			observe("push", math.Float32bits(p.X), math.Float32bits(p.Y), math.Float32bits(outer), math.Float32bits(inner), math.Float32bits(force), 0, 0, 0)
			mutate("push")
		},
		randomInt: func(min, max int) int {
			stage := ""
			switch {
			case min == 1 && max == 5:
				stage = "int:fuse"
			case min == 0 && max == 1:
				round++
				stage = fmt.Sprintf("int:choose:%d", round)
			case min == 0 && max == 255:
				stage = fmt.Sprintf("int:direction:%d", round)
			case min == 5 && max == 20:
				stage = fmt.Sprintf("int:seconds:%d", round)
			default:
				t.Fatal("unexpected random range")
			}
			observe(stage, uint32(min), uint32(max))
			mutate(stage)
			return random.IntClamp(min, max)
		},
		randomFloat: func(min, max float32) float64 {
			stage := fmt.Sprintf("float:%d", round)
			observe(stage, math.Float32bits(min), math.Float32bits(max))
			mutate(stage)
			return logicRandomFloat416030(random, min, max)
		},
		traceRay: func(from, to types.Pointf) bool {
			stage := fmt.Sprintf("trace:%d", round)
			observe(stage, math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y), 0, 0, 1)
			mutate(stage)
			return i.TraceMask>>round&1 != 0
		},
		loadFlameType:  func(index int) int { return int(int32(cache[index])) },
		storeFlameType: func(index, value int) { cache[index] = uint32(value) },
		typeIndexByID: func(name string) int {
			index := 0
			if name == "MediumFlame" {
				index = 1
			} else if name != "SmallFlame" {
				t.Fatal("unexpected flame type")
			}
			stage := "type:" + name
			observe(stage, uint32(index))
			mutate(stage)
			return int(int32(i.Lookup[index]))
		},
		newObjectByTypeIndex: newObject,
		createAt: func(o, owner *Object, p types.Pointf) {
			stage := fmt.Sprintf("create:%d", round)
			observe(stage, objIndex(o), objIndex(owner), math.Float32bits(p.X), math.Float32bits(p.Y), 0)
			mutate(stage)
			o.PosVec = p
		},
		setDecayTime: func(o *Object, delay uint32) {
			stage := fmt.Sprintf("decay:%d", round)
			observe(stage, objIndex(o), delay)
			mutate(stage)
		},
		delayedDelete: deleteObject,
	}
	switch i.Kind {
	case "barrel":
		blackPowderBarrelUpdateNative53C9A0(source, deps)
	case "burn":
		blackPowderBurnUpdateNative53CCB0(source, blackPowderBurnUpdateDeps53CCB0{frame: func() uint32 { return frame }, fps: func() uint32 { return fps }, damageUnitsAround: damage, delayedDelete: deleteObject})
	default:
		t.Fatal("unknown original root")
	}
	if !row.Preserved || mutated != row.Mutated || !reflect.DeepEqual(state(), row.blackPowderStateOriginal) || !reflect.DeepEqual(events, row.Events) {
		t.Fatalf("mutated=%t/%t state=%+v original=%+v events=%+v original=%+v", mutated, row.Mutated, state(), row.blackPowderStateOriginal, events, row.Events)
	}
	for n, p := range frames {
		if !bytes.Equal(p, want[n]) {
			t.Fatalf("native object frame %d differs", n)
		}
	}
}
