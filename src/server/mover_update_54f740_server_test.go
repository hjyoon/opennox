package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMoverUpdateNativeState0KeepsTransientPointersOutOfPE32Record54F740(t *testing.T) {
	s := newMoverStateServer(t)
	units := int32(-7)
	data := &MoverUpdateData{
		Field_0: 0,
		Field_1: math.Float32frombits(uint32(units)),
		Field_2: 0x101,
		Field_3: 0xfefefefe,
		Field_4: 0x102,
		Field_5: 0xedededed,
		Field_6: 0x103,
		Field_7: 0xdcdcdcdc,
		Field_8: 0xabcdef81,
	}
	source := &Object{
		ObjFlags:     object.Flags(moverUpdateActiveFlag54F740),
		UpdateData:   unsafe.Pointer(data),
		serverHandle: s.handle,
	}
	target := &Object{
		ObjFlags:     object.Flags(moverUpdateTargetRequiredFlag54F740),
		PosVec:       types.Pointf{X: 31.5, Y: -47.25},
		serverHandle: s.handle,
	}
	start := &Waypoint{Index: 0x101}
	preloaded3 := &Waypoint{Index: 0x102}
	preloaded5 := &Waypoint{Index: 0x103}
	waypoints := map[uint32]*Waypoint{
		0x101: start,
		0x102: preloaded3,
		0x103: preloaded5,
	}
	var moves int
	var movedObject *Object
	var movedPosition types.Pointf
	var removed bool

	moverUpdateNative54F740(source, moverUpdateNativeDeps54F740{
		objectByExtent: func(extent uint32) *Object {
			if extent != data.Field_8 {
				t.Fatalf("target extent = %#x, want %#x", extent, data.Field_8)
			}
			return target
		},
		waypointByID: func(id uint32) *Waypoint { return waypoints[id] },
		randomInt: func(_, _ int) int {
			t.Fatal("state 0 must not request random selection")
			return 0
		},
		move: func(obj *Object, position types.Pointf) {
			moves++
			movedObject, movedPosition = obj, position
		},
		removeUpdatable: func(*Object) { removed = true },
	})

	if removed || moves != 1 || movedObject != source || movedPosition != target.PosVec {
		t.Fatalf("remove/move = %v/%d/%p/%+v, want false/1/%p/%+v",
			removed, moves, movedObject, movedPosition, source, target.PosVec)
	}
	if data.Field_0 != 1 || source.SpeedBase != -1.75 || source.SpeedCur != -1.75 {
		t.Fatalf("state/speed = %d/%g/%g, want 1/-1.75/-1.75",
			data.Field_0, source.SpeedBase, source.SpeedCur)
	}
	if got := source.MoverTargetFor(data); got != target {
		t.Fatalf("native target = %p, want %p", got, target)
	}
	if got := source.MoverWaypointFor(data, 3); got != start {
		t.Fatalf("native waypoint 3 = %p, want %p", got, start)
	}
	if got := source.MoverWaypointFor(data, 5); got != nil {
		t.Fatalf("native waypoint 5 = %p, want nil", got)
	}
	if data.Field_3 != 0 || data.Field_5 != 0 || data.Field_7 != 0 {
		t.Fatalf("PE32 pointer slots = %#x/%#x/%#x, want zero",
			data.Field_3, data.Field_5, data.Field_7)
	}
	if data.Field_2 != 0x101 || data.Field_4 != 0x102 || data.Field_6 != 0x103 || data.Field_8 != 0xabcdef81 {
		t.Fatalf("fixed IDs changed: %+v", *data)
	}
}

func TestMoverUpdateNativeState1UsesNativeWaypointGraph54F740(t *testing.T) {
	s := newMoverStateServer(t)
	data := &MoverUpdateData{Field_0: 1, Field_8: 44}
	source := &Object{
		ObjFlags:     object.Flags(moverUpdateActiveFlag54F740),
		PosVec:       types.Pointf{X: 10, Y: 20},
		VelVec:       types.Pointf{X: 1, Y: 1},
		SpeedCur:     5,
		UpdateData:   unsafe.Pointer(data),
		serverHandle: s.handle,
	}
	target := &Object{ObjFlags: object.Flags(moverUpdateTargetRequiredFlag54F740), serverHandle: s.handle}
	previous := &Waypoint{Index: 1}
	next := &Waypoint{Index: 2, PosVec: types.Pointf{X: 13, Y: 24}}
	current := &Waypoint{
		Index:     3,
		PosVec:    types.Pointf{X: 9, Y: 19},
		PointsCnt: 2,
	}
	current.Points[0].Waypoint = previous
	current.Points[1].Waypoint = next
	source.SetMoverTargetFor(data, target)
	source.SetMoverWaypointFor(data, 3, current)
	source.SetMoverWaypointFor(data, 5, previous)
	var movedObject *Object
	var movedPosition types.Pointf

	moverUpdateNative54F740(source, moverUpdateNativeDeps54F740{
		objectByExtent: func(uint32) *Object {
			t.Fatal("cached native target must avoid extent lookup")
			return nil
		},
		waypointByID: func(uint32) *Waypoint {
			t.Fatal("zero waypoint IDs must avoid lookup")
			return nil
		},
		randomInt: func(minimum, maximum int) int {
			if minimum != 0 || maximum != 1 {
				t.Fatalf("random bounds = %d..%d, want 0..1", minimum, maximum)
			}
			return 1
		},
		move: func(obj *Object, position types.Pointf) {
			movedObject, movedPosition = obj, position
		},
		removeUpdatable: func(*Object) { t.Fatal("state 1 must stay updatable") },
	})

	if got := source.MoverWaypointFor(data, 3); got != next {
		t.Fatalf("selected native waypoint = %p, want %p", got, next)
	}
	if got := source.MoverWaypointFor(data, 5); got != current {
		t.Fatalf("previous native waypoint = %p, want %p", got, current)
	}
	if movedObject != target || movedPosition != source.PosVec {
		t.Fatalf("target move = %p/%+v, want %p/%+v", movedObject, movedPosition, target, source.PosVec)
	}
	if source.VelVec.X == 0 || source.VelVec.Y == 0 {
		t.Fatalf("steering velocity = %+v, want both components", source.VelVec)
	}
	if data.Field_3 != 0 || data.Field_5 != 0 || data.Field_7 != 0 {
		t.Fatalf("PE32 pointer slots = %#x/%#x/%#x, want zero",
			data.Field_3, data.Field_5, data.Field_7)
	}
}
