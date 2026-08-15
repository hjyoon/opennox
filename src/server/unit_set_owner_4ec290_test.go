package server

import (
	"reflect"
	"testing"
)

type unitSetOwnerTestObject4EC290 struct {
	name       string
	flags      uint32
	class      uint32
	owner      *unitSetOwnerTestObject4EC290
	nextOwned  *unitSetOwnerTestObject4EC290
	firstOwned *unitSetOwnerTestObject4EC290
}

func defaultUnitSetOwnerHooks4EC290() unitSetOwnerHooks4EC290[*unitSetOwnerTestObject4EC290] {
	return unitSetOwnerHooks4EC290[*unitSetOwnerTestObject4EC290]{
		clearOwner: func(obj *unitSetOwnerTestObject4EC290) {
			obj.owner = nil
		},
		loadFlags: func(obj *unitSetOwnerTestObject4EC290) uint32 {
			return obj.flags
		},
		loadOwner: func(obj *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
			return obj.owner
		},
		loadFirstOwned: func(obj *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
			return obj.firstOwned
		},
		storeNextOwned: func(obj, next *unitSetOwnerTestObject4EC290) {
			obj.nextOwned = next
		},
		storeFirstOwned: func(owner, first *unitSetOwnerTestObject4EC290) {
			owner.firstOwned = first
		},
		storeOwner: func(obj, owner *unitSetOwnerTestObject4EC290) {
			obj.owner = owner
		},
		loadClass: func(obj *unitSetOwnerTestObject4EC290) uint32 {
			return obj.class
		},
		resetMonster:   func(*unitSetOwnerTestObject4EC290) {},
		markUnitUpdate: func(*unitSetOwnerTestObject4EC290) {},
	}
}

func TestUnitSetOwner4EC290NilObjectDoesNothing(t *testing.T) {
	hooks := defaultUnitSetOwnerHooks4EC290()
	hooks.clearOwner = func(*unitSetOwnerTestObject4EC290) {
		t.Fatal("nil object called clearOwner")
	}
	unitSetOwner4EC290(&unitSetOwnerTestObject4EC290{name: "owner"}, nil, hooks)
}

func TestUnitSetOwner4EC290InsertionAndLiveSecondClassLoad(t *testing.T) {
	oldFirst := &unitSetOwnerTestObject4EC290{name: "old-first"}
	oldOwner := &unitSetOwnerTestObject4EC290{name: "old-owner"}
	owner := &unitSetOwnerTestObject4EC290{name: "owner", firstOwned: oldFirst}
	obj := &unitSetOwnerTestObject4EC290{name: "object", class: 0x80000002, owner: oldOwner}
	events := make([]string, 0, 10)
	classLoads := 0
	hooks := defaultUnitSetOwnerHooks4EC290()
	hooks.clearOwner = func(got *unitSetOwnerTestObject4EC290) {
		events = append(events, "clear")
		if got != obj {
			t.Fatalf("clear object = %p", got)
		}
		got.owner = nil
	}
	hooks.loadFlags = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "flags:"+got.name)
		return got.flags
	}
	hooks.loadFirstOwned = func(got *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
		events = append(events, "first:"+got.name)
		return got.firstOwned
	}
	hooks.storeNextOwned = func(got, next *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-next")
		if got != obj || next != oldFirst {
			t.Fatalf("next store = %p/%p", got, next)
		}
		got.nextOwned = next
	}
	hooks.storeFirstOwned = func(gotOwner, first *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-first")
		if gotOwner != owner || first != obj {
			t.Fatalf("first store = %p/%p", gotOwner, first)
		}
		gotOwner.firstOwned = first
	}
	hooks.storeOwner = func(got, gotOwner *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-owner")
		if got != obj || gotOwner != owner {
			t.Fatalf("owner store = %p/%p", got, gotOwner)
		}
		got.owner = gotOwner
	}
	hooks.loadClass = func(got *unitSetOwnerTestObject4EC290) uint32 {
		classLoads++
		events = append(events, "class")
		return got.class
	}
	hooks.resetMonster = func(got *unitSetOwnerTestObject4EC290) {
		events = append(events, "reset")
		got.class = 0
	}
	hooks.markUnitUpdate = func(*unitSetOwnerTestObject4EC290) {
		t.Fatal("second class load reused the pre-reset Monster class")
	}

	unitSetOwner4EC290(owner, obj, hooks)
	wantEvents := []string{
		"clear", "flags:owner", "first:owner", "store-next", "store-first",
		"store-owner", "class", "reset", "class",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if classLoads != 2 {
		t.Fatalf("class loads = %d, want 2", classLoads)
	}
	if obj.owner != owner || obj.nextOwned != oldFirst || owner.firstOwned != obj {
		t.Fatalf("ownership = owner %p next %p first %p", obj.owner, obj.nextOwned, owner.firstOwned)
	}
}

