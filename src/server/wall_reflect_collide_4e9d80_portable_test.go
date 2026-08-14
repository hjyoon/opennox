package server

import (
	"reflect"
	"testing"
)

// This standard-library-only contract is also compiled as a standalone test
// package for every supported GOOS/GOARCH tuple.
func TestWallReflectCollidePortableContract4E9D80(t *testing.T) {
	var events []string
	wallReflectCollide4E9D80(1, 2, 3, wallReflectCollideHooks4E9D80[int, int, int, int]{
		loadCollideData:   func(int) int { events = append(events, "data"); return 7 },
		sameTeam:          func(int, int) int32 { events = append(events, "team"); return 0 },
		gameFlagsCheck:    func(uint32) int32 { events = append(events, "quest"); return 1 },
		loadCollide:       func(int) int { events = append(events, "collide"); return 9 },
		yellowStarCollide: 9,
		loadDamage:        func(int) int32 { events = append(events, "damage"); return 7 },
		findParent:        func(int) int { events = append(events, "parent"); return 4 },
		targetDamage: func(int, int, int, int32, uint32) int32 {
			events = append(events, "target-damage")
			return 1
		},
		delayedDelete: func(int) { events = append(events, "delete") },
	})
	want := []string{"data", "team", "quest", "collide", "damage", "parent", "target-damage", "delete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
