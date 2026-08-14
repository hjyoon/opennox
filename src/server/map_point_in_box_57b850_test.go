package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func mapPointInBoxTestShape57B850(width, height float32) Shape {
	shape := Shape{Kind: ShapeKindBox}
	shape.Box.W = width
	shape.Box.H = height
	shape.Box.Calc()
	return shape
}

func TestMapPointInBox57B850StrictEdges(t *testing.T) {
	shape := mapPointInBoxTestShape57B850(20, 20)
	source := types.Ptf(30, -10)

	for _, tc := range []struct {
		name   string
		target types.Pointf
		want   bool
	}{
		{name: "center", target: source, want: true},
		{name: "inside", target: types.Ptf(31, -9), want: true},
		{name: "outside", target: types.Ptf(100, -10), want: false},
		{name: "first edge", target: types.Ptf(40, 0), want: false},
		{name: "second edge", target: types.Ptf(20, 0), want: false},
		{name: "third edge", target: types.Ptf(40, 0), want: false},
		{name: "fourth edge", target: types.Ptf(20, -20), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mapPointInBox57B850(source, &shape, tc.target); got != tc.want {
				t.Fatalf("inside = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestMapPointInBox57B850UnorderedGates(t *testing.T) {
	shape := mapPointInBoxTestShape57B850(20, 20)
	qnan := math.Float32frombits(0x7fc12345)

	firstAndFourth := shape
	firstAndFourth.Box.LeftBottom = qnan
	if !mapPointInBox57B850(types.Pointf{}, &firstAndFourth, types.Pointf{}) {
		t.Fatal("unordered C0-only gates were rejected")
	}

	second := shape
	second.Box.LeftTop2 = qnan
	if mapPointInBox57B850(types.Pointf{}, &second, types.Pointf{}) {
		t.Fatal("unordered second diagonal gate was accepted")
	}

	third := shape
	third.Box.RightTop = qnan
	if mapPointInBox57B850(types.Pointf{}, &third, types.Pointf{}) {
		t.Fatal("unordered third diagonal gate was accepted")
	}
}

func TestMapPointInBox57B850KeepsSecondDiagonalInX87Precision(t *testing.T) {
	shape := Shape{Kind: ShapeKindBox}
	shape.Box.LeftTop = 0
	shape.Box.LeftBottom = -3
	shape.Box.LeftTop2 = 1 << 24
	shape.Box.LeftBottom2 = 1 << 24
	shape.Box.RightTop = 1
	shape.Box.RightBottom = 1

	// In GAME.EXE, (1<<24)+1 remains 16777217 in the x87 register while
	// (1<<24)+0 is 16777216. Spilling both operands to binary32 would make the
	// strict second gate equal to zero and incorrectly reject this point.
	source := types.Ptf(0, 1)
	if !mapPointInBox57B850(source, &shape, types.Pointf{}) {
		t.Fatal("unspilled x87 second diagonal was rounded to binary32")
	}

	leftTop2 := float32(float64(shape.Box.LeftTop2) + float64(source.Y))
	leftBottom2 := float32(float64(shape.Box.LeftBottom2) + float64(source.X))
	if leftTop2-leftBottom2 != 0 {
		t.Fatal("regression vector no longer distinguishes a binary32 spill")
	}
}

func TestMapPointInBox57B850ConstantAndPointerBinding(t *testing.T) {
	if got := math.Float32frombits(mapPointInBoxScaleBits57B850); got != float32(0.70709997) {
		t.Fatalf("scale = %v (%#08x)", got, mapPointInBoxScaleBits57B850)
	}
	shape := mapPointInBoxTestShape57B850(20, 20)
	source := types.Ptf(-7.5, 11.25)
	target := source
	if !MapPointInBox57B850(&source, &shape, &target) {
		t.Fatal("native pointer binding rejected center")
	}
}
