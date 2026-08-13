package server

import (
	"math"
	"reflect"
	"testing"
)

type countOwnedTypeObject4E7C80 struct {
	name     string
	typeInd  uint16
	flagsLow uint8
	first    *countOwnedTypeObject4E7C80
	next     *countOwnedTypeObject4E7C80
}

func countOwnedTypeTestHooks4E7C80(events *[]string) countOwnedTypeHooks4E7C80[*countOwnedTypeObject4E7C80] {
	return countOwnedTypeHooks4E7C80[*countOwnedTypeObject4E7C80]{
		loadFirst: func(owner *countOwnedTypeObject4E7C80) *countOwnedTypeObject4E7C80 {
			*events = append(*events, "first")
			return owner.first
		},
		loadType: func(obj *countOwnedTypeObject4E7C80) uint16 {
			*events = append(*events, "type:"+obj.name)
			return obj.typeInd
		},
		loadFlagsLow: func(obj *countOwnedTypeObject4E7C80) uint8 {
			*events = append(*events, "flags:"+obj.name)
			return obj.flagsLow
		},
		loadNext: func(obj *countOwnedTypeObject4E7C80) *countOwnedTypeObject4E7C80 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestCountOwnedType4E7C80NilAndEmptyOwner(t *testing.T) {
	for _, tc := range []struct {
		name       string
		owner      *countOwnedTypeObject4E7C80
		wantEvents []string
	}{
		{name: "nil"},
		{name: "empty", owner: &countOwnedTypeObject4E7C80{}, wantEvents: []string{"first"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := countOwnedTypeTestHooks4E7C80(&events)
			if got := countOwnedType4E7C80(tc.owner, 7, hooks); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestCountOwnedType4E7C80ReadOrderAndDestroyedFilter(t *testing.T) {
	fourth := &countOwnedTypeObject4E7C80{name: "fourth", typeInd: 7}
	third := &countOwnedTypeObject4E7C80{name: "third", typeInd: 7, flagsLow: objectDestroyedFlagLow4E7C80, next: fourth}
	second := &countOwnedTypeObject4E7C80{name: "second", typeInd: 7, next: third}
	first := &countOwnedTypeObject4E7C80{name: "first", typeInd: 6, flagsLow: objectDestroyedFlagLow4E7C80, next: second}
	owner := &countOwnedTypeObject4E7C80{first: first}
	var events []string
	hooks := countOwnedTypeTestHooks4E7C80(&events)

	if got := countOwnedType4E7C80(owner, 7, hooks); got != 2 {
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
	if first.next != second || second.next != third || third.next != fourth || fourth.next != nil {
		t.Fatal("count mutated the owned list")
	}
}

func TestCountOwnedType4E7C80UsesLiveFlagsAndNext(t *testing.T) {
	replacement := &countOwnedTypeObject4E7C80{name: "replacement", typeInd: 9}
	stale := &countOwnedTypeObject4E7C80{name: "stale", typeInd: 9}
	first := &countOwnedTypeObject4E7C80{name: "first", typeInd: 9, next: stale}
	owner := &countOwnedTypeObject4E7C80{first: first}
	var events []string
	hooks := countOwnedTypeTestHooks4E7C80(&events)
	hooks.loadType = func(obj *countOwnedTypeObject4E7C80) uint16 {
		events = append(events, "type:"+obj.name)
		if obj == first {
			obj.flagsLow = objectDestroyedFlagLow4E7C80
			obj.next = replacement
		}
		return obj.typeInd
	}

	if got := countOwnedType4E7C80(owner, 9, hooks); got != 1 {
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

func TestCountOwnedType4E7C80ComparesFullInt32Type(t *testing.T) {
	for _, tc := range []struct {
		name    string
		item    uint16
		query   int32
		want    int32
		wantLog []string
	}{
		{name: "zero", want: 1},
		{name: "maximum", item: math.MaxUint16, query: math.MaxUint16, want: 1},
		{name: "high cache", item: 1, query: 0x00010001},
		{name: "negative", item: math.MaxUint16, query: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := &countOwnedTypeObject4E7C80{name: "item", typeInd: tc.item}
			owner := &countOwnedTypeObject4E7C80{first: item}
			var events []string
			hooks := countOwnedTypeTestHooks4E7C80(&events)
			if got := countOwnedType4E7C80(owner, tc.query, hooks); got != tc.want {
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

func TestIncrementOwnedTypeCount4E7C80WrapsInt32(t *testing.T) {
	if got := incrementOwnedTypeCount4E7C80(math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("increment(MaxInt32) = %d, want %d", got, int32(math.MinInt32))
	}
}
