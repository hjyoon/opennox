package server

import "testing"

func TestAudioEventCollidePortableContract4EAAD0(t *testing.T) {
	frames := map[int]uint32{1: 0}
	sounds := map[int]uint32{1: 417}
	events := make([]uint32, 0, 7)
	audioEventCollide4EAAD0(1, 2, 3, audioEventCollideHooks4EAAD0[int, int]{
		classLow: func(target int) uint8 {
			events = append(events, 1)
			return audioEventCollidePlayerClass4EAAD0
		},
		loadFrame: func() uint32 {
			events = append(events, 2)
			return 31
		},
		loadLastFrame: func(source int) uint32 {
			events = append(events, 3)
			return frames[source]
		},
		storeFrame: func(source int, frame uint32) {
			events = append(events, 4)
			frames[source] = frame
		},
		loadCollideData: func(source int) int {
			events = append(events, 5)
			return source
		},
		loadSound: func(data int) uint32 {
			events = append(events, 6)
			return sounds[data]
		},
		audio: func(id uint32, source int) {
			events = append(events, 7)
			if id != 417 || source != 1 {
				t.Fatalf("audio = %d/%d", id, source)
			}
		},
	})
	if frames[1] != 31 {
		t.Fatalf("frame = %d", frames[1])
	}
	for i, event := range events {
		if event != uint32(i+1) {
			t.Fatalf("events = %v", events)
		}
	}
}
