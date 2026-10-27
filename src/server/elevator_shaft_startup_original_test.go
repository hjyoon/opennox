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

type elevatorInputOriginal struct {
	Kind                  string    `json:"kind"`
	Name                  string    `json:"name"`
	Source                [2]uint32 `json:"source"`
	Target                [2]uint32 `json:"target"`
	Size                  [2]uint32 `json:"size"`
	Corners               [8]uint32 `json:"corners"`
	CandidateKind         ShapeKind `json:"candidate_kind"`
	CandidateSize         [2]uint32 `json:"candidate_size"`
	CandidateRadius       uint32    `json:"candidate_radius"`
	Height                int32     `json:"height"`
	AlternateHeight       int32     `json:"alternate_height"`
	LinkedHeight          int32     `json:"linked_height"`
	AlternateLinkedHeight int32     `json:"alternate_linked_height"`
	Z                     uint32    `json:"z"`
	Vertical              uint32    `json:"vertical"`
	Flags                 uint32    `json:"flags"`
	Class                 uint32    `json:"candidate_class"`
	Type                  uint16    `json:"candidate_type"`
	Move                  int32     `json:"move"`
	Initialized           uint32    `json:"initialized"`
	CachedIDs             [4]uint32 `json:"cached_ids"`
	LookupIDs             [4]uint32 `json:"lookup_ids"`
	Linked                bool      `json:"linked"`
	Mutation              int       `json:"mutation"`
}

type elevatorStateOriginal struct {
	New       [2]uint32  `json:"new"`
	Z         uint32     `json:"z"`
	Vertical  uint32     `json:"vertical"`
	Flags     uint32     `json:"flags"`
	Class     uint32     `json:"class_bits"`
	Type      uint16     `json:"type_index"`
	SyncMask  uint32     `json:"sync_mask"`
	Hole      [2]uint32  `json:"hole"`
	Field41   uint32     `json:"field41"`
	Field42   uint32     `json:"field42"`
	SyncWords [32]uint32 `json:"sync_words"`
	Update    string     `json:"update"`
}

type elevatorEventOriginal struct {
	Service         string                `json:"service"`
	Name            string                `json:"name"`
	Height          uint32                `json:"height"`
	First           elevatorStateOriginal `json:"first"`
	Candidate       elevatorStateOriginal `json:"candidate"`
	OldHeight       uint32                `json:"old_height"`
	AlternateHeight uint32                `json:"alternate_height"`
	OldLink         string                `json:"old_link"`
	AlternateLink   string                `json:"alternate_link"`
}

type elevatorRowOriginal struct {
	Mode            string                  `json:"mode"`
	Input           elevatorInputOriginal   `json:"input"`
	Events          []elevatorEventOriginal `json:"events"`
	First           elevatorStateOriginal   `json:"first"`
	Candidate       elevatorStateOriginal   `json:"candidate"`
	OldHeight       uint32                  `json:"old_height"`
	AlternateHeight uint32                  `json:"alternate_height"`
	OldLink         string                  `json:"old_link"`
	AlternateLink   string                  `json:"alternate_link"`
	CachedIDs       [4]uint32               `json:"cached_ids"`
	Initialized     uint32                  `json:"initialized"`
	Return          int                     `json:"return_int"`
	Preserved       bool                    `json:"nonoutput_bytes_preserved"`
}

func TestElevatorShaftStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/elevator_shaft_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string                `json:"original_sha256"`
		Unchanged bool                  `json:"original_unchanged"`
		CW        uint16                `json:"actual_startup_control_word"`
		Contexts  int                   `json:"contexts"`
		Faults    []json.RawMessage     `json:"memory_faults"`
		Rows      []elevatorRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != 1061 || len(fixture.Rows) != 2122 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete unchanged original lift observations")
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) { testElevatorShaftOriginal(t, row, false) })
	}
	if selected != 1061 {
		t.Fatalf("startup lift/point conditions=%d", selected)
	}
}

func TestElevatorShaftNativeLinksOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/elevator_shaft_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []elevatorRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" || row.Input.Kind != "shaft" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) { testElevatorShaftOriginal(t, row, true) })
	}
	if selected != 293 {
		t.Fatalf("native linked-shaft conditions=%d", selected)
	}
}

func testElevatorShaftOriginal(t *testing.T, row elevatorRowOriginal, nativeLinks bool) {
	t.Helper()
	i := row.Input
	a, fa := alloc.New(Object{})
	b, fb := alloc.New(Object{})
	c, fc := alloc.New(Object{})
	d, fd := alloc.New(Object{})
	u, fu := alloc.New(ElevatorUpdateData{})
	alt, falt := alloc.New(ElevatorUpdateData{})
	cu, fcu := alloc.New(ElevatorUpdateData{})
	du, fdu := alloc.New(ElevatorUpdateData{})
	for _, free := range []func(){fa, fb, fc, fd, fu, falt, fcu, fdu} {
		t.Cleanup(free)
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(c), unsafe.Pointer(d), unsafe.Pointer(u), unsafe.Pointer(alt), unsafe.Pointer(cu), unsafe.Pointer(du)} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native lift fixture below 4 GiB")
		}
	}
	s := newElevatorStateServer(t)
	for _, p := range []*Object{a, b, c, d} {
		*p = Object{TypeInd: 102, Field37: 0x12345678, Field38: 0x11223344, Pos39: rigidPointOriginal([2]uint32{0xa5a5a5a5, 0xa5a5a5a5}), Field41: 0xa5a5a5a5, Field42: 0xa5a5a5a5, Field29: 0x13572468, serverHandle: s.handle}
		for index := range p.Field140 {
			p.Field140[index] = 0xa5a50000 + uint32(257*index)
		}
	}
	a.NewPos = rigidPointOriginal(i.Source)
	b.NewPos = rigidPointOriginal(i.Target)
	c.NewPos = types.Ptf(900, 901)
	d.NewPos = types.Ptf(1200, 1201)
	b.TypeInd = i.Type
	b.ObjClass = object.Class(i.Class)
	b.ObjFlags = object.Flags(i.Flags)
	b.ZVal = math.Float32frombits(i.Z)
	b.Field27 = math.Float32frombits(i.Vertical)
	a.Shape.Kind = ShapeKindBox
	a.Shape.Box.W, a.Shape.Box.H = math.Float32frombits(i.Size[0]), math.Float32frombits(i.Size[1])
	fields := []*float32{&a.Shape.Box.LeftTop, &a.Shape.Box.LeftBottom, &a.Shape.Box.LeftBottom2, &a.Shape.Box.LeftTop2, &a.Shape.Box.RightTop, &a.Shape.Box.RightBottom, &a.Shape.Box.RightBottom2, &a.Shape.Box.RightTop2}
	for index, p := range fields {
		*p = math.Float32frombits(i.Corners[index])
	}
	b.Shape.Kind = i.CandidateKind
	b.Shape.Box.W, b.Shape.Box.H = math.Float32frombits(i.CandidateSize[0]), math.Float32frombits(i.CandidateSize[1])
	b.Shape.Circle.R = math.Float32frombits(i.CandidateRadius)
	rawU := func(p *ElevatorUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
	}
	for _, p := range []*ElevatorUpdateData{u, alt, cu, du} {
		for index := range rawU(p) {
			rawU(p)[index] = 0xa5
		}
		p.Field_1 = 0
	}
	u.Field_4 = uint32(i.Height)
	alt.Field_4 = uint32(i.AlternateHeight)
	cu.Field_4 = uint32(i.LinkedHeight)
	du.Field_4 = uint32(i.AlternateLinkedHeight)
	a.UpdateData, c.UpdateData, d.UpdateData = unsafe.Pointer(u), unsafe.Pointer(cu), unsafe.Pointer(du)
	oldLink, alternateLink := c, d
	if !i.Linked {
		oldLink = nil
	}
	if nativeLinks {
		a.SetElevatorLinkFor(unsafe.Pointer(u), oldLink)
	}
	label := func(p *Object) string {
		switch p {
		case a:
			return "first"
		case b:
			return "candidate"
		case c:
			return "linked"
		case d:
			return "alternate-linked"
		case nil:
			return "nil"
		}
		t.Fatal("unknown lift link")
		return ""
	}
	updateLabel := func(p unsafe.Pointer) string {
		switch p {
		case unsafe.Pointer(u):
			return "old"
		case unsafe.Pointer(alt):
			return "alternate"
		case unsafe.Pointer(cu):
			return "linked"
		case unsafe.Pointer(du):
			return "alternate-linked"
		case nil:
			return "nil"
		}
		t.Fatal("unknown lift update")
		return ""
	}
	state := func(p *Object) elevatorStateOriginal {
		return elevatorStateOriginal{rigidBitsOriginal(p.NewPos), math.Float32bits(p.ZVal), math.Float32bits(p.Field27), uint32(p.ObjFlags), uint32(p.ObjClass), p.TypeInd, p.Field38, rigidBitsOriginal(p.Pos39), p.Field41, p.Field42, p.Field140, updateLabel(p.UpdateData)}
	}
	rawObj := func(p *Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	wantObj := func(p *Object, w elevatorStateOriginal) []byte {
		raw := bytes.Clone(rawObj(p))
		base := uintptr(unsafe.Pointer(p))
		write := func(field unsafe.Pointer, value uint32) {
			binary.LittleEndian.PutUint32(raw[uintptr(field)-base:], value)
		}
		write(unsafe.Pointer(&p.NewPos.X), w.New[0])
		write(unsafe.Pointer(&p.NewPos.Y), w.New[1])
		write(unsafe.Pointer(&p.ZVal), w.Z)
		write(unsafe.Pointer(&p.Field27), w.Vertical)
		write(unsafe.Pointer(&p.ObjFlags), w.Flags)
		write(unsafe.Pointer(&p.ObjClass), w.Class)
		binary.LittleEndian.PutUint16(raw[unsafe.Offsetof(p.TypeInd):], w.Type)
		write(unsafe.Pointer(&p.Field38), w.SyncMask)
		write(unsafe.Pointer(&p.Pos39.X), w.Hole[0])
		write(unsafe.Pointer(&p.Pos39.Y), w.Hole[1])
		write(unsafe.Pointer(&p.Field41), w.Field41)
		write(unsafe.Pointer(&p.Field42), w.Field42)
		for index, value := range w.SyncWords {
			write(unsafe.Pointer(&p.Field140[index]), value)
		}
		var update unsafe.Pointer
		switch w.Update {
		case "old":
			update = unsafe.Pointer(u)
		case "alternate":
			update = unsafe.Pointer(alt)
		case "nil":
		default:
			t.Fatal("unknown expected update")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(raw[unsafe.Offsetof(p.UpdateData):], uint64(uintptr(update)))
		} else {
			binary.LittleEndian.PutUint32(raw[unsafe.Offsetof(p.UpdateData):], uint32(uintptr(update)))
		}
		return raw
	}
	wantA, wantB := wantObj(a, row.First), wantObj(b, row.Candidate)
	wantC, wantD := bytes.Clone(rawObj(c)), bytes.Clone(rawObj(d))
	wantU, wantAlt := bytes.Clone(rawU(u)), bytes.Clone(rawU(alt))
	wantCU, wantDU := bytes.Clone(rawU(cu)), bytes.Clone(rawU(du))
	binary.LittleEndian.PutUint32(wantU[unsafe.Offsetof(u.Field_4):], row.OldHeight)
	binary.LittleEndian.PutUint32(wantAlt[unsafe.Offsetof(alt.Field_4):], row.AlternateHeight)
	events := make([]elevatorEventOriginal, 0)
	observe := func(e elevatorEventOriginal) {
		e.First, e.Candidate = state(a), state(b)
		e.OldHeight, e.AlternateHeight = u.Field_4, alt.Field_4
		e.OldLink, e.AlternateLink = label(oldLink), label(alternateLink)
		events = append(events, e)
	}
	ids, initialized := i.CachedIDs, i.Initialized
	names := [4]string{"SmallFist", "MediumFist", "LargeFist", "Meteor"}
	readType := func(name string) int {
		if initialized == 0 {
			for index, name := range names {
				observe(elevatorEventOriginal{Service: "lookup", Name: name})
				switch {
				case i.Mutation == 1 && name == "SmallFist":
					a.UpdateData = unsafe.Pointer(alt)
				case i.Mutation == 2 && name == "SmallFist":
					u.Field_4 = 40
				case i.Mutation == 3 && name == "MediumFist":
					b.TypeInd = uint16(i.LookupIDs[0])
				case i.Mutation == 4 && name == "SmallFist":
					oldLink = d
					if nativeLinks {
						a.SetElevatorLinkFor(unsafe.Pointer(u), oldLink)
					}
				case i.Mutation == 5 && name == "LargeFist":
					b.ObjClass = 2
				}
				ids[index] = i.LookupIDs[index]
			}
			initialized = 1
		}
		for index, v := range names {
			if v == name {
				return int(ids[index])
			}
		}
		t.Fatal("unknown type query")
		return 0
	}
	raise := func(p *Object, z float32) {
		if p != b {
			t.Fatal("raise target changed")
		}
		observe(elevatorEventOriginal{Service: "raise", Height: math.Float32bits(z)})
		p.Raise(z)
	}
	ret := 0
	switch i.Kind {
	case "point":
		if MapPointInBox57B850(&a.NewPos, &a.Shape, &b.NewPos) {
			ret = 1
		}
	case "elevator":
		elevatorCollision551AE0(a, b, i.Move != 0, elevatorCollisionHooks551AE0{typeIndex: readType, raise: raise, circleBox: func(first, second *Object, move bool) {
			if first != b || second != a || move {
				t.Fatal("side collision arguments changed")
			}
			observe(elevatorEventOriginal{Service: "circle-box"})
		}, boxBox: func(first, second *Object) {
			if first != b || second != a {
				t.Fatal("box collision arguments changed")
			}
			observe(elevatorEventOriginal{Service: "box-box"})
		}})
	case "shaft":
		elevatorShaftCollision551C40(a, b, elevatorShaftCollisionHooks551C40{typeIndex: readType, raise: raise, link: func(p *Object, data unsafe.Pointer) *Object {
			if p != a {
				t.Fatal("link source changed")
			}
			if nativeLinks {
				return p.ElevatorLinkFor(data)
			}
			if data == unsafe.Pointer(u) {
				return oldLink
			}
			if data == unsafe.Pointer(alt) {
				return alternateLink
			}
			t.Fatal("unknown cached shaft data")
			return nil
		}})
	default:
		t.Fatal("unknown original lift entry", i.Kind)
	}
	if !row.Preserved || ret != row.Return || state(a) != row.First || state(b) != row.Candidate || !reflect.DeepEqual(events, row.Events) || ids != row.CachedIDs || initialized != row.Initialized || label(oldLink) != row.OldLink || label(alternateLink) != row.AlternateLink || !bytes.Equal(rawObj(a), wantA) || !bytes.Equal(rawObj(b), wantB) || !bytes.Equal(rawObj(c), wantC) || !bytes.Equal(rawObj(d), wantD) || !bytes.Equal(rawU(u), wantU) || !bytes.Equal(rawU(alt), wantAlt) || !bytes.Equal(rawU(cu), wantCU) || !bytes.Equal(rawU(du), wantDU) {
		t.Fatalf("native result=%d first=%+v candidate=%+v cache=%v/%d events=%+v; original=%d first=%+v candidate=%+v cache=%v/%d events=%+v; bytes=%t/%t/%t/%t", ret, state(a), state(b), ids, initialized, events, row.Return, row.First, row.Candidate, row.CachedIDs, row.Initialized, row.Events, bytes.Equal(rawObj(a), wantA), bytes.Equal(rawObj(b), wantB), bytes.Equal(rawU(u), wantU), bytes.Equal(rawU(alt), wantAlt))
	}
}
