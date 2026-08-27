package legacy

import (
	"math"
	"testing"
)

func TestRandomFieldGuideLossEligibilityCall4F2530UsesExactInt32(t *testing.T) {
	tests := []struct {
		guideID int32
		want    int32
	}{
		{guideID: 2, want: 1},
		{guideID: 32, want: 1},
		{guideID: 40, want: 1},
		{guideID: 0},
		{guideID: 1},
		{guideID: 5},
		{guideID: 41},
		{guideID: -1},
		{guideID: math.MinInt32},
		{guideID: math.MaxInt32},
	}
	for _, test := range tests {
		if got := randomFieldGuideLossEligibilityCall4F2530(test.guideID); got != test.want {
			t.Fatalf("guide %d eligibility = %d, want %d", test.guideID, got, test.want)
		}
	}
}
