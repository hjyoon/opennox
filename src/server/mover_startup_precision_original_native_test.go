package server

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type moverPrecisionEventR85 struct {
	Service      string    `json:"service"`
	Object       string    `json:"object"`
	PositionBits [2]uint32 `json:"position_bits"`
	Index        uint32    `json:"index"`
	Low          int       `json:"low"`
	High         int       `json:"high"`
	Result       int       `json:"result"`
}

type moverPrecisionCaseR85 struct {
	Input struct {
		Name            string    `json:"name"`
		State           uint8     `json:"state"`
		Flags           uint32    `json:"flags"`
		TargetFlags     uint32    `json:"target_flags"`
		Extent          uint32    `json:"extent"`
		ResolveTarget   bool      `json:"resolve_target"`
		MissingTarget   bool      `json:"missing_target"`
		ResolveWaypoint bool      `json:"resolve_waypoint"`
		WaypointID      uint32    `json:"waypoint_id"`
		SpeedUnits      int32     `json:"speed_units"`
		SpeedBits       uint32    `json:"speed_bits"`
		PositionBits    [2]uint32 `json:"position_bits"`
		VelocityBits    [2]uint32 `json:"velocity_bits"`
		WaypointBits    [2]uint32 `json:"waypoint_bits"`
		Links           uint8     `json:"links"`
	} `json:"input"`
	State          uint8                    `json:"state"`
	Speed          uint32                   `json:"speed_bits"`
	SpeedBase      uint32                   `json:"speed_base_bits"`
	Velocity       [2]uint32                `json:"velocity_bits"`
	Current        string                   `json:"current_waypoint"`
	Previous       string                   `json:"previous_waypoint"`
	Target         string                   `json:"target"`
	SourcePosition [2]uint32                `json:"source_position_bits"`
	TargetPosition [2]uint32                `json:"target_position_bits"`
	Events         []moverPrecisionEventR85 `json:"events"`
	Preserved      bool                     `json:"all_nonoutput_bytes_preserved"`
}

func moverPrecisionPointR85(bits [2]uint32) types.Pointf {
	return types.Pointf{X: math.Float32frombits(bits[0]), Y: math.Float32frombits(bits[1])}
}

func moverPrecisionBitsR85(point types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(point.X), math.Float32bits(point.Y)}
}

