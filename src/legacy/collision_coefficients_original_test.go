package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func rigidNativePoint(w [2]uint32) types.Pointf {
	return types.Ptf(math.Float32frombits(w[0]), math.Float32frombits(w[1]))
}

func rigidNativeBits(p types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}
}

func rigidNativeAddress(t *testing.T, p unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
		t.Fatal("C-owned collision input below 4 GiB")
	}
}

func rigidNativeTable(t *testing.T, offset uintptr, data []byte) {
	t.Helper()
	table := memmap.Slice(0x587000, offset)[:len(data)]
	before := bytes.Clone(table)
	t.Cleanup(func() { copy(table, before) })
	copy(table, data)
}

func rigidNativeObjectBytes(p *server.Object) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
}

func rigidNativeExpected(p *server.Object, next, velocity, accumulator [2]uint32) []byte {
	want := bytes.Clone(rigidNativeObjectBytes(p))
	for _, field := range []struct {
		p *types.Pointf
		w [2]uint32
	}{{&p.NewPos, next}, {&p.VelVec, velocity}, {&p.Pos24, accumulator}} {
		offset := uintptr(unsafe.Pointer(field.p)) - uintptr(unsafe.Pointer(p))
		binary.LittleEndian.PutUint32(want[offset:], field.w[0])
		binary.LittleEndian.PutUint32(want[offset+4:], field.w[1])
	}
	return want
}

