package server

import (
	"reflect"
	"testing"
)

type findOwnerChainPlayerTestObject4EC580 struct {
	name  string
	class uint8
	owner *findOwnerChainPlayerTestObject4EC580
}

func TestFindOwnerChainPlayer4EC580NilDoesNotRead(t *testing.T) {
	hooks := findOwnerChainPlayerHooks4EC580[*findOwnerChainPlayerTestObject4EC580]{
		owner: func(*findOwnerChainPlayerTestObject4EC580) *findOwnerChainPlayerTestObject4EC580 {
			t.Fatal("nil object read owner")
			return nil
		},
		classLow: func(*findOwnerChainPlayerTestObject4EC580) uint8 {
			t.Fatal("nil object read class")
			return 0
		},
	}
	if got := findOwnerChainPlayer4EC580(nil, hooks); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
}

func TestFindOwnerChainPlayer4EC580ReadsOwnerBeforeClassAndSkipsTerminalClass(t *testing.T) {
	terminal := &findOwnerChainPlayerTestObject4EC580{name: "terminal", class: 4}
	middle := &findOwnerChainPlayerTestObject4EC580{name: "middle", owner: terminal}
	first := &findOwnerChainPlayerTestObject4EC580{name: "first", owner: middle}
	var events []string
	hooks := findOwnerChainPlayerHooks4EC580[*findOwnerChainPlayerTestObject4EC580]{
		owner: func(obj *findOwnerChainPlayerTestObject4EC580) *findOwnerChainPlayerTestObject4EC580 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
		classLow: func(obj *findOwnerChainPlayerTestObject4EC580) uint8 {
			events = append(events, "class:"+obj.name)
			return obj.class
		},
	}
	if got := findOwnerChainPlayer4EC580(first, hooks); got != terminal {
		t.Fatalf("result = %p, want terminal %p", got, terminal)
	}
	want := []string{"owner:first", "class:first", "owner:middle", "class:middle", "owner:terminal"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestFindOwnerChainPlayer4EC580PlayerWithOwnerReturnsSelf(t *testing.T) {
	owner := &findOwnerChainPlayerTestObject4EC580{name: "owner"}
	player := &findOwnerChainPlayerTestObject4EC580{name: "player", class: 4, owner: owner}
	var events []string
	got := findOwnerChainPlayer4EC580(player, findOwnerChainPlayerHooks4EC580[*findOwnerChainPlayerTestObject4EC580]{
		owner: func(obj *findOwnerChainPlayerTestObject4EC580) *findOwnerChainPlayerTestObject4EC580 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
		classLow: func(obj *findOwnerChainPlayerTestObject4EC580) uint8 {
			events = append(events, "class:"+obj.name)
			return obj.class
		},
	})
	if got != player {
		t.Fatalf("result = %p, want player %p", got, player)
	}
	if want := []string{"owner:player", "class:player"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestFindOwnerChainPlayer4EC580UsesOwnerCachedBeforeClassCallback(t *testing.T) {
	original := &findOwnerChainPlayerTestObject4EC580{name: "original"}
	replacement := &findOwnerChainPlayerTestObject4EC580{name: "replacement"}
	first := &findOwnerChainPlayerTestObject4EC580{name: "first", owner: original}
	got := findOwnerChainPlayer4EC580(first, findOwnerChainPlayerHooks4EC580[*findOwnerChainPlayerTestObject4EC580]{
		owner: func(obj *findOwnerChainPlayerTestObject4EC580) *findOwnerChainPlayerTestObject4EC580 {
			return obj.owner
		},
		classLow: func(obj *findOwnerChainPlayerTestObject4EC580) uint8 {
			if obj == first {
				first.owner = replacement
			}
			return obj.class
		},
	})
	if got != original {
		t.Fatalf("result = %p, want cached original owner %p", got, original)
	}
}
