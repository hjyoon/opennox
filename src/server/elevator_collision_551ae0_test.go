package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestElevatorCollision551AE0PlacesPlayerOnPlatform(t *testing.T) {
	update := &ElevatorUpdateData{Field_4: 24}
	elevator := &Object{
		NewPos:     types.Ptf(100, 100),
		UpdateData: unsafe.Pointer(update),
	}
	elevator.Shape = mapPointInBoxTestShape57B850(40, 40)
	player := &Object{
		ObjClass: object.ClassPlayer,
		ObjFlags: object.FlagInHole,
		NewPos:   types.Ptf(100, 100),
		ZVal:     24,
		Field27:  9,
	}
	player.Shape.Kind = ShapeKindCircle
	player.Shape.Circle.R = 5

	elevatorCollision551AE0(elevator, player, true, elevatorCollisionHooks551AE0{
		typeIndex: func(string) int { return -1 },
		circleBox: func(*Object, *Object, bool) { t.Fatal("unexpected side collision") },
		boxBox:    func(*Object, *Object) { t.Fatal("unexpected side collision") },
	})

	if !player.ObjFlags.Has(object.FlagOnObject) || player.ObjFlags.Has(object.FlagInHole) {
		t.Fatalf("player support flags = %#x, want ON_OBJECT without IN_HOLE", player.ObjFlags)
	}
	if player.ZVal != 28 || player.Field27 != 0 {
		t.Fatalf("player height/vertical velocity = %g/%g, want 28/0", player.ZVal, player.Field27)
	}
	if player.Field38 != math.MaxUint32 {
		t.Fatalf("player sync mask = %#x, want MaxUint32", player.Field38)
	}
}

func TestElevatorCarryUp53B750MovesPlayerToShaft(t *testing.T) {
	update := &ElevatorUpdateData{Field_4: 40}
	elevator := &Object{
		PosVec:     types.Ptf(20, 30),
		UpdateData: unsafe.Pointer(update),
	}
	elevator.Shape = mapPointInBoxTestShape57B850(40, 40)
	shaft := &Object{PosVec: types.Ptf(120, 130)}
	shaft.Shape = mapPointInBoxTestShape57B850(40, 40)
	player := &Object{
		ObjClass: object.ClassPlayer,
		PosVec:   elevator.PosVec,
		ZVal:     40,
	}
	player.Shape.Kind = ShapeKindCircle
	player.Shape.Circle.R = 5

	elevatorCarryUp53B750(elevator, update, elevatorUpdateHooks53B5D0{
		link: func(got *Object) *Object {
			if got != elevator {
				t.Fatalf("link source = %p, want elevator %p", got, elevator)
			}
			return shaft
		},
		eachInCircle: func(center types.Pointf, radius float32, fn func(*Object) bool) {
			if center != elevator.PosVec || radius != 64 {
				t.Fatalf("carry search = %v/%g, want %v/64", center, radius, elevator.PosVec)
			}
			fn(player)
		},
		pointInBox: func(*types.Pointf, *Shape, *types.Pointf) bool { return true },
		move:       func(unit *Object, pos types.Pointf) { unit.PosVec = pos },
		raise:      func(unit *Object, z float32) { unit.Raise(z) },
	})

	if player.PosVec != shaft.PosVec || player.ZVal != -24 {
		t.Fatalf("carried player position/height = %v/%g, want %v/-24", player.PosVec, player.ZVal, shaft.PosVec)
	}
}
