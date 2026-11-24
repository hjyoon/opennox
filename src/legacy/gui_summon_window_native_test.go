package legacy

import "testing"

func TestSummonCage4C1FE0KeepsSignedAnimationCoordinates(t *testing.T) {
	check := func(state byte, y int32, wantState byte, wantY int32) {
		t.Helper()
		gotState, gotY := stepSummonCage(state, y, 0, -145)
		if gotState != wantState || gotY != wantY {
			t.Fatalf("step from (%d, %d) = (%d, %d), want (%d, %d)", state, y, gotState, gotY, wantState, wantY)
		}
	}
	check(1, -145, 1, -125)
	check(1, -5, 2, 0)
	check(2, 0, 2, 0)
	check(3, 0, 3, -20)
	check(3, -140, 0, -145)
	check(0, -145, 0, -145)
}
