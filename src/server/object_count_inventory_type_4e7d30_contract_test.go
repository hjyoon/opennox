package server

import (
	"math"
	"reflect"
	"testing"
)

type countInventoryTypeObject4E7D30 struct {
	name     string
	typeInd  uint16
	flagsLow uint8
	first    *countInventoryTypeObject4E7D30
	next     *countInventoryTypeObject4E7D30
}

func countInventoryTypeTestHooks4E7D30(events *[]string) countInventoryTypeHooks4E7D30[*countInventoryTypeObject4E7D30] {
	return countInventoryTypeHooks4E7D30[*countInventoryTypeObject4E7D30]{
		loadFirst: func(owner *countInventoryTypeObject4E7D30) *countInventoryTypeObject4E7D30 {
			*events = append(*events, "first")
			return owner.first
		},
		loadType: func(obj *countInventoryTypeObject4E7D30) uint16 {
			*events = append(*events, "type:"+obj.name)
			return obj.typeInd
		},
		loadFlagsLow: func(obj *countInventoryTypeObject4E7D30) uint8 {
			*events = append(*events, "flags:"+obj.name)
			return obj.flagsLow
		},
		loadNext: func(obj *countInventoryTypeObject4E7D30) *countInventoryTypeObject4E7D30 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestCountInventoryType4E7D30NilAndEmptyOwner(t *testing.T) {
	for _, tc := range []struct {
		name       string
		owner      *countInventoryTypeObject4E7D30
		wantEvents []string
	}{
		{name: "nil"},
		{name: "empty", owner: &countInventoryTypeObject4E7D30{}, wantEvents: []string{"first"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := countInventoryTypeTestHooks4E7D30(&events)
			if got := countInventoryType4E7D30(tc.owner, 7, hooks); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestCountInventoryType4E7D30ZeroQuerySkipsTypeAndFlags(t *testing.T) {
	third := &countInventoryTypeObject4E7D30{name: "third", typeInd: 9, flagsLow: objectDestroyedFlagLow4E7D30}
	second := &countInventoryTypeObject4E7D30{name: "second", typeInd: 8, next: third}
	first := &countInventoryTypeObject4E7D30{name: "first", typeInd: 7, flagsLow: objectDestroyedFlagLow4E7D30, next: second}
	owner := &countInventoryTypeObject4E7D30{first: first}
	var events []string
	hooks := countInventoryTypeTestHooks4E7D30(&events)

	if got := countInventoryType4E7D30(owner, 0, hooks); got != 3 {
		t.Fatalf("result = %d, want 3", got)
	}
	want := []string{"first", "next:first", "next:second", "next:third"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCountInventoryType4E7D30ReadOrderAndDestroyedFilter(t *testing.T) {
	fourth := &countInventoryTypeObject4E7D30{name: "fourth", typeInd: 7}
	third := &countInventoryTypeObject4E7D30{name: "third", typeInd: 7, flagsLow: objectDestroyedFlagLow4E7D30, next: fourth}
	second := &countInventoryTypeObject4E7D30{name: "second", typeInd: 7, next: third}
	first := &countInventoryTypeObject4E7D30{name: "first", typeInd: 6, flagsLow: objectDestroyedFlagLow4E7D30, next: second}
	owner := &countInventoryTypeObject4E7D30{first: first}
	var events []string
	hooks := countInventoryTypeTestHooks4E7D30(&events)

	if got := countInventoryType4E7D30(owner, 7, hooks); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"first",
		"type:first", "next:first",
		"type:second", "flags:second", "next:second",
		"type:third", "flags:third", "next:third",
		"type:fourth", "flags:fourth", "next:fourth",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCountInventoryType4E7D30UsesLiveFlagsAndNext(t *testing.T) {
	replacement := &countInventoryTypeObject4E7D30{name: "replacement", typeInd: 9}
	stale := &countInventoryTypeObject4E7D30{name: "stale", typeInd: 9}
	first := &countInventoryTypeObject4E7D30{name: "first", typeInd: 9, next: stale}
	owner := &countInventoryTypeObject4E7D30{first: first}
	var events []string
	hooks := countInventoryTypeTestHooks4E7D30(&events)
	hooks.loadType = func(obj *countInventoryTypeObject4E7D30) uint16 {
		events = append(events, "type:"+obj.name)
		if obj == first {
			obj.flagsLow = objectDestroyedFlagLow4E7D30
			obj.next = replacement
		}
		return obj.typeInd
	}

	if got := countInventoryType4E7D30(owner, 9, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"first", "type:first", "flags:first", "next:first",
		"type:replacement", "flags:replacement", "next:replacement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCountInventoryType4E7D30ComparesFullInt32Type(t *testing.T) {
	for _, tc := range []struct {
		name  string
		item  uint16
		query int32
		want  int32
	}{
		{name: "maximum", item: math.MaxUint16, query: math.MaxUint16, want: 1},
		{name: "high cache", item: 1, query: 0x00010001},
		{name: "negative", item: math.MaxUint16, query: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := &countInventoryTypeObject4E7D30{name: "item", typeInd: tc.item}
			owner := &countInventoryTypeObject4E7D30{first: item}
			var events []string
			hooks := countInventoryTypeTestHooks4E7D30(&events)
			if got := countInventoryType4E7D30(owner, tc.query, hooks); got != tc.want {
				t.Fatalf("type %#x query %#x result = %d, want %d", tc.item, uint32(tc.query), got, tc.want)
			}
			want := []string{"first", "type:item"}
			if tc.want != 0 {
				want = append(want, "flags:item")
			}
			want = append(want, "next:item")
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestIncrementInventoryTypeCount4E7D30WrapsInt32(t *testing.T) {
	if got := incrementInventoryTypeCount4E7D30(math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("increment(MaxInt32) = %d, want %d", got, int32(math.MinInt32))
	}
}
