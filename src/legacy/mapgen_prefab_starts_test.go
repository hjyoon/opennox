package legacy

import (
	"slices"
	"testing"
)

func TestMapgenPrefabStartLayouts(t *testing.T) {
	const size = 100
	for _, tc := range []struct {
		count, roll int
		samples     []float32
		want        [][2]float32
	}{
		{1, -1, []float32{16.5, -10}, [][2]float32{{16.5, -10}}},
		{2, 49, []float32{16.5, -10}, [][2]float32{{16.5, -size}, {-10, size}}},
		{2, 50, []float32{16.5, -10}, [][2]float32{{-size, 16.5}, {size, -10}}},
		{3, 24, nil, [][2]float32{{-size, -size}, {size, -size}, {0, size}}},
		{3, 25, nil, [][2]float32{{-size, size}, {size, size}, {0, -size}}},
		{3, 50, nil, [][2]float32{{size, -size}, {size, size}, {-size, 0}}},
		{3, 75, nil, [][2]float32{{-size, -size}, {-size, size}, {size, 0}}},
		{4, -1, nil, [][2]float32{{-size, -size}, {size, -size}, {-size, size}, {size, size}, {0, 0}}},
		{5, -1, nil, [][2]float32{{-size, -size}, {size, -size}, {-size, size}, {size, size}, {0, 0}}},
	} {
		samples, rolls := tc.samples, 0
		got := mapgenPrefabStarts(tc.count, size, func() int {
			rolls++
			return tc.roll
		}, func() float32 {
			if len(samples) == 0 {
				t.Fatalf("count %d, roll %d: unexpected coordinate sample", tc.count, tc.roll)
			}
			value := samples[0]
			samples = samples[1:]
			return value
		})
		wantRolls := 0
		if tc.roll >= 0 {
			wantRolls = 1
		}
		if len(samples) != 0 || rolls != wantRolls || !slices.Equal(got[:len(tc.want)], tc.want) {
			t.Fatalf("count %d, roll %d: starts %v, samples remaining %d, rolls %d",
				tc.count, tc.roll, got, len(samples), rolls)
		}
	}
}

func TestMapgenPrefabStartShuffle(t *testing.T) {
	for _, tc := range []struct {
		count int
		rolls []int
		want  [5]uint32
	}{
		{0, nil, [5]uint32{9, 9, 9, 9, 9}},
		{1, nil, [5]uint32{0, 9, 9, 9, 9}},
		{2, []int{49}, [5]uint32{0, 1, 9, 9, 9}},
		{2, []int{50}, [5]uint32{1, 0, 9, 9, 9}},
		{3, []int{1, 0, 1}, [5]uint32{2, 1, 0, 9, 9}},
		{4, []int{0, 2, 1, 0}, [5]uint32{3, 1, 0, 2, 9}},
		{5, []int{3, 3, 1, 0, 2}, [5]uint32{2, 0, 3, 1, 4}},
	} {
		order := [5]uint32{9, 9, 9, 9, 9}
		rolls := tc.rolls
		mapgenShufflePrefabStarts(order[:tc.count], func(lo, hi int) int {
			wantHi := tc.count - 2
			if tc.count == 2 {
				wantHi = 100
			}
			if len(rolls) == 0 || lo != 0 || hi != wantHi {
				t.Fatalf("count %d: unexpected shuffle range [%d,%d]", tc.count, lo, hi)
			}
			value := rolls[0]
			rolls = rolls[1:]
			return value
		})
		if len(rolls) != 0 || order != tc.want {
			t.Fatalf("count %d: order = %v, want %v, rolls remaining = %d", tc.count, order, tc.want, len(rolls))
		}
	}
}
