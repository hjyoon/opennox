package legacy

import (
	"math"
	"testing"
)

func TestPlayerQuestSpellAllowedNative4F2E70(t *testing.T) {
	tests := []struct {
		spellID int
		want    bool
	}{
		{math.MinInt32, false}, {-1, false}, {0, false},
		{45, false}, {46, true}, {49, true},
		{74, true}, {75, true}, {114, true}, {115, false},
		{121, false}, {122, true}, {125, true}, {126, false},
		{math.MaxInt32, false},
	}
	for _, test := range tests {
		if got := playerQuestSpellAllowedNative4F2E70(test.spellID); got != test.want {
			t.Errorf("spell %d admission = %v, want %v", test.spellID, got, test.want)
		}
	}
}
