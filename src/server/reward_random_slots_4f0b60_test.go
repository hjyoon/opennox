package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

func TestRewardSlotWeightsHalfMatchGAMEEXE4F0B60(t *testing.T) {
	tests := []struct {
		stage uint32
		want  [5]uint16
		uses  bool
	}{
		{0, [5]uint16{}, false},
		{1, [5]uint16{175, 25}, true},
		{2, [5]uint16{100, 100}, true},
		{3, [5]uint16{25, 150, 25}, true},
		{4, [5]uint16{0, 100, 100}, true},
		{5, [5]uint16{0, 25, 150, 25}, true},
		{6, [5]uint16{0, 0, 100, 100}, true},
		{7, [5]uint16{0, 0, 25, 150, 25}, true},
		{8, [5]uint16{0, 0, 0, 100, 100}, true},
		{9, [5]uint16{0, 0, 0, 25, 175}, true},
		{10, [5]uint16{}, false},
		{11, [5]uint16{}, false},
		{math.MaxUint32, [5]uint16{}, false},
	}
	for _, test := range tests {
		t.Run(fmt.Sprint(test.stage), func(t *testing.T) {
			got, uses := rewardSlotWeightsHalf4F0B60(test.stage)
			if got != test.want || uses != test.uses {
				t.Fatalf("weights/uses = %v/%t, want %v/%t", got, uses, test.want, test.uses)
			}
		})
	}
}

func TestRewardRandomSlotsAllInclusiveDraws4F0B60(t *testing.T) {
	wantCounts := map[uint32][5]int{
		1: {176, 25, 0, 0, 0},
		2: {101, 100, 0, 0, 0},
		3: {26, 150, 25, 0, 0},
		4: {1, 100, 100, 0, 0},
		5: {1, 25, 150, 25, 0},
		6: {1, 0, 100, 100, 0},
		7: {1, 0, 25, 150, 25},
		8: {1, 0, 0, 100, 100},
		9: {1, 0, 0, 25, 175},
	}
	for stage := uint32(1); stage <= 9; stage++ {
		var counts [5]int
		for draw := int32(0); draw <= 200; draw++ {
			calls := 0
			got := rewardRandomSlots4F0B60(stage, func(minimum, maximum int32) int32 {
				calls++
				if minimum != 0 || maximum != 200 {
					t.Fatalf("stage %d RNG bounds = %d..%d", stage, minimum, maximum)
				}
				return draw
			})
			if calls != 1 || got == 0 || got&uint32(got-1) != 0 || got > 16 {
				t.Fatalf("stage %d draw %d result/calls = %#x/%d", stage, draw, got, calls)
			}
			index := 0
			for uint32(1)<<index != got {
				index++
			}
			counts[index]++
		}
		if counts != wantCounts[stage] {
			t.Fatalf("stage %d counts = %v, want %v", stage, counts, wantCounts[stage])
		}
	}
}

func TestRewardRandomSlotsFixedStagesDoNotUseRNG4F0B60(t *testing.T) {
	tests := []struct {
		stage uint32
		want  uint32
	}{
		{0, 1},
		{10, 16},
		{11, 16},
		{math.MaxUint32, 16},
	}
	for _, test := range tests {
		t.Run(fmt.Sprint(test.stage), func(t *testing.T) {
			got := rewardRandomSlots4F0B60(test.stage, func(int32, int32) int32 {
				t.Fatal("fixed stage called RNG")
				return 0
			})
			if got != test.want {
				t.Fatalf("result = %#x, want %#x", got, test.want)
			}
		})
	}
}

func TestRewardRandomSlotsWeightedStagePreservesNilRNGFault4F0B60(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("weighted stage did not fault on a nil RNG callback")
		}
	}()
	rewardRandomSlots4F0B60(1, nil)
}

func TestRewardRandomSlotsBoundariesAndFallback4F0B60(t *testing.T) {
	tests := []struct {
		stage uint32
		draw  int32
		want  uint32
	}{
		{1, -1, 1},
		{1, 175, 1},
		{1, 176, 2},
		{3, 25, 1},
		{3, 26, 2},
		{3, 175, 2},
		{3, 176, 4},
		{9, 0, 1},
		{9, 1, 8},
		{9, 25, 8},
		{9, 26, 16},
		{9, 200, 16},
		{9, 201, 1},
		{9, math.MaxInt32, 1},
		{9, math.MinInt32, 1},
	}
	var events []string
	for _, test := range tests {
		got := rewardRandomSlots4F0B60(test.stage, func(minimum, maximum int32) int32 {
			events = append(events, fmt.Sprintf("%d:%d..%d=%d", test.stage, minimum, maximum, test.draw))
			return test.draw
		})
		if got != test.want {
			t.Fatalf("stage %d draw %d result = %#x, want %#x", test.stage, test.draw, got, test.want)
		}
	}
	if len(events) != len(tests) || slices.Contains(events, "") {
		t.Fatalf("RNG events = %q", events)
	}
}
