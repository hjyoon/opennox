package server

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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

type rigidInputOriginal struct {
	Kind              string    `json:"kind"`
	Name              string    `json:"name"`
	FirstNew          [2]uint32 `json:"first_new_bits32"`
	SecondNew         [2]uint32 `json:"second_new_bits32"`
	FirstSize         [2]uint32 `json:"first_size_bits32"`
	SecondSize        [2]uint32 `json:"second_size_bits32"`
	FirstMass         uint32    `json:"first_mass_bits32"`
	SecondMass        uint32    `json:"second_mass_bits32"`
	Radius            uint32    `json:"second_radius_bits32"`
	FirstAccumulator  [2]uint32 `json:"first_accumulator_bits32"`
	SecondAccumulator [2]uint32 `json:"second_accumulator_bits32"`
	Velocity          [2]uint32 `json:"second_velocity_bits32"`
	FirstFlags        uint32    `json:"first_flags"`
	SecondFlags       uint32    `json:"second_flags"`
	FirstClass        uint32    `json:"first_class"`
	SecondClass       uint32    `json:"second_class"`
	FirstTeam         byte      `json:"first_team"`
	SecondTeam        byte      `json:"second_team"`
	Owner             string    `json:"owner"`
	Move              int32     `json:"move"`
	Fractional        int16     `json:"fractional"`
	Lock              byte      `json:"lock"`
	Current           int32     `json:"current_direction"`
	Target            int32     `json:"target_direction"`
	Angular           uint32    `json:"angular_bits32"`
	Frame             uint32    `json:"frame"`
	TickRate          uint32    `json:"tick_rate"`
	Field34           uint32    `json:"field34"`
	Push              uint32    `json:"push_scale_bits32"`
	Spring            uint32    `json:"spring_bits32"`
	Mutation          int       `json:"mutation"`
	Segment           [4]uint32 `json:"segment_bits32"`
	Rect              [4]uint32 `json:"rect_bits32"`
	SegmentBounds     [4]uint32 `json:"segment_bounds_bits32"`
	Edge              [3]uint32 `json:"edge_bits32"`
	Point             [2]uint32 `json:"point_bits32"`
	AllowNegative     bool      `json:"allow_negative"`
	Limit             int       `json:"limit"`
}

type rigidObjectOriginal struct {
	New         [2]uint32 `json:"new_bits32"`
	Velocity    [2]uint32 `json:"velocity_bits32"`
	Accumulator [2]uint32 `json:"accumulator_bits32"`
	Class       uint32    `json:"class_bits"`
	Flags       uint32    `json:"flags"`
	Mass        uint32    `json:"mass_bits32"`
	Radius      uint32    `json:"radius_bits32"`
	Size        [2]uint32 `json:"size_bits32"`
	Field34     uint32    `json:"field34"`
	Team        byte      `json:"team"`
	Owner       string    `json:"owner"`
	Update      string    `json:"update"`
}

type rigidUpdateOriginal struct {
	Lock       byte   `json:"lock"`
	Target     int32  `json:"target"`
	Current    int32  `json:"current"`
	Queued     uint32 `json:"queued"`
	Angular    uint32 `json:"angular_bits32"`
	Fractional int16  `json:"fractional"`
	LastMove   uint32 `json:"last_move_frame"`
}

type rigidEventOriginal struct {
	Service         string              `json:"service"`
	Object          string              `json:"object"`
	Other           string              `json:"other"`
	Normal          []uint32            `json:"normal_bits32"`
	Data            string              `json:"data"`
	Message         string              `json:"message"`
	MessageFlags    uint32              `json:"message_flags"`
	First           rigidObjectOriginal `json:"first"`
	Second          rigidObjectOriginal `json:"second"`
	OldUpdate       rigidUpdateOriginal `json:"old_update"`
	AlternateUpdate rigidUpdateOriginal `json:"alternate_update"`
}

type rigidRowOriginal struct {
	Mode            string               `json:"mode"`
	Input           rigidInputOriginal   `json:"input"`
	Events          []rigidEventOriginal `json:"events"`
	First           rigidObjectOriginal  `json:"first"`
	Second          rigidObjectOriginal  `json:"second"`
	OldUpdate       rigidUpdateOriginal  `json:"old_update"`
	AlternateUpdate rigidUpdateOriginal  `json:"alternate_update"`
	Push            uint32               `json:"push_scale_bits32"`
	Spring          uint32               `json:"spring_bits32"`
	Count           int                  `json:"return_count"`
	Points          [][2]uint32          `json:"points_bits32"`
	Preserved       bool                 `json:"nonoutput_bytes_preserved"`
}

