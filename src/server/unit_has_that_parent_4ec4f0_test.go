package server

import (
	"reflect"
	"testing"
)

type unitHasThatParentTestObject4EC4F0 struct {
	name  string
	owner *unitHasThatParentTestObject4EC4F0
}

func TestUnitHasThatParent4EC4F0NilArgumentsDoNotReadOwner(t *testing.T) {
	obj := &unitHasThatParentTestObject4EC4F0{name: "object"}
	hooks := unitHasThatParentHooks4EC4F0[*unitHasThatParentTestObject4EC4F0]{
		loadOwner: func(*unitHasThatParentTestObject4EC4F0) *unitHasThatParentTestObject4EC4F0 {
			t.Fatal("nil argument path read an owner link")
			return nil
		},
	}
	if unitHasThatParent4EC4F0(nil, obj, hooks) {
		t.Fatal("nil object matched a non-nil owner")
	}
	if unitHasThatParent4EC4F0(obj, nil, hooks) {
		t.Fatal("non-nil object matched a nil owner")
	}
}

func TestUnitHasThatParent4EC4F0IncludesStartingObjectBeforeOwnerRead(t *testing.T) {
	obj := &unitHasThatParentTestObject4EC4F0{name: "object"}
	got := unitHasThatParent4EC4F0(obj, obj, unitHasThatParentHooks4EC4F0[*unitHasThatParentTestObject4EC4F0]{
		loadOwner: func(*unitHasThatParentTestObject4EC4F0) *unitHasThatParentTestObject4EC4F0 {
			t.Fatal("identity match read the matching object's owner")
			return nil
		},
	})
	if !got {
		t.Fatal("starting object did not match itself")
	}
}

func TestUnitHasThatParent4EC4F0ReadsEachOwnerUntilMatch(t *testing.T) {
	target := &unitHasThatParentTestObject4EC4F0{name: "target"}
	middle := &unitHasThatParentTestObject4EC4F0{name: "middle", owner: target}
	first := &unitHasThatParentTestObject4EC4F0{name: "first", owner: middle}
	var events []string
	got := unitHasThatParent4EC4F0(first, target, unitHasThatParentHooks4EC4F0[*unitHasThatParentTestObject4EC4F0]{
		loadOwner: func(obj *unitHasThatParentTestObject4EC4F0) *unitHasThatParentTestObject4EC4F0 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
	})
	if !got {
		t.Fatal("reachable target did not match")
	}
	if want := []string{"owner:first", "owner:middle"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitHasThatParent4EC4F0ReadsTerminalOwnerBeforeFailure(t *testing.T) {
	missing := &unitHasThatParentTestObject4EC4F0{name: "missing"}
	terminal := &unitHasThatParentTestObject4EC4F0{name: "terminal"}
	first := &unitHasThatParentTestObject4EC4F0{name: "first", owner: terminal}
	var events []string
	got := unitHasThatParent4EC4F0(first, missing, unitHasThatParentHooks4EC4F0[*unitHasThatParentTestObject4EC4F0]{
		loadOwner: func(obj *unitHasThatParentTestObject4EC4F0) *unitHasThatParentTestObject4EC4F0 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
	})
	if got {
		t.Fatal("unreachable target unexpectedly matched")
	}
	if want := []string{"owner:first", "owner:terminal"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitHasThatParent4EC4F0ReturnsBeforeFollowingReachableCycle(t *testing.T) {
	first := &unitHasThatParentTestObject4EC4F0{name: "first"}
	second := &unitHasThatParentTestObject4EC4F0{name: "second"}
	first.owner = second
	second.owner = first
	reads := 0
	got := unitHasThatParent4EC4F0(first, second, unitHasThatParentHooks4EC4F0[*unitHasThatParentTestObject4EC4F0]{
		loadOwner: func(obj *unitHasThatParentTestObject4EC4F0) *unitHasThatParentTestObject4EC4F0 {
			reads++
			return obj.owner
		},
	})
	if !got || reads != 1 {
		t.Fatalf("cycle result = %v, owner reads = %d, want true and 1", got, reads)
	}
}
