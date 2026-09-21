package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func newAIPathSelectionServer50AC20(t *testing.T) (*Server, *serverAIPaths) {
	t.Helper()
	s := newElevatorStateServer(t)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	return s, &serverAIPaths{s: s}
}

func addAIPathSelectionObject50AC20(s *Server, class object.Class, pos types.Pointf) *Object {
	obj := &Object{
		ObjClass:     class,
		ObjFlags:     object.FlagActive | object.FlagEnabled,
		PosVec:       pos,
		NewPos:       pos,
		serverHandle: s.handle,
	}
	obj.Shape.Kind = ShapeKindCircle
	obj.Shape.Circle.R = 1
	s.Map.AddObjectToIndex(obj)
	return obj
}

func TestAIPathHoleDestinationSentinel50AC20UsesHostPointerWidth(t *testing.T) {
	offset := unsafe.Offsetof(HoleCollideData{}.DestinationX)
	if offset != 8 {
		t.Fatalf("DestinationX offset = %d, want 8", offset)
	}
	sentinel := ^uintptr(offset - 1)
	if aiPathHoleDestinationValid50AC20(sentinel) {
		t.Fatalf("wrapped sentinel %#x was accepted", sentinel)
	}
	for _, addr := range []uintptr{0, 1, sentinel - 1, sentinel + 1} {
		if !aiPathHoleDestinationValid50AC20(addr) {
			t.Fatalf("ordinary address %#x was rejected", addr)
		}
	}
}

func TestAIPathTargetCell50AC20UsesExactGridInverseAndX87Rounding(t *testing.T) {
	if got := math.Float32bits(aiPathGridInverse50AC20); got != 0x3d321643 {
		t.Fatalf("grid inverse bits = %#08x, want 0x3d321643", got)
	}
	tests := []struct {
		value float32
		want  uint16
	}{
		{0, 0},
		{11.5, 0},
		{34.5, 2},
		{57.5, 2},
		{80.5, 4},
		{-34.5, uint16(0xfffe)},
		{float32(math.Inf(1)), 0},
		{float32(math.NaN()), 0},
	}
	for _, test := range tests {
		if got := aiPathTargetCell50AC20(test.value); got != test.want {
			t.Errorf("cell(%v) = %#x, want %#x", test.value, got, test.want)
		}
	}
}

func TestAIPathDestination50AC20AllObjectClasses(t *testing.T) {
	tests := []struct {
		name  string
		class object.Class
		flag  AIMapIndexFlags
		link  func(source, target *Object)
	}{
		{
			name:  "transporter",
			class: object.ClassTransporter,
			flag:  AIIndexTransporter,
			link: func(source, target *Object) {
				source.UpdateData = unsafe.Pointer(&TransporterUpdateData{})
				source.SetTransporterTarget(target)
			},
		},
		{
			name:  "elevator",
			class: object.ClassElevator,
			flag:  AIIndexElevator,
			link: func(source, target *Object) {
				source.UpdateData = unsafe.Pointer(&ElevatorUpdateData{})
				source.SetElevatorLink(target)
			},
		},
		{
			name:  "elevator shaft",
			class: object.ClassElevatorShaft,
			flag:  AIIndexElevatorShaft,
			link: func(source, target *Object) {
				source.UpdateData = unsafe.Pointer(&ElevatorShaftUpdateData{})
				source.SetElevatorLink(target)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, paths := newAIPathSelectionServer50AC20(t)
			const x, y = 4, 5
			source := addAIPathSelectionObject50AC20(s, test.class, types.Ptf(x*23, y*23))
			target := &Object{
				PosVec:       types.Ptf(34.5, 80.5),
				serverHandle: s.handle,
			}
			test.link(source, target)
			paths.MapIndex(x, y).Flags8 = test.flag

			out := [2]uint16{0xaaaa, 0xbbbb}
			if got := paths.Sub_50AC20(&AIVisitNode{X0: x, Y2: y}, &out); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if want := [2]uint16{2, 4}; out != want {
				t.Fatalf("destination = %v, want %v", out, want)
			}

			source.ObjFlags &^= object.FlagEnabled
			out = [2]uint16{0xaaaa, 0xbbbb}
			if got := paths.Sub_50AC20(&AIVisitNode{X0: x, Y2: y}, &out); got != 0 {
				t.Fatalf("disabled result = %d, want 0", got)
			}
			if want := [2]uint16{0xaaaa, 0xbbbb}; out != want {
				t.Fatalf("disabled destination = %v, want unchanged %v", out, want)
			}
		})
	}
}

