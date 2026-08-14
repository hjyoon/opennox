package server

import (
	"reflect"
	"testing"
)

// This standard-library-only contract is also compiled as a standalone test
// package for every supported GOOS/GOARCH tuple.
func TestDieCollidePortableContract4E99B0(t *testing.T) {
	var events []string
	var flags uint32 = 0x40000009
	dieCollide4E99B0(1, 2, 99, dieCollideHooks4E99B0[int, int]{
		unitsOnSameTeam: func(source, target int) int32 {
			events = append(events, "same")
			return 0
		},
		classLow: func(target int) uint8 {
			events = append(events, "class")
			return dieCollideUnitClassMask4E99B0
		},
		loadFlags: func(source int) uint32 {
			events = append(events, "flags")
			return flags
		},
		loadDeath: func(source int) int {
			events = append(events, "death")
			return 0
		},
		storeFlags: func(source int, value uint32) {
			events = append(events, "store")
			flags = value
		},
		delayedDelete: func(source int) {
			events = append(events, "delete")
		},
	})
	want := []string{"same", "class", "flags", "death", "store", "delete"}
	if !reflect.DeepEqual(events, want) || flags != 0x40008009 {
		t.Fatalf("events/flags = (%#v, %#x), want (%#v, 0x40008009)", events, flags, want)
	}
}
