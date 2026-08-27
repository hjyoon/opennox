package legacy

import (
	"math"
	"testing"
)

func TestRandomAbilityLossEligibilityCall4F2570UsesExactInt32(t *testing.T) {
	tests := []struct {
		abilityID int32
		want      int32
	}{
		{abilityID: math.MinInt32},
		{abilityID: -1},
		{abilityID: 0},
		{abilityID: 1, want: 1},
		{abilityID: 2, want: 1},
		{abilityID: 3, want: 1},
		{abilityID: 4, want: 1},
		{abilityID: 5, want: 1},
		{abilityID: 6},
		{abilityID: math.MaxInt32},
	}
	for _, test := range tests {
		if got := randomAbilityLossEligibilityCall4F2570(test.abilityID); got != test.want {
			t.Fatalf("ability %d eligibility = %d, want %d", test.abilityID, got, test.want)
		}
	}
}
