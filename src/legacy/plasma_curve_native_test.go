package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestPlasmaCurveUsesOriginalForwardDifferenceTable(t *testing.T) {
	basis := memmap.PtrT[[4][4]int32](0x581450, 9872)
	oldBasis := *basis
	t.Cleanup(func() { *basis = oldBasis })
	*basis = [4][4]int32{{2, -2, 1, 1}, {-3, 3, -2, -1}, {0, 0, 1, 0}, {1, 0, 0, 0}}

	steps := memmap.PtrT[[4][4]float32](0x587000, 180460)
	oldSteps := *steps
	t.Cleanup(func() { *steps = oldSteps })
	*steps = [4][4]float32{{0, 0, 0, 1}}

	a := plasmaCurvePoint{}
	b := plasmaCurvePoint{field_0: 100}
	c := plasmaCurvePoint{field_4: 100}
	d := plasmaCurvePoint{field_4: -100}
	x, y := plasmaCurveCoefficients(a, b, c, d)
	dx, dy := plasmaCurveForwardDeltas(4, x, y)
	want := []plasmaCurvePoint{{field_0: 15, field_4: 18}, {field_0: 50, field_4: 25},
		{field_0: 84, field_4: 18}, {field_0: 100}}
	for i, p := range want {
		for j := 0; j < 3; j++ {
			dx[j] += dx[j+1]
			dy[j] += dy[j+1]
		}
		got := plasmaCurvePoint{field_0: plasmaCurveFloatToInt(dx[0]), field_4: plasmaCurveFloatToInt(dy[0])}
		if got != p {
			t.Errorf("forward point %d = %+v, want %+v", i, got, p)
		}
	}
	previous := a
	count := 0
	plasmaSplineSamples(a, b, c, d, 4, 0, func(from, to *plasmaCurvePoint) {
		if count >= len(want) || *from != previous || *to != want[count] {
			t.Fatalf("spline segment %d = %+v -> %+v", count, *from, *to)
		}
		previous = *to
		count++
	})
	if count != len(want) {
		t.Fatalf("spline produced %d segments, want %d", count, len(want))
	}
}
