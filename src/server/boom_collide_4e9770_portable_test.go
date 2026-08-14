package server

import (
	"reflect"
	"testing"
)

// This small, standard-library-only test is also compiled as a standalone
// command-line package for every supported GOOS/GOARCH contract tuple.
func TestBoomCollidePortableContract4E9770(t *testing.T) {
	var events []string
	boomCollide4E9770(1, 2, 0, boomCollideHooks4E9770[int, int, int]{
		loadBalanceReady: func() uint32 {
			events = append(events, "ready")
			return 1
		},
		gameFlagsCheck: func(flag uint32) int32 {
			if flag != boomCollideQuestFlag4E9770 {
				t.Fatalf("game flag = %#x, want %#x", flag, boomCollideQuestFlag4E9770)
			}
			events = append(events, "flags")
			return 1
		},
		findParent: func(source int) int {
			events = append(events, "parent")
			return source
		},
		classLow: func(obj int) uint8 {
			events = append(events, "class")
			return boomCollidePlayerClassLow4E9770
		},
		isEnemy: func(parent, target int) int32 {
			events = append(events, "enemy")
			return 0
		},
	})
	if want := []string{"ready", "flags", "parent", "class", "class", "enemy"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}

	for _, tc := range []struct {
		x, y float32
		want int32
	}{
		{x: 1, y: 0, want: 0},
		{x: 0, y: 1, want: 64},
		{x: -1, y: 0, want: 128},
		{x: 0, y: -1, want: 192},
	} {
		if got := directionFromVector509ED0(tc.x, tc.y); got != tc.want {
			t.Fatalf("direction(%g,%g) = %d, want %d", tc.x, tc.y, got, tc.want)
		}
	}
}
