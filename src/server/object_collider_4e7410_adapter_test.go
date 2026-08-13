package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestObjectColliderAllowed4E7410NativeAdapter(t *testing.T) {
	noCollide := &Object{
		ObjFlags:  object.FlagNoCollide,
		CollideP1: types.Ptf(1, 2),
		CollideP2: types.Ptf(3, 4),
	}
	noCollide.Shape.Kind = 99
	if got := noCollide.Sub_4E7410(); got != 1 {
		t.Fatalf("NoCollide result = %d, want 1", got)
	}
	if noCollide.CollideP1 != types.Ptf(1, 2) || noCollide.CollideP2 != types.Ptf(3, 4) {
		t.Fatalf("NoCollide bounds changed to %v..%v", noCollide.CollideP1, noCollide.CollideP2)
	}

	current := &Object{PosVec: types.Ptf(10, 20), NewPos: types.Ptf(1000, 2000)}
	current.Shape.Kind = ShapeKindCircle
	current.Shape.Circle.R = 42
	if got := current.Sub_4E7410(); got != 1 {
		t.Fatalf("radius-42 result = %d, want 1", got)
	}
	if current.CollideP1 != types.Ptf(-32, -22) || current.CollideP2 != types.Ptf(52, 62) {
		t.Fatalf("bounds = %v..%v, want current-position bounds", current.CollideP1, current.CollideP2)
	}

	current.Shape.Circle.R = 42.5
	if got := current.Sub_4E7410(); got != 0 {
		t.Fatalf("radius-42.5 result = %d, want 0", got)
	}

	unordered := &Object{PosVec: types.Ptf(math.Float32frombits(0x7fa54321), 0)}
	unordered.Shape.Kind = ShapeKindCenter
	if got := unordered.Sub_4E7410(); got != 1 {
		t.Fatalf("unordered result = %d, want 1", got)
	}
}

func TestObjectColliderAllowed4E7410NativeNilFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not panic")
		}
	}()
	var obj *Object
	obj.Sub_4E7410()
}