func TestAIPathHoleDestination50AC20UsesSignedDivisionAndFlagPriority(t *testing.T) {
	s, paths := newAIPathSelectionServer50AC20(t)
	const x, y = 6, 7
	pos := types.Ptf(x*23, y*23)
	transporter := addAIPathSelectionObject50AC20(s, object.ClassTransporter, pos)
	transporter.UpdateData = unsafe.Pointer(&TransporterUpdateData{})
	transporter.SetTransporterTarget(&Object{PosVec: types.Ptf(230, 460), serverHandle: s.handle})
	paths.MapIndex(x, y).Flags8 = AIIndexHole | AIIndexTransporter

	out := [2]uint16{0xaaaa, 0xbbbb}
	if got := paths.Sub_50AC20(&AIVisitNode{X0: x, Y2: y}, &out); got != 0 {
		t.Fatalf("missing-priority-hole result = %d, want 0", got)
	}
	if want := [2]uint16{0xaaaa, 0xbbbb}; out != want {
		t.Fatalf("missing-priority-hole destination = %v, want unchanged %v", out, want)
	}

	hole := addAIPathSelectionObject50AC20(s, object.ClassHole, pos)
	hole.CollideData = unsafe.Pointer(&HoleCollideData{
		DestinationX: -68,
		DestinationY: 68,
	})
	if got := paths.Sub_50AC20(&AIVisitNode{X0: x, Y2: y}, &out); got != 1 {
		t.Fatalf("hole result = %d, want 1", got)
	}
	if want := [2]uint16{uint16(0xfffe), 2}; out != want {
		t.Fatalf("hole destination = %v, want %v", out, want)
	}
}

func TestAIPathFlagPriority50AC20(t *testing.T) {
	s, paths := newAIPathSelectionServer50AC20(t)
	const x, y = 10, 11
	pos := types.Ptf(x*23, y*23)

	hole := addAIPathSelectionObject50AC20(s, object.ClassHole, pos)
	hole.CollideData = unsafe.Pointer(&HoleCollideData{DestinationX: 23, DestinationY: 46})
	transporter := addAIPathSelectionObject50AC20(s, object.ClassTransporter, pos)
	transporter.UpdateData = unsafe.Pointer(&TransporterUpdateData{})
	transporter.SetTransporterTarget(&Object{PosVec: types.Ptf(69, 92), serverHandle: s.handle})
	elevator := addAIPathSelectionObject50AC20(s, object.ClassElevator, pos)
	elevator.UpdateData = unsafe.Pointer(&ElevatorUpdateData{})
	elevator.SetElevatorLink(&Object{PosVec: types.Ptf(115, 138), serverHandle: s.handle})
	shaft := addAIPathSelectionObject50AC20(s, object.ClassElevatorShaft, pos)
	shaft.UpdateData = unsafe.Pointer(&ElevatorShaftUpdateData{})
	shaft.SetElevatorLink(&Object{PosVec: types.Ptf(161, 184), serverHandle: s.handle})

	tests := []struct {
		name  string
		flags AIMapIndexFlags
		want  [2]uint16
	}{
		{"hole", AIIndexHole | AIIndexTransporter | AIIndexElevator | AIIndexElevatorShaft, [2]uint16{1, 2}},
		{"transporter", AIIndexTransporter | AIIndexElevator | AIIndexElevatorShaft, [2]uint16{3, 4}},
		{"elevator", AIIndexElevator | AIIndexElevatorShaft, [2]uint16{5, 6}},
		{"elevator shaft", AIIndexElevatorShaft, [2]uint16{7, 8}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paths.MapIndex(x, y).Flags8 = test.flags
			out := [2]uint16{0xaaaa, 0xbbbb}
			if got := paths.Sub_50AC20(&AIVisitNode{X0: x, Y2: y}, &out); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if out != test.want {
				t.Fatalf("destination = %v, want %v", out, test.want)
			}
		})
	}
}

func TestAIPathObjectPosition50AEA0KeepsCallbackSelectionAndOutputs(t *testing.T) {
	s, paths := newAIPathSelectionServer50AC20(t)
	const x, y = 8, 9
	first := addAIPathSelectionObject50AC20(s, object.ClassHole, types.Ptf(x*23, y*23))
	first.NetCode = 0x11112222
	second := addAIPathSelectionObject50AC20(s, object.ClassHole, types.Ptf(x*23+1, y*23+1))
	second.NetCode = 0x33334444
	paths.MapIndex(x, y).Flags8 = AIIndexHole

	var (
		outPos  types.Pointf
		netCode uint32
	)
	if !paths.sub50AEA0(&AIVisitNode{X0: x, Y2: y}, &outPos, &netCode) {
		t.Fatal("enabled matching object was rejected")
	}
	// Map insertion prepends, while 0050AE80 overwrites the retained object for
	// every match. The earlier inserted object is therefore the final result.
	if outPos != first.PosVec || netCode != first.NetCode {
		t.Fatalf("selected position/netcode = %v/%#x, want %v/%#x",
			outPos, netCode, first.PosVec, first.NetCode)
	}

	first.ObjFlags &^= object.FlagEnabled
	outPos = types.Ptf(-1, -1)
	netCode = math.MaxUint32
	if paths.sub50AEA0(&AIVisitNode{X0: x, Y2: y}, &outPos, &netCode) {
		t.Fatal("disabled final matching object was accepted")
	}
	if outPos != types.Ptf(-1, -1) || netCode != math.MaxUint32 {
		t.Fatalf("failure outputs changed to %v/%#x", outPos, netCode)
	}
}
