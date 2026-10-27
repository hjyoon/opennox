package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type elevatorPassengerInputOriginal struct {
	Kind            string    `json:"kind"`
	Name            string    `json:"name"`
	Source          [2]uint32 `json:"source"`
	Target          [2]uint32 `json:"target"`
	Corners         [8]uint32 `json:"corners"`
	Size            [2]uint32 `json:"size"`
	CandidateKind   ShapeKind `json:"candidate_kind"`
	CandidateSize   [2]uint32 `json:"candidate_size"`
	Radius          uint32    `json:"radius"`
	Height          int64     `json:"height"`
	LinkedHeight    int64     `json:"linked_height"`
	State           byte      `json:"state"`
	LinkedState     byte      `json:"linked_state"`
	AlternateState  byte      `json:"alternate_state"`
	AlternateHeight int64     `json:"alternate_height"`
	ShaftState      byte      `json:"shaft_state"`
	Z               uint32    `json:"z"`
	Flags           uint32    `json:"flags"`
	Class           uint32    `json:"cls"`
	Linked          bool      `json:"linked"`
	Count           int       `json:"count"`
	Last            uint32    `json:"last"`
	TickRate        uint32    `json:"tick_rate"`
	Ticks           []uint32  `json:"ticks"`
	Material        uint16    `json:"material"`
	Subclass        uint32    `json:"subclass"`
	Direction       int32     `json:"direction"`
	MutationEvent   string    `json:"mutation_event"`
	MutationAction  string    `json:"mutation_action"`
}

type elevatorPassengerStateOriginal struct {
	Objects       [][]uint32 `json:"objects"`
	ObjectUpdates []int      `json:"object_updates"`
	Updates       [][5]int64 `json:"updates"`
}

type elevatorPassengerEventOriginal struct {
	Service  string    `json:"service"`
	Object   int       `json:"object"`
	Value    uint32    `json:"value"`
	Position [2]uint32 `json:"position"`
	elevatorPassengerStateOriginal
}

type elevatorPassengerRowOriginal struct {
	Mode       string                           `json:"mode"`
	Input      elevatorPassengerInputOriginal   `json:"input"`
	Events     []elevatorPassengerEventOriginal `json:"events"`
	Mutated    bool                             `json:"mutated"`
	ReturnByte uint8                            `json:"return_byte"`
	Preserved  bool                             `json:"nonoutput_bytes_preserved"`
	elevatorPassengerStateOriginal
}

func elevatorPassengerOriginalRows(t *testing.T) []elevatorPassengerRowOriginal {
	t.Helper()
	blob, err := os.ReadFile("testdata/elevator_passenger_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string                         `json:"original_sha256"`
		Unchanged  bool                           `json:"original_unchanged"`
		CW         uint16                         `json:"actual_startup_control_word"`
		Contexts   int                            `json:"contexts"`
		Executions int                            `json:"root_executions"`
		Faults     []json.RawMessage              `json:"memory_faults"`
		Rows       []elevatorPassengerRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !f.Unchanged || f.CW != 0xe3f || f.Contexts != 1229 || f.Executions != 2458 || len(f.Rows) != 2458 || len(f.Faults) != 0 {
		t.Fatal("incomplete original passenger fixture")
	}
	return f.Rows
}

func TestElevatorPassengerStartupOriginal(t *testing.T) {
	count := 0
	for _, row := range elevatorPassengerOriginalRows(t) {
		if row.Mode != "actual-startup" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) { testElevatorPassengerOriginal(t, row, false) })
	}
	if count != 1229 {
		t.Fatalf("startup conditions=%d", count)
	}
}

func TestElevatorPassengerNativeLinksOriginal(t *testing.T) {
	count := 0
	for _, row := range elevatorPassengerOriginalRows(t) {
		if row.Mode != "actual-startup" || row.Input.MutationAction == "owner-update" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) { testElevatorPassengerOriginal(t, row, true) })
	}
	if count != 1221 {
		t.Fatalf("native link conditions=%d", count)
	}
}

