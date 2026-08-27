package legacy

import (
	"math"
	"testing"
)

func TestRandomSpellLossEligibilityCall4F24E0UsesExactInt32(t *testing.T) {
	tests := []struct {
		spellID int32
		want    int32
	}{
		{spellID: 71, want: 1},
		{spellID: 136, want: 1},
		{spellID: 9},
		{spellID: 27},
		{spellID: 34},
		{spellID: 41},
		{spellID: 19},
		{spellID: -1},
		{spellID: math.MinInt32},
		{spellID: math.MaxInt32},
	}
	for _, test := range tests {
		if got := randomSpellLossEligibilityCall4F24E0(test.spellID); got != test.want {
			t.Fatalf("spell %d eligibility = %d, want %d", test.spellID, got, test.want)
		}
	}
}
