package server

import "testing"

func TestDeathBallFragmentCollidePortableContract4E9FE0(t *testing.T) {
	events := make([]uint32, 0, 3)
	deathBallFragmentCollide4E9FE0(1, 0, 2, deathBallFragmentCollideHooks4E9FE0[int, int]{
		wallReflect: func(collision, source int) {
			if collision != 2 || source != 1 {
				t.Fatalf("reflect = %d/%d", collision, source)
			}
			events = append(events, 1)
		},
		audio: func(id uint32, source int) {
			if id != deathBallFragmentWallAudio4E9FE0 || source != 1 {
				t.Fatalf("audio = %d/%d", id, source)
			}
			events = append(events, 2)
		},
		loadNewPosY: func(int) float32 { return 23 },
		loadNewPosX: func(int) float32 { return 46 },
		floatToInt: func(value float32) int32 {
			return int32(value)
		},
		damageMap: func(x, y, damage int32, damageType uint32, source int) {
			if x != 2 || y != 1 || damage != 20 || damageType != 2 || source != 1 {
				t.Fatalf("map = %d/%d/%d/%d/%d", x, y, damage, damageType, source)
			}
			events = append(events, 3)
		},
		delayedDelete: func(int) { t.Fatal("wall path deleted source") },
	})
	if len(events) != 3 || events[0] != 1 || events[1] != 2 || events[2] != 3 {
		t.Fatalf("events = %v", events)
	}
}
