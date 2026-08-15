package server

import "testing"

func TestUnitClearOwnerPortable4EC300(t *testing.T) {
	owner := 1
	obj := 2
	tail := 3
	owners := map[int]int{obj: owner}
	classes := map[int]uint32{owner: 0, obj: unitClearOwnerMonsterClass4EC300}
	first := map[int]int{owner: obj}
	next := map[int]int{obj: tail}
	events := make([]string, 0, 2)
	hooks := unitClearOwnerHooks4EC300[int, int, int]{
		loadOwner:       func(v int) int { return owners[v] },
		loadClass:       func(v int) uint32 { return classes[v] },
		isMonitored:     func(int, int) bool { return false },
		loadSubClass:    func(int) uint32 { return 0 },
		loadPlayerData:  func(int) int { return 0 },
		storeSubClass:   func(int, uint32) {},
		loadPlayer:      func(int) int { return 0 },
		loadPlayerIndex: func(int) uint8 { return 0 },
		netFxShield:     func(uint8, int) {},
		unmarkMinimap:   func(uint8, int, uint32) {},
		loadFirstOwned:  func(v int) int { return first[v] },
		loadNextOwned:   func(v int) int { return next[v] },
		storeNextOwned:  func(v, n int) { next[v] = n },
		storeFirstOwned: func(v, n int) { first[v] = n },
		storeOwner:      func(v, o int) { owners[v] = o },
		resetMonster: func(int) {
			events = append(events, "reset")
			classes[obj] = unitClearOwnerPlayerClass4EC300
		},
		markUnitUpdate: func(int) { events = append(events, "mark") },
	}
	unitClearOwner4EC300(obj, hooks)
	if owners[obj] != 0 || first[owner] != tail || next[obj] != tail {
		t.Fatalf("portable ownership = owner %d first %d next %d", owners[obj], first[owner], next[obj])
	}
	if len(events) != 2 || events[0] != "reset" || events[1] != "mark" {
		t.Fatalf("portable events = %v", events)
	}
}
