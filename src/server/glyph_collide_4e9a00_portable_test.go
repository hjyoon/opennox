package server

import (
	"reflect"
	"testing"
)

// This standard-library-only contract is also compiled as a standalone test
// package for every supported GOOS/GOARCH tuple.
func TestGlyphCollidePortableContract4E9A00(t *testing.T) {
	var events []string
	gate := glyphCollideGate4E9A30(11, 22, glyphCollideGateHooks4E9A30[int]{
		gameFlag: func(flag uint32) int32 {
			events = append(events, "game")
			return 0
		},
		unitsOnSameTeam: func(source, target int) int32 {
			events = append(events, "same")
			return 0
		},
		abilityActive: func(target int, ability int32) int32 {
			events = append(events, "ability")
			return 0
		},
	})
	glyphCollide4E9A00(11, 22, 99, glyphCollideHooks4E9A00[int]{
		allowed: func(source, target int) int32 {
			events = append(events, "allowed")
			return gate
		},
		trigger: func(source, target int) {
			events = append(events, "trigger")
		},
	})
	want := []string{"game", "same", "game", "ability", "allowed", "trigger"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
