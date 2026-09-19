package legacy

import (
	"math"
	"testing"
)

func TestDeathmatchLowScoreWinnerExport5095E0PreservesInt32Result(t *testing.T) {
	old := deathmatchLowScoreWinnerCall5095E0
	t.Cleanup(func() { deathmatchLowScoreWinnerCall5095E0 = old })
	for _, want := range []int32{math.MinInt32, math.MaxInt32, -1, 0, 1} {
		deathmatchLowScoreWinnerCall5095E0 = func() int32 { return want }
		if got := deathmatchLowScoreWinnerExportCall5095E0(); got != want {
			t.Fatalf("C ABI result = %d, want %d", got, want)
		}
	}
}