type rigidNativeState struct {
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

type rigidNativeUpdate struct {
	Lock       byte   `json:"lock"`
	Target     int32  `json:"target"`
	Current    int32  `json:"current"`
	Queued     uint32 `json:"queued"`
	Angular    uint32 `json:"angular_bits32"`
	Fractional int16  `json:"fractional"`
	LastMove   uint32 `json:"last_move_frame"`
}

type rigidNativeEvent struct {
	Service   string            `json:"service"`
	Object    string            `json:"object"`
	Other     string            `json:"other"`
	Normal    []uint32          `json:"normal_bits32"`
	First     rigidNativeState  `json:"first"`
	Second    rigidNativeState  `json:"second"`
	OldUpdate rigidNativeUpdate `json:"old_update"`
}

// Exercise the public native collision bodies with their actual C-owned
// coefficient providers and mapped direction tables. Queue/wakeup callbacks
// are observers, matching the declared services of the unchanged PE oracle.
func TestRigidCollisionNativeProvidersOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/rigid_collision_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Tables   []struct {
			Name string `json:"name"`
			Hex  string `json:"hex"`
		} `json:"tables"`
		Rows []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name              string    `json:"name"`
				Kind              string    `json:"kind"`
				FirstNew          [2]uint32 `json:"first_new_bits32"`
				SecondNew         [2]uint32 `json:"second_new_bits32"`
				FirstSize         [2]uint32 `json:"first_size_bits32"`
				SecondSize        [2]uint32 `json:"second_size_bits32"`
				FirstAccumulator  [2]uint32 `json:"first_accumulator_bits32"`
				SecondAccumulator [2]uint32 `json:"second_accumulator_bits32"`
				Velocity          [2]uint32 `json:"second_velocity_bits32"`
				FirstMass         uint32    `json:"first_mass_bits32"`
				SecondMass        uint32    `json:"second_mass_bits32"`
				Radius            uint32    `json:"second_radius_bits32"`
				Angular           uint32    `json:"angular_bits32"`
				Push              uint32    `json:"push_scale_bits32"`
				Spring            uint32    `json:"spring_bits32"`
				Move              int32     `json:"move"`
				Frame             uint32    `json:"frame"`
			} `json:"input"`
			First     rigidNativeState   `json:"first"`
			Second    rigidNativeState   `json:"second"`
			Update    rigidNativeUpdate  `json:"old_update"`
			Events    []rigidNativeEvent `json:"events"`
			Preserved bool               `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 1263 {
		t.Fatal("incomplete original fixture")
	}
	for _, table := range fixture.Tables {
		var offset uintptr
		switch table.Name {
		case "integer-directions":
			offset = 192088
		case "float-directions":
			offset = 194136
		default:
			continue
		}
		data, err := hex.DecodeString(table.Hex)
		if err != nil || len(data) != 2048 {
			t.Fatal("invalid sealed direction table")
		}
		rigidNativeTable(t, offset, data)
	}
	selected := map[string]bool{
		"box-box/origin-stock-coefficient":      false,
		"box-box/origin-live-coefficient-0":     false,
		"box-box/origin-live-coefficient-37.5":  false,
		"box-box/origin-live-coefficient--3.25": false,
		"door-circle/native-live-spring":        false,
		"door-box/native-live-spring":           false,
	}
	for _, row := range fixture.Rows {
		if _, ok := selected[row.Input.Name]; !ok || row.Mode != "actual-startup" {
			continue
		}
		selected[row.Input.Name] = true
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			a, freeA := alloc.New(server.Object{})
			b, freeB := alloc.New(server.Object{})
			u, freeU := alloc.New(server.DoorUpdateData{})
			for _, free := range []func(){freeA, freeB, freeU} {
				t.Cleanup(free)
			}
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(u)} {
				rigidNativeAddress(t, ptr)
			}
			*a = server.Object{NewPos: rigidNativePoint(i.FirstNew), Pos24: rigidNativePoint(i.FirstAccumulator), Mass: math.Float32frombits(i.FirstMass), UpdateData: unsafe.Pointer(u), Field29: 0x13572468, Field33: 77, PosVec: types.Ptf(10, 20)}
			*b = server.Object{NewPos: rigidNativePoint(i.SecondNew), Pos24: rigidNativePoint(i.SecondAccumulator), VelVec: rigidNativePoint(i.Velocity), Mass: math.Float32frombits(i.SecondMass), Field29: 0x87654321, Field33: 99, PosVec: types.Ptf(30, 40)}
			a.Shape.Kind, b.Shape.Kind = server.ShapeKindBox, server.ShapeKindBox
			a.Shape.Box.W, a.Shape.Box.H = math.Float32frombits(i.FirstSize[0]), math.Float32frombits(i.FirstSize[1])
			b.Shape.Box.W, b.Shape.Box.H = math.Float32frombits(i.SecondSize[0]), math.Float32frombits(i.SecondSize[1])
			a.Shape.Circle.R, b.Shape.Circle.R = 4, math.Float32frombits(i.Radius)
			if i.Kind == "door-circle" {
				b.Shape.Kind = server.ShapeKindCircle
			}
			rawU := unsafe.Slice((*byte)(unsafe.Pointer(u)), int(unsafe.Sizeof(*u)))
			for index := range rawU {
				rawU[index] = 0xa5
			}
			u.LockCode, u.TargetDirection, u.CurrentDirection, u.Queued, u.FractionalDir = 0, 0, 0, 0, 0
			u.AngularVelocity, u.LastMoveFrame = math.Float32frombits(i.Angular), 7
			wantA := rigidNativeExpected(a, row.First.New, row.First.Velocity, row.First.Accumulator)
			wantB := rigidNativeExpected(b, row.Second.New, row.Second.Velocity, row.Second.Accumulator)
			wantU := bytes.Clone(rawU)
			binary.LittleEndian.PutUint32(wantU[unsafe.Offsetof(u.LastMoveFrame):], row.Update.LastMove)
			label := func(p *server.Object) string {
				switch p {
				case a:
					return "first"
				case b:
					return "second"
				case nil:
					return "nil"
				}
				t.Fatal("unknown native object")
				return ""
			}
			state := func(p *server.Object) rigidNativeState {
				update := "nil"
				if p.UpdateData == unsafe.Pointer(u) {
					update = "old"
				} else if p.UpdateData != nil {
					t.Fatal("changed update pointer")
				}
				return rigidNativeState{New: rigidNativeBits(p.NewPos), Velocity: rigidNativeBits(p.VelVec), Accumulator: rigidNativeBits(p.Pos24), Class: uint32(p.ObjClass), Flags: uint32(p.ObjFlags), Mass: math.Float32bits(p.Mass), Radius: math.Float32bits(p.Shape.Circle.R), Size: [2]uint32{math.Float32bits(p.Shape.Box.W), math.Float32bits(p.Shape.Box.H)}, Field34: p.Field34, Team: byte(p.TeamVal.ID), Owner: label(p.ObjOwner), Update: update}
			}
			update := func() rigidNativeUpdate {
				return rigidNativeUpdate{u.LockCode, u.TargetDirection, u.CurrentDirection, u.Queued, math.Float32bits(u.AngularVelocity), u.FractionalDir, u.LastMoveFrame}
			}
			events := make([]rigidNativeEvent, 0)
			observe := func(e rigidNativeEvent) {
				e.First, e.Second, e.OldUpdate = state(a), state(b), update()
				events = append(events, e)
			}
			collision := func(first, second *server.Object, normal types.Pointf) {
				w := rigidNativeBits(normal)
				observe(rigidNativeEvent{Service: "collision", Object: label(first), Other: label(second), Normal: []uint32{w[0], w[1]}})
			}
			wake := func(p *server.Object) { observe(rigidNativeEvent{Service: "wake", Object: label(p)}) }
			queue := func(*server.DoorUpdateData) { t.Fatal("unexpected door queue in linear-force case") }
			push, spring := collisionPushScalePtr550F80(), collisionSpringPtr5488B0()
			oldPush, oldSpring := *push, *spring
			t.Cleanup(func() { *push = oldPush; *spring = oldSpring })
			*push, *spring = math.Float32frombits(i.Push), math.Float32frombits(i.Spring)
			s := &server.Server{}
			s.SetFrame(i.Frame)
			switch i.Kind {
			case "box-box":
				s.BoxBoxCollision550F80(a, b, collisionPushScaleNative550F80, collision, wake)
			case "door-circle":
				s.DoorCircleCollision5488B0(a, b, i.Move, collisionSpringNative5488B0, collision, wake, queue)
			case "door-box":
				s.DoorBoxCollision551250(a, b, i.Move, collisionSpringNative5488B0, collision, wake, queue)
			}
			if !row.Preserved || state(a) != row.First || state(b) != row.Second || update() != row.Update || !bytes.Equal(rigidNativeObjectBytes(a), wantA) || !bytes.Equal(rigidNativeObjectBytes(b), wantB) || !bytes.Equal(rawU, wantU) || !reflect.DeepEqual(events, row.Events) {
				t.Fatalf("native public body with live C coefficients differs: first=%+v second=%+v update=%+v events=%+v original=%+v", state(a), state(b), update(), events, row.Events)
			}
		})
	}
	for name, found := range selected {
		if !found {
			t.Error("missing original condition", name)
		}
	}
}

func TestRigidCollisionNativeBoxWallSpringOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/box_wall_coefficients_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type event struct {
		Service     string    `json:"service"`
		Grid        []int32   `json:"grid"`
		Impulse     []uint32  `json:"impulse_bits32"`
		New         [2]uint32 `json:"new_bits32"`
		Velocity    [2]uint32 `json:"velocity_bits32"`
		Accumulator [2]uint32 `json:"accumulator_bits32"`
	}
	var fixture struct {
		Contexts    int    `json:"contexts"`
		Definitions string `json:"original_wall_definitions_hex"`
		Rows        []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name        string    `json:"name"`
				Class       uint32    `json:"class_bits"`
				Collide     [4]uint32 `json:"collide_bits32"`
				New         [2]uint32 `json:"new_bits32"`
				Previous    [2]uint32 `json:"previous_bits32"`
				Velocity    [2]uint32 `json:"velocity_bits32"`
				Size        [2]uint32 `json:"size_bits32"`
				Accumulator [2]uint32 `json:"accumulator_bits32"`
				Mass        uint32    `json:"mass_bits32"`
				WallType    byte      `json:"wall_type"`
				Grid        [2]int    `json:"wall_grid"`
				Spring      uint32    `json:"spring_bits32"`
			} `json:"input"`
			Events      []event   `json:"events"`
			New         [2]uint32 `json:"new_bits32"`
			Velocity    [2]uint32 `json:"velocity_bits32"`
			Accumulator [2]uint32 `json:"accumulator_bits32"`
			Preserved   bool      `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	defs, err := hex.DecodeString(fixture.Definitions)
	if err != nil || fixture.Contexts != 8 || len(fixture.Rows) != 16 || len(defs) != 264 {
		t.Fatal("incomplete original live-spring fixture")
	}
	const offset = 292520
	rigidNativeTable(t, offset, defs)
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			obj, free := alloc.New(server.Object{})
			t.Cleanup(free)
			rigidNativeAddress(t, unsafe.Pointer(obj))
			*obj = server.Object{ObjClass: object.Class(i.Class), NewPos: rigidNativePoint(i.New), PrevPos: rigidNativePoint(i.Previous), VelVec: rigidNativePoint(i.Velocity), Pos24: rigidNativePoint(i.Accumulator), Mass: math.Float32frombits(i.Mass), CollideP1: rigidNativePoint([2]uint32{i.Collide[0], i.Collide[1]}), CollideP2: rigidNativePoint([2]uint32{i.Collide[2], i.Collide[3]}), Field29: 0x13572468}
			obj.Shape.Kind = server.ShapeKindBox
			obj.Shape.Box.W, obj.Shape.Box.H = math.Float32frombits(i.Size[0]), math.Float32frombits(i.Size[1])
			want := rigidNativeExpected(obj, row.New, row.Velocity, row.Accumulator)
			s := &server.Server{}
			if s.Walls.Init() != 1 {
				t.Fatal("native wall allocation failed")
			}
			t.Cleanup(s.Walls.Free)
			grid := image.Pt(i.Grid[0], i.Grid[1])
			wall := s.Walls.CreateAtGrid(grid)
			if wall == nil {
				t.Fatal("native wall insertion failed")
			}
			rigidNativeAddress(t, unsafe.Pointer(wall))
			wall.Dir0 = i.WallType
			if s.Sub_57B500(grid, 0x40) != int8(i.WallType) {
				t.Fatal("real indexed wall lookup failed")
			}
			events := make([]event, 0)
			state := func(service string) event {
				return event{Service: service, New: rigidNativeBits(obj.NewPos), Velocity: rigidNativeBits(obj.VelVec), Accumulator: rigidNativeBits(obj.Pos24)}
			}
			add := func(got *server.Object, impulse types.Pointf) {
				if got != obj {
					t.Fatal("collision object changed")
				}
				e := state("collision")
				w := rigidNativeBits(impulse)
				e.Impulse = []uint32{w[0], w[1]}
				events = append(events, e)
			}
			secret := func(grid image.Point, got *server.Object) {
				if got != obj {
					t.Fatal("secret contact object changed")
				}
				e := state("secret")
				e.Grid = []int32{int32(grid.X), int32(grid.Y)}
				events = append(events, e)
			}
			spring := collisionSpringPtr5488B0()
			old := *spring
			t.Cleanup(func() { *spring = old })
			*spring = math.Float32frombits(i.Spring)
			s.BoxWallCollision5504B0(obj, collisionSpringNative5488B0, add, secret)
			wantEvents := make([]event, 0)
			for _, e := range row.Events {
				if e.Service != "wall" {
					wantEvents = append(wantEvents, e)
				}
			}
			if !row.Preserved || !bytes.Equal(rigidNativeObjectBytes(obj), want) || !reflect.DeepEqual(events, wantEvents) || !bytes.Equal(memmap.Slice(0x587000, offset)[:len(defs)], defs) || math.Float32bits(*spring) != i.Spring {
				t.Fatalf("native indexed box/wall with live C spring differs: events=%+v original=%+v bytes=%t", events, wantEvents, bytes.Equal(rigidNativeObjectBytes(obj), want))
			}
		})
	}
	if selected != 8 {
		t.Fatalf("original live-spring conditions=%d", selected)
	}
}
