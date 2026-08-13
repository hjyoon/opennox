package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectDistance4E6C00ShapeContracts(t *testing.T) {
	tests := []struct {
		name string
		a    server.Object
		b    server.Object
		want float64
	}{
		{
			name: "center distance",
			a:    server.Object{PosVec: types.Ptf(0, 0)},
			b:    server.Object{PosVec: types.Ptf(3, 4)},
			want: 5,
		},
		{
			name: "circle radii",
			a: server.Object{PosVec: types.Ptf(0, 0), Shape: server.Shape{
				Kind: server.ShapeKindCircle,
				Circle: struct {
					R  float32
					R2 float32
				}{R: 2},
			}},
			b: server.Object{PosVec: types.Ptf(10, 0), Shape: server.Shape{
				Kind: server.ShapeKindCircle,
				Circle: struct {
					R  float32
					R2 float32
				}{R: 3},
			}},
			want: 5,
		},
		{
			name: "box larger width and height",
			a:    boxDistanceObject4E6C00(0, 0, 6, 4),
			b:    boxDistanceObject4E6C00(10, 0, 2, 8),
			want: 3,
		},
		{
			name: "unrecognized shape ignored",
			a: server.Object{PosVec: types.Ptf(0, 0), Shape: server.Shape{
				Kind: 99,
				Circle: struct {
					R  float32
					R2 float32
				}{R: 100},
			}},
			b:    server.Object{PosVec: types.Ptf(6, 8)},
			want: 10,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := objectDistance_4E6C00(&tc.a, &tc.b); got != tc.want {
				t.Fatalf("got %.17g, want %.17g", got, tc.want)
			}
		})
	}
}

func TestObjectDistance4E6C00X87UnorderedAndFloor(t *testing.T) {
	minimum := float64(float32(0.01))
	tests := []struct {
		name string
		a    server.Object
		b    server.Object
		want float64
	}{
		{
			name: "overlap clamps",
			a:    boxDistanceObject4E6C00(0, 0, 20, 20),
			b:    server.Object{PosVec: types.Ptf(1, 0)},
			want: minimum,
		},
		{
			name: "position NaN clamps",
			a:    server.Object{PosVec: types.Ptf(float32(math.NaN()), 0)},
			b:    server.Object{},
			want: minimum,
		},
		{
			name: "circle NaN clamps",
			a: server.Object{PosVec: types.Ptf(10, 0), Shape: server.Shape{
				Kind: server.ShapeKindCircle,
				Circle: struct {
					R  float32
					R2 float32
				}{R: float32(math.NaN())},
			}},
			b:    server.Object{},
			want: minimum,
		},
		{
			name: "box width NaN selects finite height",
			a:    boxDistanceObject4E6C00(10, 0, float32(math.NaN()), 4),
			b:    server.Object{},
			want: 8,
		},
		{
			name: "box height NaN survives and clamps",
			a:    boxDistanceObject4E6C00(10, 0, 4, float32(math.NaN())),
			b:    server.Object{},
			want: minimum,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := objectDistance_4E6C00(&tc.a, &tc.b)
			if math.Float64bits(got) != math.Float64bits(tc.want) {
				t.Fatalf("got %.17g (%#x), want %.17g (%#x)", got, math.Float64bits(got), tc.want, math.Float64bits(tc.want))
			}
		})
	}
}

func TestObjectDistanceShapeExtent4E6C00RoundsOnlyHeight(t *testing.T) {
	shape := server.Shape{Kind: server.ShapeKindBox}
	shape.Box.W = math.SmallestNonzeroFloat32
	shape.Box.H = 0
	if got, want := objectDistanceShapeExtent_4E6C00(&shape), float64(math.SmallestNonzeroFloat32)*0.5; got != want {
		t.Fatalf("width half = %.17g, want %.17g", got, want)
	}
	shape.Box.W = 0
	shape.Box.H = math.SmallestNonzeroFloat32
	if got := objectDistanceShapeExtent_4E6C00(&shape); got != 0 {
		t.Fatalf("rounded height half = %.17g, want 0", got)
	}
}

func TestObjectDistance4E6C00InfinityAndNil(t *testing.T) {
	a := server.Object{PosVec: types.Ptf(float32(math.Inf(1)), 0)}
	if got := objectDistance_4E6C00(&a, &server.Object{}); !math.IsInf(got, 1) {
		t.Fatalf("positive infinity = %v, want +Inf", got)
	}
	for _, tc := range []struct {
		name string
		a    *server.Object
		b    *server.Object
	}{
		{name: "nil first", b: &server.Object{}},
		{name: "nil second", a: &server.Object{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("nil object did not panic")
				}
			}()
			objectDistance_4E6C00(tc.a, tc.b)
		})
	}
}

func boxDistanceObject4E6C00(x, y, width, height float32) server.Object {
	obj := server.Object{PosVec: types.Ptf(x, y)}
	obj.Shape.Kind = server.ShapeKindBox
	obj.Shape.Box.W = width
	obj.Shape.Box.H = height
	return obj
}
