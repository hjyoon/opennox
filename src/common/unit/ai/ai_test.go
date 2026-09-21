package ai

import (
	"math"
	"testing"
)

func TestActionTypeIsCondition50A010SignedBoundary(t *testing.T) {
	tests := []struct {
		action ActionType
		want   bool
	}{
		{action: ActionType(0x80000000), want: false},
		{action: ACTION_INVALID, want: false},
		{action: ActionType(39), want: false},
		{action: DEPENDENCY_OR, want: true},
		{action: DEPENDENCY_NOT_MOVED, want: true},
		{action: ActionType(72), want: true},
		{action: ActionType(math.MaxInt32), want: true},
		{action: ActionType(math.MaxUint32), want: false},
	}
	for _, tc := range tests {
		if got := tc.action.IsCondition(); got != tc.want {
			t.Errorf("action %#x IsCondition = %t, want %t", uint32(tc.action), got, tc.want)
		}
	}
}