func TestUnitSetOwner4EC290DestroyedOwnerTraversal(t *testing.T) {
	terminal := &unitSetOwnerTestObject4EC290{name: "terminal"}
	second := &unitSetOwnerTestObject4EC290{name: "second", flags: 0x120, owner: terminal}
	first := &unitSetOwnerTestObject4EC290{name: "first", flags: 0x20, owner: second}
	obj := &unitSetOwnerTestObject4EC290{name: "object"}
	events := make([]string, 0, 10)
	hooks := defaultUnitSetOwnerHooks4EC290()
	hooks.clearOwner = func(*unitSetOwnerTestObject4EC290) {
		events = append(events, "clear")
	}
	hooks.loadFlags = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "flags:"+got.name)
		return got.flags
	}
	hooks.loadOwner = func(got *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
		events = append(events, "owner:"+got.name)
		return got.owner
	}
	hooks.loadFirstOwned = func(got *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
		events = append(events, "first:"+got.name)
		return got.firstOwned
	}
	hooks.storeNextOwned = func(got, next *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-next")
		got.nextOwned = next
	}
	hooks.storeFirstOwned = func(gotOwner, firstOwned *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-first")
		gotOwner.firstOwned = firstOwned
	}
	hooks.storeOwner = func(got, gotOwner *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-owner:"+gotOwner.name)
		got.owner = gotOwner
	}
	hooks.loadClass = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "class")
		return got.class
	}

	unitSetOwner4EC290(first, obj, hooks)
	wantEvents := []string{
		"clear", "flags:first", "owner:first", "flags:second", "owner:second",
		"flags:terminal", "first:terminal", "store-next", "store-first",
		"store-owner:terminal", "class", "class",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if obj.owner != terminal || terminal.firstOwned != obj {
		t.Fatalf("resolved ownership = owner %p first %p", obj.owner, terminal.firstOwned)
	}
}

func TestUnitSetOwner4EC290DestroyedChainCanResolveToNil(t *testing.T) {
	destroyed := &unitSetOwnerTestObject4EC290{name: "destroyed", flags: 0x20}
	obj := &unitSetOwnerTestObject4EC290{name: "object", class: 0x04}
	events := make([]string, 0, 8)
	hooks := defaultUnitSetOwnerHooks4EC290()
	hooks.clearOwner = func(*unitSetOwnerTestObject4EC290) {
		events = append(events, "clear")
	}
	hooks.loadFlags = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "flags")
		return got.flags
	}
	hooks.loadOwner = func(got *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
		events = append(events, "owner")
		return got.owner
	}
	hooks.loadFirstOwned = func(*unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
		t.Fatal("nil resolved owner loaded an owned-list head")
		return nil
	}
	hooks.storeNextOwned = func(*unitSetOwnerTestObject4EC290, *unitSetOwnerTestObject4EC290) {
		t.Fatal("nil resolved owner stored an owned-list link")
	}
	hooks.storeFirstOwned = func(*unitSetOwnerTestObject4EC290, *unitSetOwnerTestObject4EC290) {
		t.Fatal("nil resolved owner stored an owned-list head")
	}
	hooks.storeOwner = func(got, gotOwner *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-owner")
		if gotOwner != nil {
			t.Fatalf("resolved owner = %p", gotOwner)
		}
		got.owner = gotOwner
	}
	hooks.loadClass = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "class")
		return got.class
	}
	hooks.markUnitUpdate = func(*unitSetOwnerTestObject4EC290) {
		events = append(events, "mark")
	}

	unitSetOwner4EC290(destroyed, obj, hooks)
	wantEvents := []string{"clear", "flags", "owner", "store-owner", "class", "class", "mark"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if obj.owner != nil {
		t.Fatalf("object owner = %p", obj.owner)
	}
}

func TestUnitSetOwner4EC290NilOwnerStillClearsAndUpdatesUnit(t *testing.T) {
	oldOwner := &unitSetOwnerTestObject4EC290{name: "old-owner"}
	obj := &unitSetOwnerTestObject4EC290{name: "object", class: 0x04, owner: oldOwner}
	events := make([]string, 0, 5)
	hooks := defaultUnitSetOwnerHooks4EC290()
	hooks.clearOwner = func(got *unitSetOwnerTestObject4EC290) {
		events = append(events, "clear")
		got.owner = nil
	}
	hooks.loadFlags = func(*unitSetOwnerTestObject4EC290) uint32 {
		t.Fatal("nil owner loaded flags")
		return 0
	}
	hooks.storeOwner = func(got, gotOwner *unitSetOwnerTestObject4EC290) {
		events = append(events, "store-owner")
		got.owner = gotOwner
	}
	hooks.loadClass = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "class")
		return got.class
	}
	hooks.markUnitUpdate = func(*unitSetOwnerTestObject4EC290) {
		events = append(events, "mark")
	}

	unitSetOwner4EC290(nil, obj, hooks)
	if !reflect.DeepEqual(events, []string{"clear", "store-owner", "class", "class", "mark"}) {
		t.Fatalf("events = %v", events)
	}
	if obj.owner != nil {
		t.Fatalf("object owner = %p", obj.owner)
	}
}

func TestUnitSetOwner4EC290ClearPrecedesLiveOwnerTraversal(t *testing.T) {
	redirect := &unitSetOwnerTestObject4EC290{name: "redirect"}
	owner := &unitSetOwnerTestObject4EC290{name: "owner"}
	obj := &unitSetOwnerTestObject4EC290{name: "object"}
	events := make([]string, 0, 4)
	hooks := defaultUnitSetOwnerHooks4EC290()
	hooks.clearOwner = func(*unitSetOwnerTestObject4EC290) {
		events = append(events, "clear")
		owner.flags = 0x20
		owner.owner = redirect
	}
	hooks.loadFlags = func(got *unitSetOwnerTestObject4EC290) uint32 {
		events = append(events, "flags:"+got.name)
		return got.flags
	}
	hooks.loadOwner = func(got *unitSetOwnerTestObject4EC290) *unitSetOwnerTestObject4EC290 {
		events = append(events, "owner:"+got.name)
		return got.owner
	}

	unitSetOwner4EC290(owner, obj, hooks)
	if len(events) < 4 || !reflect.DeepEqual(events[:4], []string{"clear", "flags:owner", "owner:owner", "flags:redirect"}) {
		t.Fatalf("traversal prefix = %v", events)
	}
	if obj.owner != redirect || redirect.firstOwned != obj {
		t.Fatalf("live owner = %p, redirect first = %p", obj.owner, redirect.firstOwned)
	}
}
