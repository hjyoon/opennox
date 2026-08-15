package server

import "testing"

func TestUnitSetOwnerPortableContract4EC290(t *testing.T) {
	type state struct {
		flags uint32
		class uint32
		owner int
		next  int
		first int
	}
	states := map[int]*state{
		1: {class: unitSetOwnerMonsterClass4EC290},
		2: {flags: unitSetOwnerDestroyedFlag4EC290, owner: 3},
		3: {first: 4},
		4: {},
	}
	events := make([]uint32, 0, 12)
	unitSetOwner4EC290(2, 1, unitSetOwnerHooks4EC290[int]{
		clearOwner: func(obj int) {
			events = append(events, 1)
			states[obj].owner = 0
		},
		loadFlags: func(obj int) uint32 {
			events = append(events, uint32(len(events)+1))
			return states[obj].flags
		},
		loadOwner: func(obj int) int {
			events = append(events, uint32(len(events)+1))
			return states[obj].owner
		},
		loadFirstOwned: func(owner int) int {
			events = append(events, uint32(len(events)+1))
			return states[owner].first
		},
		storeNextOwned: func(obj, next int) {
			events = append(events, uint32(len(events)+1))
			states[obj].next = next
		},
		storeFirstOwned: func(owner, first int) {
			events = append(events, uint32(len(events)+1))
			states[owner].first = first
		},
		storeOwner: func(obj, owner int) {
			events = append(events, uint32(len(events)+1))
			states[obj].owner = owner
		},
		loadClass: func(obj int) uint32 {
			events = append(events, uint32(len(events)+1))
			return states[obj].class
		},
		resetMonster: func(obj int) {
			events = append(events, uint32(len(events)+1))
			states[obj].class = 4
		},
		markUnitUpdate: func(int) {
			events = append(events, uint32(len(events)+1))
		},
	})
	if states[1].owner != 3 || states[1].next != 4 || states[3].first != 1 {
		t.Fatalf("ownership = owner %d next %d first %d", states[1].owner, states[1].next, states[3].first)
	}
	if len(events) != 12 {
		t.Fatalf("events = %v", events)
	}
	for i, event := range events {
		if event != uint32(i+1) {
			t.Fatalf("events = %v", events)
		}
	}
}
