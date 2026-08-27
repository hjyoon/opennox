package legacy

import (
	"math"
	"strconv"
	"testing"
)

func TestPlayerQuestGuideAllowedNative4F2EF0(t *testing.T) {
	tests := []struct {
		guide int
		want  bool
	}{
		{math.MinInt32, false}, {-1, false}, {0, true}, {1, false},
		{2, true}, {7, true}, {8, true}, {12, false}, {24, true},
		{25, true}, {26, true}, {28, false}, {40, true}, {41, false},
		{math.MaxInt32, false},
	}
	for _, test := range tests {
		if got := playerQuestGuideAllowedNative4F2EF0(test.guide); got != test.want {
			t.Errorf("guide %d admission = %v, want %v", test.guide, got, test.want)
		}
	}
	if strconv.IntSize == 64 {
		if got := playerQuestGuideAllowedNative4F2EF0(int(int64(1)<<32 | 7)); !got {
			t.Fatal("native wrapper did not preserve the original low int32 bits")
		}
	}
}
