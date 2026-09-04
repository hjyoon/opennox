package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

func TestWindowHiddenAncestor46C2A0NilDoesNotLoad(t *testing.T) {
	got := windowHiddenAncestor46C2A0(uint64(0), windowHiddenAncestorHooks46C2A0[uint64]{
		loadFlagsLowByte: func(uint64) uint8 {
			t.Fatal("nil window must not load flags")
			return 0
		},
		loadParent: func(uint64) uint64 {
			t.Fatal("nil window must not load parent")
			return 0
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
}

func TestWindowHiddenAncestor46C2A0HiddenSelfStopsBeforeParent(t *testing.T) {
	const win = uint64(0x100001234)
	var events []string
	got := windowHiddenAncestor46C2A0(win, windowHiddenAncestorHooks46C2A0[uint64]{
		loadFlagsLowByte: func(value uint64) uint8 {
			events = append(events, fmt.Sprintf("flags:%#x", value))
			return 0x90
		},
		loadParent: func(uint64) uint64 {
			t.Fatal("hidden self must stop before parent load")
			return 0
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{"flags:0x100001234"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestWindowHiddenAncestor46C2A0UsesLiveParentChain(t *testing.T) {
	const (
		win         = uint64(0x100001234)
		parent      = uint64(0x200002345)
		staleParent = uint64(0x300003456)
		liveParent  = uint64(0x400004567)
	)
	parents := map[uint64]uint64{win: parent, parent: staleParent}
	flags := map[uint64]uint8{liveParent: 0x10}
	var events []string
	got := windowHiddenAncestor46C2A0(win, windowHiddenAncestorHooks46C2A0[uint64]{
		loadFlagsLowByte: func(value uint64) uint8 {
			events = append(events, fmt.Sprintf("flags:%#x", value))
			if value == parent {
				parents[parent] = liveParent
			}
			return flags[value]
		},
		loadParent: func(value uint64) uint64 {
			events = append(events, fmt.Sprintf("parent:%#x", value))
			return parents[value]
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{
		"flags:0x100001234", "parent:0x100001234",
		"flags:0x200002345", "parent:0x200002345",
		"flags:0x400004567",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestWindowHiddenAncestor46C2A0VisibleChainEndsAtNil(t *testing.T) {
	const (
		win    = uint64(0x100001234)
		parent = uint64(0x200002345)
	)
	parents := map[uint64]uint64{win: parent}
	var events []string
	got := windowHiddenAncestor46C2A0(win, windowHiddenAncestorHooks46C2A0[uint64]{
		loadFlagsLowByte: func(value uint64) uint8 {
			events = append(events, fmt.Sprintf("flags:%#x", value))
			return 0x08
		},
		loadParent: func(value uint64) uint64 {
			events = append(events, fmt.Sprintf("parent:%#x", value))
			return parents[value]
		},
	})
	if got != 0 {
		t.Fatalf("result = %d, want canonical 0", got)
	}
	want := []string{
		"flags:0x100001234", "parent:0x100001234",
		"flags:0x200002345", "parent:0x200002345",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestWindowHiddenAncestor46C2A0FaultPrefixes(t *testing.T) {
	const (
		win     = uint64(0x100001234)
		parent1 = uint64(0x200002345)
		parent2 = uint64(0x300003456)
	)
	allEvents := []string{"flags-win", "parent-win", "flags-parent1", "parent-parent1", "flags-parent2", "parent-parent2"}
	stop := &struct{}{}

	for failAt := range allEvents {
		t.Run(fmt.Sprintf("fault-%02d-%s", failAt, allEvents[failAt]), func(t *testing.T) {
			events := make([]string, 0, len(allEvents))
			observe := func(event string) {
				if len(events) == failAt {
					panic(stop)
				}
				events = append(events, event)
			}
			parents := map[uint64]uint64{win: parent1, parent1: parent2}
			labels := map[uint64]string{win: "win", parent1: "parent1", parent2: "parent2"}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				windowHiddenAncestor46C2A0(win, windowHiddenAncestorHooks46C2A0[uint64]{
					loadFlagsLowByte: func(value uint64) uint8 {
						observe("flags-" + labels[value])
						return 0
					},
					loadParent: func(value uint64) uint64 {
						observe("parent-" + labels[value])
						return parents[value]
					},
				})
			}()
			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			if want := allEvents[:failAt]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %#v, want prefix %#v", events, want)
			}
		})
	}
}