func testElevatorPassengerOriginal(t *testing.T, row elevatorPassengerRowOriginal, nativeLinks bool) {
	t.Helper()
	i := row.Input
	s := newElevatorStateServer(t)
	var objects [5]*Object
	var updates [4]*ElevatorUpdateData
	for n := range objects {
		p, free := alloc.New(Object{})
		t.Cleanup(free)
		objects[n] = p
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
			t.Fatal("object below 4 GiB")
		}
		*p = Object{TypeInd: 40, serverHandle: s.handle, Field37: 0x12345678, Field38: 0x11223344, Field29: 0x87654321}
		for k := range p.Field140 {
			p.Field140[k] = 0xa5a50000 + uint32(257*k)
		}
		pos := i.Target
		if n == 0 {
			pos = i.Source
		} else if n >= 3 {
			pos = [2]uint32{math.Float32bits(float32(900 + 300*(n-3))), math.Float32bits(float32(901 + 300*(n-3)))}
		}
		p.PosVec = rigidPointOriginal(pos)
		p.NewPos = p.PosVec
		p.PrevPos = p.PosVec
		p.Shape.Kind = ShapeKindBox
		p.Shape.Box.W, p.Shape.Box.H = math.Float32frombits(i.Size[0]), math.Float32frombits(i.Size[1])
		fields := []*float32{&p.Shape.Box.LeftTop, &p.Shape.Box.LeftBottom, &p.Shape.Box.LeftBottom2, &p.Shape.Box.LeftTop2, &p.Shape.Box.RightTop, &p.Shape.Box.RightBottom, &p.Shape.Box.RightBottom2, &p.Shape.Box.RightTop2}
		for index, field := range fields {
			*field = math.Float32frombits(i.Corners[index])
		}
		if n == 1 || n == 2 {
			p.Shape.Kind = i.CandidateKind
			p.Shape.Box.W, p.Shape.Box.H = math.Float32frombits(i.CandidateSize[0]), math.Float32frombits(i.CandidateSize[1])
			p.Shape.Circle.R = math.Float32frombits(i.Radius)
			p.ObjClass = object.Class(i.Class)
			p.ObjFlags = 0x100040
			p.ZVal = math.Float32frombits(i.Z)
			p.Field27 = 3.5
		}
	}
	owner := objects[0]
	owner.ObjFlags = object.Flags(i.Flags)
	owner.Field34 = i.Last
	owner.Material = i.Material
	owner.ObjSubClass = object.SubClass(i.Subclass)
	for n := range updates {
		p, free := alloc.New(ElevatorUpdateData{})
		t.Cleanup(free)
		updates[n] = p
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
			t.Fatal("update below 4 GiB")
		}
		raw := unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
		for k := range raw {
			raw[k] = 0xa5
		}
		p.Field_0 = 0x10203040 + uint32(n)
		p.Field_1 = 0
		p.Field_2 = 0x55667700 + uint32(n)
		p.Field_3 = [4]byte{i.State, i.AlternateState, i.LinkedState, i.AlternateState}[n]
		if n == 0 && i.Kind == "shaft" {
			p.Field_3 = i.ShaftState
		}
		p.Field_4 = uint32([4]int64{i.Height, i.AlternateHeight, i.LinkedHeight, i.AlternateHeight}[n])
	}
	owner.UpdateData = unsafe.Pointer(updates[0])
	objects[3].UpdateData = unsafe.Pointer(updates[2])
	objects[4].UpdateData = unsafe.Pointer(updates[3])
	links := [4]*Object{objects[3], objects[4], owner, owner}
	if !i.Linked {
		links[0] = nil
	}
	if nativeLinks {
		owner.SetElevatorLinkFor(unsafe.Pointer(updates[0]), links[0])
		objects[3].SetElevatorLinkFor(unsafe.Pointer(updates[2]), owner)
		objects[4].SetElevatorLinkFor(unsafe.Pointer(updates[3]), owner)
	}
	objIndex := func(p *Object) int {
		if p == nil {
			return -1
		}
		for n, v := range objects {
			if p == v {
				return n
			}
		}
		t.Fatal("unknown object")
		return -2
	}
	updateIndex := func(p unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for n, v := range updates {
			if p == unsafe.Pointer(v) {
				return n
			}
		}
		t.Fatal("unknown update")
		return -2
	}
	wordOffsets := []uintptr{unsafe.Offsetof(owner.ObjClass), unsafe.Offsetof(owner.ObjSubClass), unsafe.Offsetof(owner.ObjFlags), unsafe.Offsetof(owner.Material), unsafe.Offsetof(owner.PosVec), unsafe.Offsetof(owner.PosVec) + 4, unsafe.Offsetof(owner.NewPos), unsafe.Offsetof(owner.NewPos) + 4, unsafe.Offsetof(owner.PrevPos), unsafe.Offsetof(owner.PrevPos) + 4, unsafe.Offsetof(owner.ZVal), unsafe.Offsetof(owner.Field27), unsafe.Offsetof(owner.Field34), unsafe.Offsetof(owner.Field37), unsafe.Offsetof(owner.Field38)}
	for n := 0; n < 32; n++ {
		wordOffsets = append(wordOffsets, unsafe.Offsetof(owner.Field140)+uintptr(4*n))
	}
	rawObj := func(p *Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	rawUD := func(p *ElevatorUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
	}
	state := func() elevatorPassengerStateOriginal {
		v := elevatorPassengerStateOriginal{Objects: make([][]uint32, 5), ObjectUpdates: make([]int, 5), Updates: make([][5]int64, 4)}
		for n, p := range objects {
			for _, off := range wordOffsets {
				v.Objects[n] = append(v.Objects[n], binary.LittleEndian.Uint32(rawObj(p)[off:]))
			}
			v.ObjectUpdates[n] = updateIndex(p.UpdateData)
		}
		for n, p := range updates {
			v.Updates[n] = [5]int64{int64(p.Field_0), int64(objIndex(links[n])), int64(p.Field_2), int64(binary.LittleEndian.Uint32(rawUD(p)[12:])), int64(p.Field_4)}
		}
		return v
	}
	wantObjects := make([][]byte, 5)
	wantUpdates := make([][]byte, 4)
	for n, p := range objects {
		want := bytes.Clone(rawObj(p))
		wantObjects[n] = want
		for k, off := range wordOffsets {
			binary.LittleEndian.PutUint32(want[off:], row.Objects[n][k])
		}
		var ptr unsafe.Pointer
		if row.ObjectUpdates[n] >= 0 {
			ptr = unsafe.Pointer(updates[row.ObjectUpdates[n]])
		}
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(want[unsafe.Offsetof(p.UpdateData):], uint64(uintptr(ptr)))
		} else {
			binary.LittleEndian.PutUint32(want[unsafe.Offsetof(p.UpdateData):], uint32(uintptr(ptr)))
		}
	}
	for n, p := range updates {
		want := bytes.Clone(rawUD(p))
		wantUpdates[n] = want
		for k, value := range row.Updates[n] {
			if k != 1 {
				binary.LittleEndian.PutUint32(want[4*k:], uint32(value))
			}
		}
	}
	events := make([]elevatorPassengerEventOriginal, 0)
	mutated := false
	observe := func(e elevatorPassengerEventOriginal) {
		e.elevatorPassengerStateOriginal = state()
		events = append(events, e)
	}
	mutate := func(stage string) {
		if mutated || i.MutationEvent != stage {
			return
		}
		mutated = true
		switch i.MutationAction {
		case "owner-update":
			owner.UpdateData = unsafe.Pointer(updates[1])
		case "owner-height":
			updates[0].Field_4 = 16777219
		case "linked-height":
			updates[2].Field_4 = 16777219
		case "owner-link":
			links[0] = objects[4]
			if nativeLinks {
				owner.SetElevatorLinkFor(unsafe.Pointer(updates[0]), links[0])
			}
		case "linked-state":
			updates[2].Field_3 = 5
		case "owner-state":
			updates[0].Field_3 = 6
		case "linked-update":
			objects[3].UpdateData = unsafe.Pointer(updates[3])
		case "candidate-class":
			objects[1].ObjClass = 2
		default:
			t.Fatal("unknown mutation")
		}
	}
	linkFor := func(p *Object, data unsafe.Pointer) *Object {
		if nativeLinks {
			return p.ElevatorLinkFor(data)
		}
		n := updateIndex(data)
		if n < 0 {
			return nil
		}
		return links[n]
	}
	var frame uint32
	hooks := elevatorUpdateHooks53B5D0{
		frame: func() uint32 { return frame }, tickRate: func() uint32 { return i.TickRate },
		link: linkFor,
		needSync: func(p *Object) {
			observe(elevatorPassengerEventOriginal{Service: "sync", Object: objIndex(p)})
			p.NeedSync()
		},
		audio: func(p *Object, up bool) {
			if id := elevatorSound53B490(p, up); id != 0 {
				observe(elevatorPassengerEventOriginal{Service: "audio", Object: objIndex(p), Value: uint32(id)})
				mutate("audio")
			}
		},
		pointInBox: MapPointInBox57B850,
		move: func(p *Object, pos types.Pointf) {
			observe(elevatorPassengerEventOriginal{Service: "move", Object: objIndex(p), Position: rigidBitsOriginal(pos)})
			p.PosVec = pos
			p.NewPos = pos
			p.PrevPos = pos
			mutate("move")
		},
		raise: func(p *Object, z float32) {
			observe(elevatorPassengerEventOriginal{Service: "raise", Object: objIndex(p), Value: math.Float32bits(z)})
			p.Raise(z)
		},
		queueCollision: func(p *Object) {
			observe(elevatorPassengerEventOriginal{Service: "queue", Object: objIndex(p)})
			mutate("queue")
		},
	}
	hooks.eachInCircle = func(pos types.Pointf, radius float32, fn func(*Object) bool) {
		if i.Kind == "elevator" || i.Kind == "shaft" {
			observe(elevatorPassengerEventOriginal{Service: "enumerate", Position: rigidBitsOriginal(pos), Value: math.Float32bits(radius)})
			mutate("enumerate")
		}
		for n := 0; n < i.Count; n++ {
			fn(objects[n+1])
		}
	}
	for _, tick := range i.Ticks {
		frame = tick
		switch i.Kind {
		case "elevator":
			elevatorUpdate53B5D0(owner, hooks)
		case "shaft":
			elevatorShaftUpdate53B380(owner, hooks)
		case "up":
			elevatorCarryUp53B750(owner, hooks)
		case "down":
			elevatorShaftCarryDown53B410(owner, hooks)
		case "audio":
			hooks.audio(owner, i.Direction != 0)
		default:
			t.Fatal("unknown original root")
		}
	}
	if !row.Preserved || mutated != row.Mutated || !reflect.DeepEqual(state(), row.elevatorPassengerStateOriginal) || !reflect.DeepEqual(events, row.Events) {
		t.Fatalf("native state/events differ: mutated=%t/%t state=%+v original=%+v events=%+v original=%+v", mutated, row.Mutated, state(), row.elevatorPassengerStateOriginal, events, row.Events)
	}
	for n, p := range objects {
		if !bytes.Equal(rawObj(p), wantObjects[n]) {
			t.Fatalf("object %d bytes differ", n)
		}
	}
	for n, p := range updates {
		if !bytes.Equal(rawUD(p), wantUpdates[n]) {
			t.Fatalf("update %d bytes differ", n)
		}
	}
}
