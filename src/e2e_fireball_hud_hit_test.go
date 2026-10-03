package opennox

import "testing"

func TestE2EFireballHUDHitRequiresActualCurrentAndOriginalMaximum(t *testing.T) {
	for _, tc := range []struct {
		name             string
		current, maximum uint32
		ready, want      bool
	}{
		{"actual hit", 1936, 75, true, true},
		{"stale current", 2000, 75, true, false},
		{"different hit", 1935, 75, true, false},
		{"fixture maximum is not published", 1936, 2000, true, false},
		{"unavailable HUD", 1936, 75, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eFireballHUDHit(e2eHUDMeterState{Current: tc.current, Maximum: tc.maximum}, tc.ready, 2000, 64, 75); got != tc.want {
				t.Fatalf("hit=%t want=%t", got, tc.want)
			}
		})
	}
}
