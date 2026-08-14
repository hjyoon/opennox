package server

import (
	"reflect"
	"testing"
)

// This standard-library-only contract is also compiled as a standalone test
// package for every supported GOOS/GOARCH tuple.
func TestDeathBallCollidePortableContract4E9E90(t *testing.T) {
	var events []string
	deathBallCollide4E9E90(1, 2, 3, deathBallCollideHooks4E9E90[int, int, int, int]{
		loadClassByte: func(int) uint8 {
			events = append(events, "class")
			return 0
		},
		balanceFloat: func(string) float64 {
			events = append(events, "balance")
			return 10.5
		},
		floatToInt: func(float32) int32 {
			events = append(events, "round")
			return 10
		},
		findParent: func(int) int {
			events = append(events, "parent")
			return 4
		},
		targetDamage: func(int, int, int, int32, uint32) int32 {
			events = append(events, "damage")
			return 1
		},
	})
	want := []string{"class", "balance", "round", "parent", "damage"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
