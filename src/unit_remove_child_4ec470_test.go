package opennox

import (
	"reflect"
	"testing"
)

type unitRemoveChildTestObject4EC470 struct {
	name       string
	owner      *unitRemoveChildTestObject4EC470
	nextOwned  *unitRemoveChildTestObject4EC470
	firstOwned *unitRemoveChildTestObject4EC470
}

func TestUnitRemoveChildContract4EC470NilAndEmptyParent(t *testing.T) {
	parent := &unitRemoveChildTestObject4EC470{name: "parent"}
	events := make([]string, 0, 2)
	hooks := unitRemoveChildHooks4EC470[*unitRemoveChildTestObject4EC470]{
		loadFirstOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			events = append(events, "first:"+obj.name)
			return obj.firstOwned
		},
		loadNextOwned: func(*unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			t.Fatal("empty parent loaded a child successor")
			return nil
		},
		storeOwner: func(*unitRemoveChildTestObject4EC470, *unitRemoveChildTestObject4EC470) {
			t.Fatal("empty parent stored a child owner")
		},
		storeNextOwned: func(*unitRemoveChildTestObject4EC470, *unitRemoveChildTestObject4EC470) {
			t.Fatal("empty parent stored a child successor")
		},
		storeFirstOwned: func(obj, first *unitRemoveChildTestObject4EC470) {
			events = append(events, "store-first:"+obj.name)
			obj.firstOwned = first
		},
	}

	unitRemoveChildContract4EC470(nil, hooks)
	if len(events) != 0 {
		t.Fatalf("nil parent events = %v", events)
	}
	unitRemoveChildContract4EC470(parent, hooks)
	want := []string{"first:parent", "store-first:parent"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("empty parent events = %v, want %v", events, want)
	}
}

func TestUnitRemoveChildContract4EC470TraversalAndStoreOrder(t *testing.T) {
	foreignOwner := &unitRemoveChildTestObject4EC470{name: "foreign-owner"}
	first := &unitRemoveChildTestObject4EC470{name: "first", owner: foreignOwner}
	second := &unitRemoveChildTestObject4EC470{name: "second"}
	parent := &unitRemoveChildTestObject4EC470{name: "parent", firstOwned: first}
	first.nextOwned = second
	events := make([]string, 0, 8)
	hooks := unitRemoveChildHooks4EC470[*unitRemoveChildTestObject4EC470]{
		loadFirstOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			events = append(events, "first:"+obj.name)
			return obj.firstOwned
		},
		loadNextOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			events = append(events, "next:"+obj.name)
			return obj.nextOwned
		},
		storeOwner: func(obj, owner *unitRemoveChildTestObject4EC470) {
			events = append(events, "store-owner:"+obj.name)
			obj.owner = owner
		},
		storeNextOwned: func(obj, next *unitRemoveChildTestObject4EC470) {
			events = append(events, "store-next:"+obj.name)
			obj.nextOwned = next
		},
		storeFirstOwned: func(obj, first *unitRemoveChildTestObject4EC470) {
			events = append(events, "store-first:"+obj.name)
			obj.firstOwned = first
		},
	}

	unitRemoveChildContract4EC470(parent, hooks)
	want := []string{
		"first:parent",
		"next:first", "store-owner:first", "store-next:first",
		"next:second", "store-owner:second", "store-next:second",
		"store-first:parent",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if parent.firstOwned != nil || first.owner != nil || first.nextOwned != nil || second.owner != nil || second.nextOwned != nil {
		t.Fatalf("links were not cleared: parent=%p first=(%p,%p) second=(%p,%p)",
			parent.firstOwned, first.owner, first.nextOwned, second.owner, second.nextOwned)
	}
}

func TestUnitRemoveChildContract4EC470CachesNextBeforeStores(t *testing.T) {
	cached := &unitRemoveChildTestObject4EC470{name: "cached"}
	replacement := &unitRemoveChildTestObject4EC470{name: "replacement"}
	first := &unitRemoveChildTestObject4EC470{name: "first", nextOwned: cached}
	parent := &unitRemoveChildTestObject4EC470{name: "parent", firstOwned: first}
	visited := make([]string, 0, 2)
	hooks := unitRemoveChildHooks4EC470[*unitRemoveChildTestObject4EC470]{
		loadFirstOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			return obj.firstOwned
		},
		loadNextOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			visited = append(visited, obj.name)
			return obj.nextOwned
		},
		storeOwner: func(obj, owner *unitRemoveChildTestObject4EC470) {
			obj.owner = owner
			if obj == first {
				obj.nextOwned = replacement
			}
		},
		storeNextOwned: func(obj, next *unitRemoveChildTestObject4EC470) {
			obj.nextOwned = next
		},
		storeFirstOwned: func(obj, first *unitRemoveChildTestObject4EC470) {
			obj.firstOwned = first
		},
	}

	unitRemoveChildContract4EC470(parent, hooks)
	if want := []string{"first", "cached"}; !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %v, want cached successor %v", visited, want)
	}
	if replacement.nextOwned != nil || replacement.owner != nil {
		t.Fatal("replacement successor was unexpectedly visited")
	}
}

func TestUnitRemoveChildContract4EC470SelfCycleRevisitsClearedNode(t *testing.T) {
	child := &unitRemoveChildTestObject4EC470{name: "child"}
	child.nextOwned = child
	parent := &unitRemoveChildTestObject4EC470{name: "parent", firstOwned: child}
	nextLoads := 0
	ownerStores := 0
	nextStores := 0
	hooks := unitRemoveChildHooks4EC470[*unitRemoveChildTestObject4EC470]{
		loadFirstOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			return obj.firstOwned
		},
		loadNextOwned: func(obj *unitRemoveChildTestObject4EC470) *unitRemoveChildTestObject4EC470 {
			nextLoads++
			return obj.nextOwned
		},
		storeOwner: func(obj, owner *unitRemoveChildTestObject4EC470) {
			ownerStores++
			obj.owner = owner
		},
		storeNextOwned: func(obj, next *unitRemoveChildTestObject4EC470) {
			nextStores++
			obj.nextOwned = next
		},
		storeFirstOwned: func(obj, first *unitRemoveChildTestObject4EC470) {
			obj.firstOwned = first
		},
	}

	unitRemoveChildContract4EC470(parent, hooks)
	if nextLoads != 2 || ownerStores != 2 || nextStores != 2 {
		t.Fatalf("self-cycle counts = next loads %d, owner stores %d, next stores %d", nextLoads, ownerStores, nextStores)
	}
	if parent.firstOwned != nil || child.owner != nil || child.nextOwned != nil {
		t.Fatalf("self-cycle links = parent %p owner %p next %p", parent.firstOwned, child.owner, child.nextOwned)
	}
}
