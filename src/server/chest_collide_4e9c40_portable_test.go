package server

import (
	"reflect"
	"testing"
)

// This standard-library-only contract is also compiled as a standalone test
// package for every supported GOOS/GOARCH tuple.
func TestChestCollidePortableContract4E9C40(t *testing.T) {
	var events []string
	chestCollide4E9C40(1, 2, 123, chestCollideHooks4E9C40[int, int]{
		loadClassLow:   func(int) uint8 { events = append(events, "class"); return chestCollidePlayerClass4E9C40 },
		loadFlags:      func(int) uint32 { events = append(events, "flags"); return 0 },
		gameFlagsCheck: func(uint32) int32 { events = append(events, "quest"); return 0 },
		loadDeath:      func(int) int { events = append(events, "death"); return 7 },
		callDeath:      func(int, int) { events = append(events, "call") },
		chestOpen:      func(int, int) { events = append(events, "open") },
		dropAllItems:   func(int) { events = append(events, "drop") },
	})
	want := []string{"class", "flags", "quest", "death", "call", "open", "drop"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
