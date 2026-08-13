package legacy

import (
	"math"
	"reflect"
	"testing"
)

type countSlavesObject4E7CF0 struct {
	name     string
	class    uint32
	subclass uint32
	first    *countSlavesObject4E7CF0
	next     *countSlavesObject4E7CF0
}

func countSlavesTestHooks4E7CF0(events *[]string) countSlavesHooks4E7CF0[*countSlavesObject4E7CF0] {
	return countSlavesHooks4E7CF0[*countSlavesObject4E7CF0]{
		loadFirst: func(owner *countSlavesObject4E7CF0) *countSlavesObject4E7CF0 {
			*events = append(*events, "first")
			return owner.first
		},
		loadClass: func(obj *countSlavesObject4E7CF0) uint32 {
			*events = append(*events, "class:"+obj.name)
			return obj.class
		},
		loadSubclass: func(obj *countSlavesObject4E7CF0) uint32 {
			*events = append(*events, "subclass:"+obj.name)
			return obj.subclass
		},
		loadNext: func(obj *countSlavesObject4E7CF0) *countSlavesObject4E7CF0 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestCountSlaves4E7CF0GuardAndEmptyOrder(t *testing.T) {
	owner := &countSlavesObject4E7CF0{}
	for _, tc := range []struct {
		name         string
		owner        *countSlavesObject4E7CF0
		classMask    uint32
		subclassMask uint32
		wantEvents   []string
	}{
		{name: "nil owner", classMask: 1, subclassMask: 1},
		{name: "zero class", owner: owner, subclassMask: 1},
		{name: "zero subclass", owner: owner, classMask: 1},
		{name: "empty owner", owner: owner, classMask: 1, subclassMask: 1, wantEvents: []string{"first"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			got := countSlaves4E7CF0(
				tc.owner, tc.classMask, tc.subclassMask, countSlavesTestHooks4E7CF0(&events),
			)
			if got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestCountSlaves4E7CF0ReadOrderAndIntersections(t *testing.T) {
	fourth := &countSlavesObject4E7CF0{name: "fourth", class: 0x80000000, subclass: 0x40000000}
	third := &countSlavesObject4E7CF0{name: "third", class: 0x2, subclass: 0x4, next: fourth}
	second := &countSlavesObject4E7CF0{name: "second", class: 0x6, subclass: 0x1, next: third}
	first := &countSlavesObject4E7CF0{name: "first", class: 0x1, subclass: 0x40000000, next: second}
	owner := &countSlavesObject4E7CF0{first: first}
	var events []string

	if got := countSlaves4E7CF0(owner, 0x80000002, 0x40000004, countSlavesTestHooks4E7CF0(&events)); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"first",
		"class:first", "next:first",
		"class:second", "subclass:second", "next:second",
		"class:third", "subclass:third", "next:third",
		"class:fourth", "subclass:fourth", "next:fourth",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if first.next != second || second.next != third || third.next != fourth || fourth.next != nil {
		t.Fatal("count mutated the owned list")
	}
}

func TestCountSlaves4E7CF0UsesLiveSubclassAndNext(t *testing.T) {
	replacement := &countSlavesObject4E7CF0{name: "replacement", class: 0x8, subclass: 0x10}
	stale := &countSlavesObject4E7CF0{name: "stale", class: 0x8, subclass: 0x10}
	first := &countSlavesObject4E7CF0{name: "first", class: 0x8, next: stale}
	owner := &countSlavesObject4E7CF0{first: first}
	var events []string
	hooks := countSlavesTestHooks4E7CF0(&events)
	hooks.loadClass = func(obj *countSlavesObject4E7CF0) uint32 {
		events = append(events, "class:"+obj.name)
		if obj == first {
			obj.subclass = 0x10
			obj.next = replacement
		}
		return obj.class
	}

	if got := countSlaves4E7CF0(owner, 0x8, 0x10, hooks); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"first", "class:first", "subclass:first", "next:first",
		"class:replacement", "subclass:replacement", "next:replacement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCountSlaves4E7CF0UsesCompleteUint32Masks(t *testing.T) {
	item := &countSlavesObject4E7CF0{name: "item", class: math.MaxUint32, subclass: math.MaxUint32}
	owner := &countSlavesObject4E7CF0{first: item}
	for _, tc := range []struct {
		name         string
		classMask    uint32
		subclassMask uint32
		want         int32
	}{
		{name: "low", classMask: 1, subclassMask: 2, want: 1},
		{name: "high", classMask: 0x80000000, subclassMask: 0x40000000, want: 1},
		{name: "zero class", subclassMask: 1},
		{name: "zero subclass", classMask: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			got := countSlaves4E7CF0(
				owner, tc.classMask, tc.subclassMask, countSlavesTestHooks4E7CF0(&events),
			)
			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if (tc.classMask == 0 || tc.subclassMask == 0) && len(events) != 0 {
				t.Fatalf("zero-mask events = %v, want none", events)
			}
		})
	}
}

func TestIncrementSlaveCount4E7CF0WrapsInt32(t *testing.T) {
	if got := incrementSlaveCount4E7CF0(math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("increment(MaxInt32) = %d, want %d", got, int32(math.MinInt32))
	}
}