// The original whole-function oracle declares lookup, position-only Move,
// RNG and RemoveUpdatable services. Arithmetic, branch decisions, PE32 IDs
// and native sidecar bindings execute together here on C-owned records.
func TestMoverStartupPrecisionOriginal54F740Native(t *testing.T) {
	data, err := os.ReadFile("testdata/mover_startup_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Original string                  `json:"original_sha256"`
		Cases    []moverPrecisionCaseR85 `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Cases) != 117 {
		t.Fatal("original mover fixture identity/count")
	}
	for _, want := range fixture.Cases {
		t.Run(want.Input.Name, func(t *testing.T) {
			if !want.Preserved {
				t.Fatal("uncertified original frame")
			}
			s := newMoverStateServer(t)
			source := playerCloudGreatSwordAllocateR78[Object](t)
			target := playerCloudGreatSwordAllocateR78[Object](t)
			update := playerCloudGreatSwordAllocateR78[MoverUpdateData](t)
			current := playerCloudGreatSwordAllocateR78[Waypoint](t)
			next := playerCloudGreatSwordAllocateR78[Waypoint](t)
			alternate := playerCloudGreatSwordAllocateR78[Waypoint](t)
			*source = Object{serverHandle: s.handle, ObjFlags: object.Flags(want.Input.Flags),
				PosVec: moverPrecisionPointR85(want.Input.PositionBits), VelVec: moverPrecisionPointR85(want.Input.VelocityBits),
				SpeedCur: math.Float32frombits(want.Input.SpeedBits), SpeedBase: 1, UpdateData: unsafe.Pointer(update)}
			*target = Object{serverHandle: s.handle, ObjFlags: object.Flags(want.Input.TargetFlags), PosVec: types.Pointf{X: 30, Y: 40}}
			*update = MoverUpdateData{Field_0: want.Input.State, Field_1: math.Float32frombits(uint32(want.Input.SpeedUnits)),
				Field_2: int32(want.Input.WaypointID), Field_8: want.Input.Extent}
			*current = Waypoint{Index: 7, PosVec: moverPrecisionPointR85(want.Input.WaypointBits), PointsCnt: want.Input.Links}
			*next = Waypoint{Index: 8, PosVec: types.Pointf{X: 13, Y: 24}}
			*alternate = Waypoint{Index: 9, PosVec: types.Pointf{X: -7, Y: 9}}
			current.Points[0].Waypoint, current.Points[1].Waypoint = next, alternate
			// Allocate the native extension before freezing the object frame.
			// This primes only its handle; target/waypoint resolution stays live.
			source.SetExt()
			if !want.Input.ResolveTarget {
				source.SetMoverTargetFor(update, target)
			}
			if want.Input.ResolveWaypoint {
				update.Field_4 = 7
			} else {
				source.SetMoverWaypointFor(update, 3, current)
			}
			if want.Input.Links == 2 {
				source.SetMoverWaypointFor(update, 5, next)
			}
			type frame struct {
				name string
				ptr  unsafe.Pointer
				size uintptr
			}
			frames := []frame{{"source", unsafe.Pointer(source), unsafe.Sizeof(*source)},
				{"target", unsafe.Pointer(target), unsafe.Sizeof(*target)}, {"update", unsafe.Pointer(update), unsafe.Sizeof(*update)},
				{"current", unsafe.Pointer(current), unsafe.Sizeof(*current)}, {"next", unsafe.Pointer(next), unsafe.Sizeof(*next)},
				{"alternate", unsafe.Pointer(alternate), unsafe.Sizeof(*alternate)}}
			before := make([][]byte, len(frames))
			for i, f := range frames {
				before[i] = append([]byte(nil), unsafe.Slice((*byte)(f.ptr), f.size)...)
			}
			t.Cleanup(func() {
				preserved := true
				for i, f := range frames {
					actual := unsafe.Slice((*byte)(f.ptr), f.size)
					expected := append([]byte(nil), before[i]...)
					copyOutput := func(offset, size uintptr) { copy(expected[offset:offset+size], actual[offset:offset+size]) }
					switch i {
					case 0:
						copyOutput(unsafe.Offsetof(source.PosVec), unsafe.Sizeof(source.PosVec))
						copyOutput(unsafe.Offsetof(source.VelVec), unsafe.Sizeof(source.VelVec))
						copyOutput(unsafe.Offsetof(source.SpeedCur), 4)
						copyOutput(unsafe.Offsetof(source.SpeedBase), 4)
					case 1:
						copyOutput(unsafe.Offsetof(target.PosVec), unsafe.Sizeof(target.PosVec))
					case 2:
						copyOutput(unsafe.Offsetof(update.Field_0), 1)
					}
					if !reflect.DeepEqual(actual, expected) {
						preserved = false
						t.Errorf("nonoutput %s frame changed", f.name)
					}
				}
				if preserved {
					t.Log("all_native_nonoutput_frames_preserved:true")
				}
			})
			var events []moverPrecisionEventR85
			moverUpdateNative54F740(source, moverUpdateNativeDeps54F740{
				objectByExtent: func(extent uint32) *Object {
					if extent != 123 {
						t.Fatal("object lookup argument")
					}
					events = append(events, moverPrecisionEventR85{Service: "object-lookup"})
					if want.Input.MissingTarget {
						return nil
					}
					return target
				},
				waypointByID: func(id uint32) *Waypoint {
					events = append(events, moverPrecisionEventR85{Service: "waypoint-lookup", Index: id})
					return map[uint32]*Waypoint{7: current, 8: next}[id]
				},
				randomInt: func(low, high int) int {
					if low != 0 || high != 1 {
						t.Fatal("random bounds")
					}
					events = append(events, moverPrecisionEventR85{Service: "random", Low: low, High: high, Result: 1})
					return 1
				},
				move: func(unit *Object, position types.Pointf) {
					name := "source"
					if unit == target {
						name = "target"
					} else if unit != source {
						t.Fatal("move native pointer identity")
					}
					events = append(events, moverPrecisionEventR85{Service: "move", Object: name, PositionBits: moverPrecisionBitsR85(position)})
					unit.PosVec = position
				},
				removeUpdatable: func(unit *Object) {
					if unit != source {
						t.Fatal("remove native pointer identity")
					}
					events = append(events, moverPrecisionEventR85{Service: "remove-updatable"})
				},
			})
			if len(events) != len(want.Events) || len(events) != 0 && !reflect.DeepEqual(events, want.Events) {
				t.Errorf("services = %+v, want %+v", events, want.Events)
			}
			if update.Field_0 != want.State || math.Float32bits(source.SpeedCur) != want.Speed || math.Float32bits(source.SpeedBase) != want.SpeedBase || moverPrecisionBitsR85(source.VelVec) != want.Velocity {
				t.Errorf("state/speed/base/velocity = %d/%08x/%08x/%08x, want %d/%08x/%08x/%08x", update.Field_0, math.Float32bits(source.SpeedCur), math.Float32bits(source.SpeedBase), moverPrecisionBitsR85(source.VelVec), want.State, want.Speed, want.SpeedBase, want.Velocity)
			}
			if moverPrecisionBitsR85(source.PosVec) != want.SourcePosition || moverPrecisionBitsR85(target.PosVec) != want.TargetPosition {
				t.Error("original service position outputs differ")
			}
			waypointName := map[*Waypoint]string{nil: "nil", current: "current", next: "next", alternate: "alternate"}
			if waypointName[source.MoverWaypointFor(update, 3)] != want.Current || waypointName[source.MoverWaypointFor(update, 5)] != want.Previous {
				t.Error("native waypoint sidecar outputs differ")
			}
			gotTarget := "nil"
			if got := source.MoverTargetFor(update); got == target {
				gotTarget = "target"
			} else if got != nil {
				t.Fatal("unexpected native target identity")
			}
			if gotTarget != want.Target {
				t.Errorf("target = %s, want %s", gotTarget, want.Target)
			}
		})
	}
}
