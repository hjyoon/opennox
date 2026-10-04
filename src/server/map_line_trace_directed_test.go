package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/types"
)

func TestLineTraceXxxDirectedSegments(t *testing.T) {
	// These integer-coordinate crossings are independent of the implementation.
	// Rectf holds directed endpoints here, NOT an axis-aligned search rectangle.
	for _, tc := range []struct {
		name       string
		line, edge types.Rectf
		want       bool
	}{
		{"descending-crosses-vertical", types.Rectf{Min: types.Ptf(0, 10), Max: types.Ptf(10, 0)}, types.Rectf{Min: types.Ptf(2, 7), Max: types.Ptf(2, 9)}, true},
		{"descending-misses-mirrored-vertical", types.Rectf{Min: types.Ptf(0, 10), Max: types.Ptf(10, 0)}, types.Rectf{Min: types.Ptf(2, 1), Max: types.Ptf(2, 3)}, false},
		{"opposite-diagonals-cross", types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(10, 10)}, types.Rectf{Min: types.Ptf(0, 10), Max: types.Ptf(10, 0)}, true},
		{"descending-crosses-horizontal", types.Rectf{Min: types.Ptf(0, 10), Max: types.Ptf(10, 0)}, types.Rectf{Min: types.Ptf(5, 4), Max: types.Ptf(7, 4)}, true},
		{"descending-misses-mirrored-horizontal", types.Rectf{Min: types.Ptf(0, 10), Max: types.Ptf(10, 0)}, types.Rectf{Min: types.Ptf(3, 4), Max: types.Ptf(5, 4)}, false},
		{"ascending-crosses-vertical", types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(10, 10)}, types.Rectf{Min: types.Ptf(2, 1), Max: types.Ptf(2, 3)}, true},
	} {
		for direction := 0; direction < 4; direction++ {
			t.Run(fmt.Sprintf("%s/reverse-%d", tc.name, direction), func(t *testing.T) {
				line, edge := tc.line, tc.edge
				if direction&1 != 0 {
					line.Min, line.Max = line.Max, line.Min
				}
				if direction&2 != 0 {
					edge.Min, edge.Max = edge.Max, edge.Min
				}
				beforeLine, beforeEdge := line, edge
				if got := LineTraceXxx(line, edge); got != tc.want {
					t.Fatalf("intersection(%+v, %+v)=%t want=%t", line, edge, got, tc.want)
				}
				if line != beforeLine || edge != beforeEdge {
					t.Fatal("input endpoint mutation")
				}
			})
		}
	}
}

func TestLineTraceXxxOriginalBoundaryGates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		line, edge types.Rectf
		want       bool
	}{
		{"horizontal-overlap", types.Rectf{Min: types.Ptf(8, 2), Max: types.Ptf(0, 2)}, types.Rectf{Min: types.Ptf(3, 2), Max: types.Ptf(5, 2)}, true},
		{"horizontal-end-touch", types.Rectf{Min: types.Ptf(0, 2), Max: types.Ptf(3, 2)}, types.Rectf{Min: types.Ptf(3, 2), Max: types.Ptf(5, 2)}, true},
		{"horizontal-disjoint", types.Rectf{Min: types.Ptf(0, 2), Max: types.Ptf(2, 2)}, types.Rectf{Min: types.Ptf(3, 2), Max: types.Ptf(5, 2)}, false},
		{"vertical-collinear-original-rejection", types.Rectf{Min: types.Ptf(2, 0), Max: types.Ptf(2, 8)}, types.Rectf{Min: types.Ptf(2, 3), Max: types.Ptf(2, 5)}, false},
		{"parallel-diagonals", types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(8, 8)}, types.Rectf{Min: types.Ptf(0, 2), Max: types.Ptf(6, 8)}, false},
		{"edge-start-on-line-original-rejection", types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(8, 8)}, types.Rectf{Min: types.Ptf(4, 4), Max: types.Ptf(4, 6)}, false},
		{"zero-length-original-rejection", types.Rectf{Min: types.Ptf(4, 4), Max: types.Ptf(4, 4)}, types.Rectf{Min: types.Ptf(3, 3), Max: types.Ptf(5, 5)}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LineTraceXxx(tc.line, tc.edge); got != tc.want {
				t.Fatalf("got=%t want=%t", got, tc.want)
			}
		})
	}
}

func TestLineTraceXxxOriginalCrossProductSpill(t *testing.T) {
	// Positive, almost parallel segments cross inside both bounds. The original
	// spills the deltas to binary32, but only spills the cross products AFTER
	// subtracting the x87 products. Rounding each product early loses 1/16 here.
	line := types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(5000, 4999.75)}
	edge := types.Rectf{Min: types.Ptf(0, 0.00000625), Max: types.Ptf(4999.75, 4999.5)}
	a1w, a1h := line.Max.X-line.Min.X, line.Max.Y-line.Min.Y
	a2w, a2h := edge.Max.X-edge.Min.X, edge.Max.Y-edge.Min.Y
	early := float32(a2w*a1h) - float32(a2h*a1w)
	exact := float64(a2w)*float64(a1h) - float64(a2h)*float64(a1w)
	if early != 0 || exact != 0.0625 {
		t.Fatalf("precision precondition: early=%g exact=%g", early, exact)
	}
	if !LineTraceXxx(line, edge) {
		t.Fatal("original nonzero cross product was rounded into a parallel ray")
	}
}
