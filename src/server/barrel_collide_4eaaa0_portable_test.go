package server

import "testing"

func TestBarrelCollidePortableContract4EAAA0(t *testing.T) {
	frames := map[int]uint32{1: 5}
	events := make([]uint32, 0, 4)
	barrelCollide4EAAA0(1, 2, 3, barrelCollideHooks4EAAA0[int]{
		loadFrame: func() uint32 {
			events = append(events, 1)
			return 9
		},
		loadLastFrame: func(source int) uint32 {
			events = append(events, 2)
			return frames[source]
		},
		storeFrame: func(source int, frame uint32) {
			events = append(events, 3)
			frames[source] = frame
		},
		audio: func(id uint32, source int) {
			events = append(events, 4)
			if id != barrelCollideSound4EAAA0 || source != 1 {
				t.Fatalf("audio = %d/%d", id, source)
			}
		},
	})
	if frames[1] != 9 {
		t.Fatalf("frame = %d", frames[1])
	}
	for i, event := range events {
		if event != uint32(i+1) {
			t.Fatalf("events = %v", events)
		}
	}
}