func rigidPointOriginal(w [2]uint32) types.Pointf {
	return types.Ptf(math.Float32frombits(w[0]), math.Float32frombits(w[1]))
}
func rigidBitsOriginal(p types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}
}
func rigidRectOriginal(w [4]uint32) doorBoxRect551250 {
	return doorBoxRect551250{minX: math.Float32frombits(w[0]), minY: math.Float32frombits(w[1]), maxX: math.Float32frombits(w[2]), maxY: math.Float32frombits(w[3])}
}

func TestRigidCollisionStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/rigid_collision_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Contexts  int               `json:"contexts"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Tables    []struct {
			Name string `json:"name"`
			Hex  string `json:"hex"`
		} `json:"tables"`
		Rows []rigidRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != 1263 || len(fixture.Rows) != 2526 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete unchanged original rigid collision observations")
	}
	tables := make(map[string][]byte)
	for _, table := range fixture.Tables {
		raw, err := hex.DecodeString(table.Hex)
		if err != nil {
			t.Fatal(err)
		}
		tables[table.Name] = raw
	}
	if len(tables["integer-directions"]) != 2048 || len(tables["float-directions"]) != 2048 {
		t.Fatal("incomplete original direction tables")
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			if !row.Preserved {
				t.Fatal("original nonoutput canary changed")
			}
			if row.Input.Kind != "box-box" && row.Input.Kind != "door-circle" && row.Input.Kind != "door-box" {
				testRigidGeometryOriginal(t, row)
				return
			}
			testRigidObjectsOriginal(t, row, tables)
		})
	}
	if selected != 1263 {
		t.Fatalf("startup conditions=%d", selected)
	}
}

func testRigidGeometryOriginal(t *testing.T, row rigidRowOriginal) {
	t.Helper()
	i := row.Input
	segment := doorBoxSegment551250{start: rigidPointOriginal([2]uint32{i.Segment[0], i.Segment[1]}), end: rigidPointOriginal([2]uint32{i.Segment[2], i.Segment[3]})}
	rect := rigidRectOriginal(i.Rect)
	points := make([][2]uint32, 0)
	count := 0
	switch i.Kind {
	case "horizontal":
		p, ok := doorBoxHorizontalIntersection551780(segment, math.Float32frombits(i.Edge[0]), math.Float32frombits(i.Edge[1]), math.Float32frombits(i.Edge[2]), i.AllowNegative)
		if ok {
			count = 1
			points = append(points, rigidBitsOriginal(p))
		}
	case "vertical":
		p, ok := doorBoxVerticalIntersection551870(segment, math.Float32frombits(i.Edge[0]), math.Float32frombits(i.Edge[1]), math.Float32frombits(i.Edge[2]), i.AllowNegative)
		if ok {
			count = 1
			points = append(points, rigidBitsOriginal(p))
		}
	case "intersections":
		for _, p := range doorBoxSegmentRectIntersections5516A0(segment, rect, i.Limit, i.AllowNegative) {
			points = append(points, rigidBitsOriginal(p))
			count++
		}
	case "midpoint":
		p, ok := doorBoxCollisionMidpoint551960(segment, rigidRectOriginal(i.SegmentBounds), rect)
		if ok {
			count = 1
			points = append(points, rigidBitsOriginal(p))
		}
	case "point-in-rect":
		if doorBoxPointInRect551A90(rigidPointOriginal(i.Point), rect) {
			count = 1
		}
	default:
		t.Fatal("unknown original geometry entry", i.Kind)
	}
	if count != row.Count || !reflect.DeepEqual(points, row.Points) {
		t.Fatalf("native geometry count=%d points=%08x original count=%d points=%08x", count, points, row.Count, row.Points)
	}
}

func testRigidObjectsOriginal(t *testing.T, row rigidRowOriginal, tables map[string][]byte) {
	t.Helper()
	i := row.Input
	a, freeA := alloc.New(Object{})
	b, freeB := alloc.New(Object{})
	other, freeOther := alloc.New(Object{})
	u, freeU := alloc.New(DoorUpdateData{})
	alt, freeAlt := alloc.New(DoorUpdateData{})
	for _, free := range []func(){freeA, freeB, freeOther, freeU, freeAlt} {
		t.Cleanup(free)
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(other), unsafe.Pointer(u), unsafe.Pointer(alt)} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native rigid input below 4 GiB")
		}
	}
	*a = Object{NewPos: rigidPointOriginal(i.FirstNew), Pos24: rigidPointOriginal(i.FirstAccumulator), Mass: math.Float32frombits(i.FirstMass), ObjClass: object.Class(i.FirstClass), ObjFlags: object.Flags(i.FirstFlags), Field34: i.Field34, Field29: 0x13572468, Field33: 77, PosVec: types.Ptf(10, 20), UpdateData: unsafe.Pointer(u)}
	*b = Object{NewPos: rigidPointOriginal(i.SecondNew), Pos24: rigidPointOriginal(i.SecondAccumulator), VelVec: rigidPointOriginal(i.Velocity), Mass: math.Float32frombits(i.SecondMass), ObjClass: object.Class(i.SecondClass), ObjFlags: object.Flags(i.SecondFlags), Field29: 0x87654321, Field33: 99, PosVec: types.Ptf(30, 40)}
	*other = Object{}
	a.Shape.Kind, b.Shape.Kind = ShapeKindBox, ShapeKindBox
	a.Shape.Box.W, a.Shape.Box.H = math.Float32frombits(i.FirstSize[0]), math.Float32frombits(i.FirstSize[1])
	b.Shape.Box.W, b.Shape.Box.H = math.Float32frombits(i.SecondSize[0]), math.Float32frombits(i.SecondSize[1])
	a.Shape.Circle.R = 4
	b.Shape.Circle.R = math.Float32frombits(i.Radius)
	if i.Kind == "door-circle" {
		b.Shape.Kind = ShapeKindCircle
	}
	a.TeamVal.ID, b.TeamVal.ID = TeamID(i.FirstTeam), TeamID(i.SecondTeam)
	objectByLabel := func(label string) *Object {
		switch label {
		case "first":
			return a
		case "second":
			return b
		case "other":
			return other
		case "nil":
			return nil
		default:
			t.Fatal("unknown pointer label", label)
			return nil
		}
	}
	a.ObjOwner = objectByLabel(i.Owner)
	rawU := func(p *DoorUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
	}
	for _, p := range []*DoorUpdateData{u, alt} {
		for index := range rawU(p) {
			rawU(p)[index] = 0xa5
		}
		p.Queued = 0
	}
	u.LockCode = i.Lock
	u.TargetDirection = i.Target
	u.CurrentDirection = i.Current
	u.AngularVelocity = math.Float32frombits(i.Angular)
	u.FractionalDir = i.Fractional
	u.LastMoveFrame = 7
	alt.LockCode = 1
	alt.TargetDirection = 9
	alt.CurrentDirection = 9
	alt.AngularVelocity = -.5
	alt.FractionalDir = 64
	alt.LastMoveFrame = 13
	label := func(p *Object) string {
		switch p {
		case a:
			return "first"
		case b:
			return "second"
		case other:
			return "other"
		case nil:
			return "nil"
		default:
			t.Fatal("unknown native object", p)
			return ""
		}
	}
	updateLabel := func(p unsafe.Pointer) string {
		switch p {
		case unsafe.Pointer(u):
			return "old"
		case unsafe.Pointer(alt):
			return "alternate"
		case nil:
			return "nil"
		default:
			t.Fatal("unknown native update pointer")
			return ""
		}
	}
	state := func(p *Object) rigidObjectOriginal {
		return rigidObjectOriginal{New: rigidBitsOriginal(p.NewPos), Velocity: rigidBitsOriginal(p.VelVec), Accumulator: rigidBitsOriginal(p.Pos24), Class: uint32(p.ObjClass), Flags: uint32(p.ObjFlags), Mass: math.Float32bits(p.Mass), Radius: math.Float32bits(p.Shape.Circle.R), Size: [2]uint32{math.Float32bits(p.Shape.Box.W), math.Float32bits(p.Shape.Box.H)}, Field34: p.Field34, Team: byte(p.TeamVal.ID), Owner: label(p.ObjOwner), Update: updateLabel(p.UpdateData)}
	}
	updateState := func(p *DoorUpdateData) rigidUpdateOriginal {
		return rigidUpdateOriginal{Lock: p.LockCode, Target: p.TargetDirection, Current: p.CurrentDirection, Queued: p.Queued, Angular: math.Float32bits(p.AngularVelocity), Fractional: p.FractionalDir, LastMove: p.LastMoveFrame}
	}
	rawObj := func(p *Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	wantObj := func(p *Object, w rigidObjectOriginal) []byte {
		result := bytes.Clone(rawObj(p))
		base := uintptr(unsafe.Pointer(p))
		write := func(field unsafe.Pointer, value uint32) {
			binary.LittleEndian.PutUint32(result[uintptr(field)-base:], value)
		}
		writePoint := func(field *types.Pointf, value [2]uint32) {
			write(unsafe.Pointer(&field.X), value[0])
			write(unsafe.Pointer(&field.Y), value[1])
		}
		writePointer := func(field unsafe.Pointer, value unsafe.Pointer) {
			off := uintptr(field) - base
			if unsafe.Sizeof(uintptr(0)) == 8 {
				binary.LittleEndian.PutUint64(result[off:], uint64(uintptr(value)))
			} else {
				binary.LittleEndian.PutUint32(result[off:], uint32(uintptr(value)))
			}
		}
		writePoint(&p.NewPos, w.New)
		writePoint(&p.VelVec, w.Velocity)
		writePoint(&p.Pos24, w.Accumulator)
		write(unsafe.Pointer(&p.ObjClass), w.Class)
		write(unsafe.Pointer(&p.ObjFlags), w.Flags)
		write(unsafe.Pointer(&p.Mass), w.Mass)
		write(unsafe.Pointer(&p.Shape.Circle.R), w.Radius)
		write(unsafe.Pointer(&p.Shape.Box.W), w.Size[0])
		write(unsafe.Pointer(&p.Shape.Box.H), w.Size[1])
		write(unsafe.Pointer(&p.Field34), w.Field34)
		result[uintptr(unsafe.Pointer(&p.TeamVal.ID))-base] = w.Team
		writePointer(unsafe.Pointer(&p.ObjOwner), unsafe.Pointer(objectByLabel(w.Owner)))
		var update unsafe.Pointer
		switch w.Update {
		case "old":
			update = unsafe.Pointer(u)
		case "alternate":
			update = unsafe.Pointer(alt)
		case "nil":
		default:
			t.Fatal("unknown expected update label")
		}
		writePointer(unsafe.Pointer(&p.UpdateData), update)
		return result
	}
	wantU := func(p *DoorUpdateData, w rigidUpdateOriginal) []byte {
		result := bytes.Clone(rawU(p))
		binary.LittleEndian.PutUint32(result[unsafe.Offsetof(p.Queued):], w.Queued)
		binary.LittleEndian.PutUint32(result[unsafe.Offsetof(p.AngularVelocity):], w.Angular)
		binary.LittleEndian.PutUint32(result[unsafe.Offsetof(p.LastMoveFrame):], w.LastMove)
		return result
	}
	expectedA, expectedB, expectedU, expectedAlt := wantObj(a, row.First), wantObj(b, row.Second), wantU(u, row.OldUpdate), wantU(alt, row.AlternateUpdate)
	events := make([]rigidEventOriginal, 0)
	appendEvent := func(e rigidEventOriginal) {
		e.First = state(a)
		e.Second = state(b)
		e.OldUpdate = updateState(u)
		e.AlternateUpdate = updateState(alt)
		events = append(events, e)
	}
	integerDirections := bytes.Clone(tables["integer-directions"])
	floatDirections := tables["float-directions"]
	pushScale, spring := math.Float32frombits(i.Push), math.Float32frombits(i.Spring)
	directionInt := func(dir byte) (int32, int32) {
		offset := int(dir) * 8
		return int32(binary.LittleEndian.Uint32(integerDirections[offset:])), int32(binary.LittleEndian.Uint32(integerDirections[offset+4:]))
	}
	directionFloat := func(dir byte) types.Pointf {
		offset := int(dir) * 8
		return types.Ptf(math.Float32frombits(binary.LittleEndian.Uint32(floatDirections[offset:])), math.Float32frombits(binary.LittleEndian.Uint32(floatDirections[offset+4:])))
	}
	addCollision := func(first, second *Object, normal types.Pointf) {
		bits := rigidBitsOriginal(normal)
		appendEvent(rigidEventOriginal{Service: "collision", Object: label(first), Other: label(second), Normal: []uint32{bits[0], bits[1]}})
		switch i.Mutation {
		case 1:
			b.Mass = 7.25
			b.VelVec = types.Ptf(2, -3.5)
			b.NewPos = types.Ptf(301.25, 298.5)
			b.Shape.Circle.R = 12
			b.Shape.Box.W = 22
			b.Shape.Box.H = 18
		case 2:
			a.UpdateData = unsafe.Pointer(alt)
		case 3:
			pushScale = 23.75
			spring = 77.25
		case 4:
			a.ObjFlags ^= 8
			b.ObjFlags ^= 0x2000
		case 7:
			index := (int(i.Fractional) + 160) % 256
			binary.LittleEndian.PutUint32(integerDirections[index*8:], 11)
			binary.LittleEndian.PutUint32(integerDirections[index*8+4:], uint32(0xfffffffd))
		}
	}
	wake := func(p *Object) {
		appendEvent(rigidEventOriginal{Service: "wake", Object: label(p)})
		if i.Mutation == 5 {
			p.ObjFlags ^= 0x40
			if p == a {
				b.ObjFlags |= 0x08000000
			}
		} else if i.Mutation == 6 && p == a {
			b.Shape.Circle.R = 14
			b.Mass = 6.25
			b.NewPos = types.Ptf(297.5, 302.25)
			a.TeamVal.ID = 1
			b.TeamVal.ID = 1
		}
	}
	queue := func(p *DoorUpdateData) {
		data := ""
		if p == u {
			data = "old"
		} else if p == alt {
			data = "alternate"
		} else {
			t.Fatal("unknown queued native update")
		}
		appendEvent(rigidEventOriginal{Service: "door-queue", Data: data})
	}
	updatable := func(p *Object) { appendEvent(rigidEventOriginal{Service: "updatable", Object: label(p)}) }
	message := func(p *Object, text string) {
		appendEvent(rigidEventOriginal{Service: "message", Object: label(p), Message: text})
	}
	switch i.Kind {
	case "box-box":
		boxBoxCollision550F80(a, b, boxBoxCollisionHooks550F80{pushScale: func() float32 { return pushScale }, addCollision: addCollision, wake: wake})
	case "door-circle":
		doorCircleCollision5488B0(a, b, i.Move, doorCircleCollisionHooks5488B0{spring: func() float32 { return spring }, frame: func() uint32 { return i.Frame }, tickRate: func() uint32 { return i.TickRate }, direction: directionInt, addCollision: addCollision, wake: wake, queueDoor: queue, addToUpdatable: updatable, priorityMessage: message})
	case "door-box":
		doorBoxCollision551250(a, b, i.Move, doorBoxCollisionHooks551250{spring: func() float32 { return spring }, frame: func() uint32 { return i.Frame }, tickRate: func() uint32 { return i.TickRate }, directionFloat: directionFloat, directionInt: directionInt, addCollision: addCollision, wake: wake, queueDoor: queue, addToUpdatable: updatable, priorityMessage: message})
	}
	if !bytes.Equal(rawObj(a), expectedA) || !bytes.Equal(rawObj(b), expectedB) || !bytes.Equal(rawU(u), expectedU) || !bytes.Equal(rawU(alt), expectedAlt) || !reflect.DeepEqual(events, row.Events) || math.Float32bits(pushScale) != row.Push || math.Float32bits(spring) != row.Spring {
		firstDifference := -1
		for index := 0; index < len(events) && index < len(row.Events); index++ {
			if !reflect.DeepEqual(events[index], row.Events[index]) {
				firstDifference = index
				break
			}
		}
		t.Fatalf("native first=%+v second=%+v old=%+v alternate=%+v; original first=%+v second=%+v old=%+v alternate=%+v; event counts=%d/%d first difference=%d; native bytes preserved=%t/%t/%t/%t", state(a), state(b), updateState(u), updateState(alt), row.First, row.Second, row.OldUpdate, row.AlternateUpdate, len(events), len(row.Events), firstDifference, bytes.Equal(rawObj(a), expectedA), bytes.Equal(rawObj(b), expectedB), bytes.Equal(rawU(u), expectedU), bytes.Equal(rawU(alt), expectedAlt))
	}
}
