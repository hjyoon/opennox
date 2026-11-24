package legacy

import (
	"math"
	"slices"
	"testing"
)

func TestMapgenFloorFillBeveledStaging(t *testing.T) {
	area := mapgenFloorFillArea{
		origin: mapgenWorldPoint{x: math.Float32frombits(0xc2c32956), y: -100},
		width:  9,
		height: 2,
	}
	type row struct {
		x, y  uint32
		width int32
	}
	var observed []row
	area.drawBeveled(func(point mapgenWorldPoint, width, height int32) {
		if height != 1 {
			t.Fatalf("beveled row has height %d", height)
		}
		observed = append(observed, row{math.Float32bits(point.x), math.Float32bits(point.y), width})
	})
	want := []row{
		{0x36800000, 0xc2c80000, 3},
		{0xc2021b8e, 0xc286f238, 5},
		{0, 0xc286f238, 3},
	}
	if !slices.Equal(observed, want) {
		t.Fatalf("beveled rows = %x, want %x", observed, want)
	}
	point := area.center()
	if math.Float32bits(point.x) != 0x42432958 || math.Float32bits(point.y) != 0xc286f238 {
		t.Fatalf("beveled center = %v", point)
	}
}

func TestMapgenFloorFillIrregularCallbacks(t *testing.T) {
	const (
		rectangle = "rectangle"
		column    = "column"
		roll      = "roll"
	)
	type event struct {
		kind string
		x, y uint32
		a, b int32
	}
	for _, tc := range []struct {
		name          string
		width, height int32
		rolls         []int32
		want          []event
	}{
		{"3x3", 3, 3, nil, []event{
			{rectangle, 0x430486e4, 0xc286f238, 1, 1},
			{rectangle, 0x430486e4, 0xc2c80000, 1, 1},
			{rectangle, 0x430486e4, 0xc20bc8e2, 1, 1},
			{rectangle, 0x42c80000, 0xc286f238, 1, 1},
			{rectangle, 0x43250dc8, 0xc286f238, 1, 1},
		}},
		{"6x5", 6, 5, []int32{1, 2, 1, 2, 1, 2, 1, 1}, []event{
			{rectangle, 0x430486e4, 0xc286f238, 4, 3},
			{roll, 0, 0, 1, 3}, {roll, 0, 0, 1, 3},
			{rectangle, 0x43250dc8, 0xc2c80000, 1, 1},
			{roll, 0, 0, 1, 3}, {roll, 0, 0, 1, 3},
			{rectangle, 0x43250dc8, 0x41f0dc78, 1, 1},
			{roll, 0, 0, 1, 2}, {roll, 0, 0, 1, 2},
			{column, 0x42c80000, 0xc20bc8e2, 1, 0},
			{roll, 0, 0, 1, 2}, {roll, 0, 0, 1, 2},
			{column, 0x43835139, 0xc286f238, 1, 0},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			area := mapgenFloorFillArea{origin: mapgenWorldPoint{x: 100, y: -100}, width: tc.width, height: tc.height}
			var observed []event
			rolls := tc.rolls
			area.drawIrregular(mapgenFloorFillOps{
				rectangle: func(point mapgenWorldPoint, w, h int32) {
					observed = append(observed, event{rectangle, math.Float32bits(point.x), math.Float32bits(point.y), w, h})
				},
				column: func(point mapgenWorldPoint, h int32) {
					observed = append(observed, event{column, math.Float32bits(point.x), math.Float32bits(point.y), h, 0})
				},
				random: func(lo, hi int32) int32 {
					if len(rolls) == 0 {
						t.Fatalf("unexpected RNG call [%d,%d]", lo, hi)
					}
					observed = append(observed, event{roll, 0, 0, lo, hi})
					result := rolls[0]
					rolls = rolls[1:]
					return result
				},
			})
			if len(rolls) != 0 || !slices.Equal(observed, tc.want) {
				t.Fatalf("fill callbacks = %#v, want %#v", observed, tc.want)
			}
		})
	}
}

func TestMapgenFloorFillMissingRecord(t *testing.T) {
	if got := sub_524950(73, 0, nil, nil); got != 0 {
		t.Fatalf("rectangle returned %d for a missing fill", got)
	}
	if got := sub_5249C0(73, 0, nil, nil); got != 0 {
		t.Fatalf("beveled fill returned %d for a missing fill", got)
	}
	sub_524B50(73, 0, nil, nil)
}
