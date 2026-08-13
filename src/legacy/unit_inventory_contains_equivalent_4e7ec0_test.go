package legacy

import (
	"reflect"
	"testing"
)

type inventoryContainsEquivalentObject4E7EC0 struct {
	name  string
	match bool
	first *inventoryContainsEquivalentObject4E7EC0
	next  *inventoryContainsEquivalentObject4E7EC0
}

func inventoryContainsEquivalentTestHooks4E7EC0(events *[]string) inventoryContainsEquivalentHooks4E7EC0[*inventoryContainsEquivalentObject4E7EC0] {
	return inventoryContainsEquivalentHooks4E7EC0[*inventoryContainsEquivalentObject4E7EC0]{
		loadFirst: func(owner *inventoryContainsEquivalentObject4E7EC0) *inventoryContainsEquivalentObject4E7EC0 {
			*events = append(*events, "first:"+owner.name)
			return owner.first
		},
		equivalent: func(candidate, item *inventoryContainsEquivalentObject4E7EC0) bool {
			*events = append(*events, "equivalent:"+candidate.name+":"+item.name)
			return candidate.match
		},
		loadNext: func(candidate *inventoryContainsEquivalentObject4E7EC0) *inventoryContainsEquivalentObject4E7EC0 {
			*events = append(*events, "next:"+candidate.name)
			return candidate.next
		},
	}
}

func TestInventoryContainsEquivalent4E7EC0GuardsAndEmptyOrder(t *testing.T) {
	owner := &inventoryContainsEquivalentObject4E7EC0{name: "owner"}
	item := &inventoryContainsEquivalentObject4E7EC0{name: "item"}
	for _, tc := range []struct {
		name       string
		owner      *inventoryContainsEquivalentObject4E7EC0
		item       *inventoryContainsEquivalentObject4E7EC0
		wantEvents []string
	}{
		{name: "nil owner", item: item},
		{name: "nil item", owner: owner},
		{name: "empty inventory", owner: owner, item: item, wantEvents: []string{"first:owner"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			got := inventoryContainsEquivalent4E7EC0(
				tc.owner, tc.item, inventoryContainsEquivalentTestHooks4E7EC0(&events),
			)
			if got {
				t.Fatal("result = true, want false")
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestInventoryContainsEquivalent4E7EC0StopsBeforeMatchedNext(t *testing.T) {
	invalid := &inventoryContainsEquivalentObject4E7EC0{name: "must-not-load"}
	second := &inventoryContainsEquivalentObject4E7EC0{name: "second", match: true, next: invalid}
	first := &inventoryContainsEquivalentObject4E7EC0{name: "first", next: second}
	owner := &inventoryContainsEquivalentObject4E7EC0{name: "owner", first: first}
	item := &inventoryContainsEquivalentObject4E7EC0{name: "item"}
	var events []string

	if !inventoryContainsEquivalent4E7EC0(owner, item, inventoryContainsEquivalentTestHooks4E7EC0(&events)) {
		t.Fatal("second candidate match returned false")
	}
	want := []string{
		"first:owner", "equivalent:first:item", "next:first", "equivalent:second:item",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestInventoryContainsEquivalent4E7EC0NoMatchLoadsEveryNext(t *testing.T) {
	second := &inventoryContainsEquivalentObject4E7EC0{name: "second"}
	first := &inventoryContainsEquivalentObject4E7EC0{name: "first", next: second}
	owner := &inventoryContainsEquivalentObject4E7EC0{name: "owner", first: first}
	item := &inventoryContainsEquivalentObject4E7EC0{name: "item"}
	var events []string

	if inventoryContainsEquivalent4E7EC0(owner, item, inventoryContainsEquivalentTestHooks4E7EC0(&events)) {
		t.Fatal("nonmatching inventory returned true")
	}
	want := []string{
		"first:owner", "equivalent:first:item", "next:first",
		"equivalent:second:item", "next:second",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestInventoryContainsEquivalent4E7EC0UsesLiveNextAfterMismatch(t *testing.T) {
	replacement := &inventoryContainsEquivalentObject4E7EC0{name: "replacement", match: true}
	stale := &inventoryContainsEquivalentObject4E7EC0{name: "stale"}
	first := &inventoryContainsEquivalentObject4E7EC0{name: "first", next: stale}
	owner := &inventoryContainsEquivalentObject4E7EC0{name: "owner", first: first}
	item := &inventoryContainsEquivalentObject4E7EC0{name: "item"}
	var events []string
	hooks := inventoryContainsEquivalentTestHooks4E7EC0(&events)
	equivalent := hooks.equivalent
	hooks.equivalent = func(candidate, gotItem *inventoryContainsEquivalentObject4E7EC0) bool {
		match := equivalent(candidate, gotItem)
		if candidate == first {
			candidate.next = replacement
		}
		return match
	}

	if !inventoryContainsEquivalent4E7EC0(owner, item, hooks) {
		t.Fatal("live replacement match returned false")
	}
	want := []string{
		"first:owner", "equivalent:first:item", "next:first", "equivalent:replacement:item",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
