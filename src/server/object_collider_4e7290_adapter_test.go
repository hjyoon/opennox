package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestObjectUpdateCollider4E7290NativeAdapter(t *testing.T) {
	center := &Object{PosVec: types.Ptf(math.Float32frombits(0x7fa12345), math.Float32frombits(0x80000000))}
	center.Shape.Kind = ShapeKindCenter
	if got := center.Nox_xxx_objectUnkUpdateCoords_4E7290(); got != center {
		t.Fatalf("return = %p, want %p", got, center)
	}
	if got, want := [4]uint32{
		math.Float32bits(center.CollideP1.X), math.Float32bits(center.CollideP1.Y),
		math.Float32bits(center.CollideP2.X), math.Float32bits(center.CollideP2.Y),
	}, [4]uint32{0x7fa12345, 0x80000000, 0x7fa12345, 0x80000000}; got != want {
		t.Fatalf("center bounds = %#v, want %#v", got, want)
	}

	circle := &Object{PosVec: types.Ptf(12.5, -3.25)}
	circle.Shape.Kind = ShapeKindCircle
	circle.Shape.Circle.R = 2.5
	circle.Nox_xxx_objectUnkUpdateCoords_4E7290()
	if got, want := [4]float32{
		circle.CollideP1.X, circle.CollideP1.Y,
		circle.CollideP2.X, circle.CollideP2.Y,
	}, [4]float32{10, -5.75, 15, -0.75}; got != want {
		t.Fatalf("circle bounds = %v, want %v", got, want)
	}

	box := &Object{PosVec: types.Ptf(10, 20)}
	box.Shape.Kind = ShapeKindBox
	box.Shape.Box.LeftBottom2 = -4
	box.Shape.Box.LeftBottom = -7
	box.Shape.Box.RightTop = 6
	box.Shape.Box.RightTop2 = 9
	box.Nox_xxx_objectUnkUpdateCoords_4E7290()
	if got, want := [4]float32{
		box.CollideP1.X, box.CollideP1.Y,
		box.CollideP2.X, box.CollideP2.Y,
	}, [4]float32{6, 13, 16, 29}; got != want {
		t.Fatalf("box bounds = %v, want %v", got, want)
	}

	unknown := &Object{CollideP1: types.Ptf(1, 2), CollideP2: types.Ptf(3, 4)}
	unknown.Shape.Kind = 99
	unknown.Nox_xxx_objectUnkUpdateCoords_4E7290()
	if unknown.CollideP1 != types.Ptf(1, 2) || unknown.CollideP2 != types.Ptf(3, 4) {
		t.Fatalf("unknown shape changed bounds to %v..%v", unknown.CollideP1, unknown.CollideP2)
	}
}

func TestObjectUpdateCollider4E7290NativeNilFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not panic")
		}
	}()
	var obj *Object
	obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
}
