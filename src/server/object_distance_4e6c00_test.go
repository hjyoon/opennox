package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestObjectDistanceNative4E6C00ShapeAndUnorderedContract(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		kind                  ShapeKind
		radius, width, height float32
		want                  float64
	}{
		{"center", ShapeKindCenter, 0, 0, 0, 10},
		{"circle", ShapeKindCircle, 3, 0, 0, 7},
		{"box-width", ShapeKindBox, 0, 6, 4, 7},
		{"box-height", ShapeKindBox, 0, 2, 8, 6},
		{"ignored-shape", ShapeKind(99), 100, 100, 100, 10},
		{"width-NaN-selects-height", ShapeKindBox, 0, float32(math.NaN()), 4, 8},
		{"height-NaN-clamps", ShapeKindBox, 0, 4, float32(math.NaN()), float64(float32(0.01))},
		{"circle-NaN-clamps", ShapeKindCircle, float32(math.NaN()), 0, 0, float64(float32(0.01))},
		{"overlap-clamps", ShapeKindCircle, 20, 0, 0, float64(float32(0.01))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := Object{PosVec: types.Ptf(10, 0)}
			a.Shape.Kind, a.Shape.Circle.R = tc.kind, tc.radius
			a.Shape.Box.W, a.Shape.Box.H = tc.width, tc.height
			if got := ObjectDistance4E6C00(&a, &Object{}); math.Float64bits(got) != math.Float64bits(tc.want) {
				t.Fatalf("collision distance = %.17g, want %.17g", got, tc.want)
			}
		})
	}
	shape := Shape{Kind: ShapeKindBox}
	shape.Box.W = math.SmallestNonzeroFloat32
	if got := objectDistanceShapeExtent4E6C00(&shape); got != float64(math.SmallestNonzeroFloat32)*0.5 {
		t.Fatal("width was rounded before maximum")
	}
	shape.Box.W, shape.Box.H = 0, math.SmallestNonzeroFloat32
	if got := objectDistanceShapeExtent4E6C00(&shape); got != 0 {
		t.Fatal("height was not rounded before maximum")
	}
	a, b := Object{PosVec: types.Ptf(math.MaxFloat32, 0)}, Object{PosVec: types.Ptf(-math.MaxFloat32, 0)}
	if got, want := ObjectDistance4E6C00(&a, &b), 2*float64(math.MaxFloat32); got != want {
		t.Fatal("position subtraction narrowed before distance")
	}
	a.PosVec.X = float32(math.Inf(1))
	if !math.IsInf(ObjectDistance4E6C00(&a, &b), 1) {
		t.Fatal("ordered infinity was clamped")
	}
	a.PosVec.X = float32(math.NaN())
	if ObjectDistance4E6C00(&a, &b) != float64(float32(0.01)) {
		t.Fatal("unordered position was not clamped")
	}
}
