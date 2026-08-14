package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGlyphCollide4E9A00RejectsOnlyZeroAndLeavesCollisionUnread(t *testing.T) {
	for _, allowed := range []int32{1, -1, 0x40000000} {
		t.Run(fmt.Sprintf("%d", allowed), func(t *testing.T) {
			var events []string
			glyphCollide4E9A00(11, 22, struct{ unread int }{9}, glyphCollideHooks4E9A00[int]{
				allowed: func(source, target int) int32 {
					events = append(events, "allowed")
					if source != 11 || target != 22 {
						t.Fatalf("allowed args = (%d, %d), want (11, 22)", source, target)
					}
					return allowed
				},
				trigger: func(source, target int) {
					events = append(events, "trigger")
					if source != 11 || target != 22 {
						t.Fatalf("trigger args = (%d, %d), want (11, 22)", source, target)
					}
				},
			})
			if !reflect.DeepEqual(events, []string{"allowed", "trigger"}) {
				t.Fatalf("events = %#v, want [allowed trigger]", events)
			}
		})
	}
}

func TestGlyphCollide4E9A00ZeroGateSkipsTrigger(t *testing.T) {
	var events []string
	glyphCollide4E9A00(11, 22, nil, glyphCollideHooks4E9A00[int]{
		allowed: func(source, target int) int32 {
			events = append(events, "allowed")
			return 0
		},
		trigger: func(int, int) {
			t.Fatal("zero eligibility triggered the glyph")
		},
	})
	if !reflect.DeepEqual(events, []string{"allowed"}) {
		t.Fatalf("events = %#v, want [allowed]", events)
	}
}
