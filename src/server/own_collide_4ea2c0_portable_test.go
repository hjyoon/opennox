package server

import "testing"

func TestOwnCollidePortableContract4EA2C0(t *testing.T) {
	type state struct {
		owner int
		frame uint32
	}
	states := map[int]*state{1: {}}
	events := make([]uint32, 0, 5)
	ownCollide4EA2C0(1, 2, ownCollideHooks4EA2C0[int]{
		loadTargetClass: func(target int) uint32 {
			events = append(events, 1)
			if target != 2 {
				t.Fatalf("target = %d", target)
			}
			return 4
		},
		loadSourceOwner: func(source int) int {
			events = append(events, 2)
			return states[source].owner
		},
		loadFrame: func() uint32 {
			events = append(events, 3)
			return 0xfedcba98
		},
		storeSourceFrame: func(source int, frame uint32) {
			events = append(events, 4)
			states[source].frame = frame
		},
		setOwner: func(owner, source int) {
			events = append(events, 5)
			states[source].owner = owner
		},
	})
	if states[1].owner != 2 || states[1].frame != 0xfedcba98 {
		t.Fatalf("state = %+v", states[1])
	}
	for i, event := range events {
		if event != uint32(i+1) {
			t.Fatalf("events = %v", events)
		}
	}
}
