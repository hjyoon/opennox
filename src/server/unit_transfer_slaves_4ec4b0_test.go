package server

import (
	"reflect"
	"testing"
)

type unitTransferSlavesTestObject4EC4B0 struct {
	name       string
	owner      *unitTransferSlavesTestObject4EC4B0
	nextOwned  *unitTransferSlavesTestObject4EC4B0
	firstOwned *unitTransferSlavesTestObject4EC4B0
}

func TestUnitTransferSlaves4EC4B0NilAndEmptySource(t *testing.T) {
	source := &unitTransferSlavesTestObject4EC4B0{name: "source"}
	events := make([]string, 0, 1)
	hooks := unitTransferSlavesHooks4EC4B0[*unitTransferSlavesTestObject4EC4B0]{
		loadFirstOwned: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			events = append(events, "first:"+obj.name)
			return obj.firstOwned
		},
		loadNextOwned: func(*unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			t.Fatal("empty source loaded a child successor")
			return nil
		},
		loadOwner: func(*unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			t.Fatal("empty source loaded its owner")
			return nil
		},
		setOwner: func(*unitTransferSlavesTestObject4EC4B0, *unitTransferSlavesTestObject4EC4B0) {
			t.Fatal("empty source reassigned a child")
		},
	}

	unitTransferSlaves4EC4B0(nil, hooks)
	if len(events) != 0 {
		t.Fatalf("nil source events = %v", events)
	}
	unitTransferSlaves4EC4B0(source, hooks)
	if want := []string{"first:source"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("empty source events = %v, want %v", events, want)
	}
}

func TestUnitTransferSlaves4EC4B0CachesNextAndReloadsSourceOwner(t *testing.T) {
	ownerA := &unitTransferSlavesTestObject4EC4B0{name: "owner-a"}
	ownerB := &unitTransferSlavesTestObject4EC4B0{name: "owner-b"}
	replacement := &unitTransferSlavesTestObject4EC4B0{name: "replacement"}
	first := &unitTransferSlavesTestObject4EC4B0{name: "first"}
	second := &unitTransferSlavesTestObject4EC4B0{name: "second"}
	third := &unitTransferSlavesTestObject4EC4B0{name: "third"}
	source := &unitTransferSlavesTestObject4EC4B0{name: "source", owner: ownerA, firstOwned: first}
	first.nextOwned = second
	second.nextOwned = third
	events := make([]string, 0, 10)
	ownerName := func(obj *unitTransferSlavesTestObject4EC4B0) string {
		if obj == nil {
			return "nil"
		}
		return obj.name
	}
	hooks := unitTransferSlavesHooks4EC4B0[*unitTransferSlavesTestObject4EC4B0]{
		loadFirstOwned: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			events = append(events, "first:"+obj.name)
			return obj.firstOwned
		},
		loadNextOwned: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			events = append(events, "next:"+obj.name)
			return obj.nextOwned
		},
		loadOwner: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
		setOwner: func(owner, child *unitTransferSlavesTestObject4EC4B0) {
			events = append(events, "set:"+ownerName(owner)+":"+child.name)
			child.owner = owner
			switch child {
			case first:
				child.nextOwned = replacement
				source.owner = ownerB
			case second:
				source.owner = nil
			}
		},
	}

	unitTransferSlaves4EC4B0(source, hooks)
	want := []string{
		"first:source",
		"next:first", "owner:source", "set:owner-a:first",
		"next:second", "owner:source", "set:owner-b:second",
		"next:third", "owner:source", "set:nil:third",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.firstOwned != first {
		t.Fatalf("source first-owned was explicitly changed to %p", source.firstOwned)
	}
	if replacement.owner != nil {
		t.Fatal("replacement successor was unexpectedly visited")
	}
	if first.owner != ownerA || second.owner != ownerB || third.owner != nil {
		t.Fatalf("owners = first %p second %p third %p", first.owner, second.owner, third.owner)
	}
}

func TestUnitTransferSlaves4EC4B0LoadsEachSuccessorOnArrival(t *testing.T) {
	owner := &unitTransferSlavesTestObject4EC4B0{name: "owner"}
	first := &unitTransferSlavesTestObject4EC4B0{name: "first"}
	second := &unitTransferSlavesTestObject4EC4B0{name: "second"}
	originalThird := &unitTransferSlavesTestObject4EC4B0{name: "original-third"}
	replacement := &unitTransferSlavesTestObject4EC4B0{name: "replacement"}
	source := &unitTransferSlavesTestObject4EC4B0{name: "source", owner: owner, firstOwned: first}
	first.nextOwned = second
	second.nextOwned = originalThird
	visited := make([]string, 0, 3)
	unitTransferSlaves4EC4B0(source, unitTransferSlavesHooks4EC4B0[*unitTransferSlavesTestObject4EC4B0]{
		loadFirstOwned: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			return obj.firstOwned
		},
		loadNextOwned: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			visited = append(visited, obj.name)
			return obj.nextOwned
		},
		loadOwner: func(obj *unitTransferSlavesTestObject4EC4B0) *unitTransferSlavesTestObject4EC4B0 {
			return obj.owner
		},
		setOwner: func(_ *unitTransferSlavesTestObject4EC4B0, child *unitTransferSlavesTestObject4EC4B0) {
			if child == first {
				second.nextOwned = replacement
			}
		},
	})

	if want := []string{"first", "second", "replacement"}; !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %v, want %v", visited, want)
	}
}
