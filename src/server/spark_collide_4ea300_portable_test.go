package server

import "testing"

func TestSparkCollidePortableContract4EA300(t *testing.T) {
	var count uint8 = 0xff
	var timer uint16
	events := make([]uint32, 0, 7)
	sparkCollide4EA300(1, 2, 0, sparkCollideHooks4EA300[int, int, uint32]{
		loadUpdateData: func(source int) uint32 {
			events = append(events, 1)
			return 5
		},
		loadKind: func(data uint32) uint32 {
			events = append(events, 2)
			return data
		},
		wallReflect: func(int, int, int) {
			t.Fatal("webbing kind reached WallReflect")
		},
		audio: func(id uint32, source int) {
			events = append(events, 3)
			if id != 351 || source != 1 {
				t.Fatalf("audio = %d/%d", id, source)
			}
		},
		delayedDelete: func(source int) {
			events = append(events, 4)
		},
		loadSlowCount: func(target int) uint8 {
			events = append(events, 5)
			return count
		},
		loadClassLow: func(target int) uint8 {
			events = append(events, 6)
			return 4
		},
		storeSlowCount: func(target int, value uint8) {
			events = append(events, 7)
			count = value
		},
		storeSlowTimer: func(target int, value uint16) {
			events = append(events, 8)
			timer = value
		},
		priorityMessage: func(target int, message string) {
			events = append(events, 9)
		},
	})
	if count != 0 || timer != 1000 {
		t.Fatalf("slow state = %d/%d", count, timer)
	}
	for i, event := range events {
		if event != uint32(i+1) {
			t.Fatalf("events = %v", events)
		}
	}
}
