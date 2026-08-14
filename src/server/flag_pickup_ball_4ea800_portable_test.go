package server

import "testing"

func TestFlagPickupBallPortableContract4EA800(t *testing.T) {
	events := make([]uint32, 0, 5)
	flagPickupBall4EA800(1, 2, 3, flagPickupBallHooks4EA800[int, int, int, int]{
		loadBallCache: func() uint32 {
			events = append(events, 1)
			return 0
		},
		lookupType: func(name string) uint32 {
			events = append(events, 2)
			if name != "GameBall" {
				t.Fatalf("lookup = %q", name)
			}
			return 7
		},
		storeBallCache: func(ind uint32) {
			events = append(events, 3)
			if ind != 7 {
				t.Fatalf("cache = %d", ind)
			}
		},
		loadClassLow: func(target int) uint8 {
			events = append(events, 4)
			if target != 2 {
				t.Fatalf("target = %d", target)
			}
			return 4
		},
		unitIsGameBall: func(target int) int32 {
			events = append(events, 5)
			return 0
		},
	})
	want := []uint32{1, 2, 3, 4, 5}
	if !reflectUint32sEqual4EA800(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func reflectUint32sEqual4EA800(got, want []uint32) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
