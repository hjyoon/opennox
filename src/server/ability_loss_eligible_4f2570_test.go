package server

import (
	"math"
	"testing"
)

func TestRandomAbilityLossEligibilityExactRange4F2570(t *testing.T) {
	for abilityID := int32(-8); abilityID <= 12; abilityID++ {
		var want int32
		if abilityID >= 1 && abilityID <= 5 {
			want = 1
		}
		if got := randomAbilityLossEligible4F2570(abilityID); got != want {
			t.Fatalf("ability %d eligibility = %d, want %d", abilityID, got, want)
		}
	}
}

func TestRandomAbilityLossEligibilitySignedInt32Boundaries4F2570(t *testing.T) {
	tests := []struct {
		abilityID int32
		want      int32
	}{
		{abilityID: math.MinInt32},
		{abilityID: -1},
		{abilityID: 0},
		{abilityID: 1, want: 1},
		{abilityID: 5, want: 1},
		{abilityID: 6},
		{abilityID: math.MaxInt32},
	}
	for _, test := range tests {
		if got := RandomAbilityLossEligible4F2570(test.abilityID); got != test.want {
			t.Fatalf("ability %d eligibility = %d, want %d", test.abilityID, got, test.want)
		}
	}
}
