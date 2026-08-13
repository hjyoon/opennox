package legacy

import (
	"math"
	"reflect"
	"testing"
)

type countOwnedClassObject4E7CC0 struct {
	name  string
	class uint32
	first *countOwnedClassObject4E7CC0
	next  *countOwnedClassObject4E7CC0
}

func countOwnedClassTestHooks4E7CC0(events *[]string) countOwnedClassHooks4E7CC0[*countOwnedClassObject4E7CC0] {
	return countOwnedClassHooks4E7CC0[*countOwnedClassObject4E7CC0]{
		loadFirst: func(owner *countOwnedClassObject4E7CC0) *countOwnedClassObject4E7CC0 {
			*events = append(*events, "first")
			return owner.first
		},
		loadClass: func(obj *countOwnedClassObject4E7CC0) uint32 {
			*events = append(*events, "class:"+obj.name)
			return obj.class
		},
		loadNext: func(obj *countOwnedClassObject4E7CC0) *countOwnedClassObject4E7CC0 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestCountOwnedClass4E7CC0GuardAndEmptyOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		owner      *countOwnedClassObject4E7CC0
		mask       uint32
		wantEvents []string
	}{
		{name: "nil owner", mask: 1},
		{name: "zero mask", owner: &countOwnedClassObject4E7CC0{}},
		{name: "empty owner", owner: &countOwnedClassObject4E7CC0{}, mask: 1, wantEvents: []string{"first"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			if got := countOwnedClass4E7CC0(tc.owner, tc.mask, countOwnedClassTestHooks4E7CC0(&events)); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestCountOwnedClass4E7CC0ReadOrderAndIntersection(t *testing.T) {
	fourth := &countOwnedClassObject4E7CC0{name: "fourth", class: 0x80000000}
	third := &countOwnedClassObject4E7CC0{name: "third", class: 0, next: fourth}
	second := &countOwnedClassObject4E7CC0{name: "second", class: 0x6, next: third}
	first := &countOwnedClassObject4E7CC0{name: "first", class: 0x1, next: second}
	owner := &countOwnedClassObject4E7CC0{first: first}
	var events []string

	if got := countOwnedClass4E7CC0(owner, 0x80000002, countOwnedClassTestHooks4E7CC0(&events)); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"first",
		"class:first", "next:first",
		"class:second", "next:second",
		"class:third", "next:third",
		"class:fourth", "next:fourth",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if first.next != second || second.next != third || third.next != fourth || fourth.next != nil {
		t.Fatal("count mutated the owned list")
	}
}

func TestCountOwnedClass4E7CC0UsesLiveClassAndNext(t *testing.T) {
	replacement := &countOwnedClassObject4E7CC0{name: "replacement", class: 0x8}
	stale := &countOwnedClassObject4E7CC0{name: "stale", class: 0x8}
	first := &countOwnedClassObject4E7CC0{name: "first", class: 0, next: stale}
	owner := &countOwnedClassObject4E7CC0{first: first}
	var events []string
	hooks := countOwnedClassTestHooks4E7CC0(&events)
	hooks.loadClass = func(obj *countOwnedClassObject4E7CC0) uint32 {
		events = append(events, "class:"+obj.name)
		if obj == first {
			obj.class = 0x8
			obj.next = replacement
		}
		return obj.class
	}

	if got := countOwnedClass4E7CC0(owner, 0x8, hooks); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	want := []string{
		"first", "class:first", "next:first",
		"class:replacement", "next:replacement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCountOwnedClass4E7CC0UsesCompleteUint32Mask(t *testing.T) {
	item := &countOwnedClassObject4E7CC0{name: "item", class: math.MaxUint32}
	owner := &countOwnedClassObject4E7CC0{first: item}
	for _, tc := range []struct {
		name string
		mask uint32
		want int32
	}{
		{name: "low", mask: 1, want: 1},
		{name: "high", mask: 0x80000000, want: 1},
		{name: "zero"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			if got := countOwnedClass4E7CC0(owner, tc.mask, countOwnedClassTestHooks4E7CC0(&events)); got != tc.want {
				t.Fatalf("mask %#x result = %d, want %d", tc.mask, got, tc.want)
			}
			if tc.mask == 0 && len(events) != 0 {
				t.Fatalf("zero mask events = %v, want none", events)
			}
		})
	}
}

func TestIncrementOwnedClassCount4E7CC0WrapsInt32(t *testing.T) {
	if got := incrementOwnedClassCount4E7CC0(math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("increment(MaxInt32) = %d, want %d", got, int32(math.MinInt32))
	}
}
