package legacy

import (
	"math"
	"testing"
)

func TestMatchStateExports5096F0PreserveInt32Results(t *testing.T) {
	oldLimit := matchLimitStateCall5096F0
	oldHigh := deathmatchHighScoreWinnerCall5098A0
	oldTeam := teamHighScoreWinnerCall5099B0
	t.Cleanup(func() {
		matchLimitStateCall5096F0 = oldLimit
		deathmatchHighScoreWinnerCall5098A0 = oldHigh
		teamHighScoreWinnerCall5099B0 = oldTeam
	})

	for _, want := range []int32{math.MinInt32, math.MaxInt32, -1, 0, 1} {
		matchLimitStateCall5096F0 = func() int32 { return want }
		if got := matchLimitStateExportCall5096F0(); got != want {
			t.Fatalf("sub_5096F0 C ABI result = %d, want %d", got, want)
		}
		deathmatchHighScoreWinnerCall5098A0 = func() int32 { return want }
		if got := deathmatchHighScoreWinnerExportCall5098A0(); got != want {
			t.Fatalf("sub_5098A0 C ABI result = %d, want %d", got, want)
		}
		teamHighScoreWinnerCall5099B0 = func() int32 { return want }
		if got := teamHighScoreWinnerExportCall5099B0(); got != want {
			t.Fatalf("sub_5099B0 C ABI result = %d, want %d", got, want)
		}
	}
}
