package legacy

import (
	"math"
	"reflect"
	"testing"
)

type countInventoryClassObject4E7D70 struct {
	name  string
	class uint32
	first *countInventoryClassObject4E7D70
	next  *countInventoryClassObject4E7D70
}

func countInventoryClassTestHooks4E7D70(events *[]string) countInventoryClassHooks4E7D70[*countInventoryClassObject4E7D70] {
	return countInventoryClassHooks4E7D70[*countInventoryClassObject4E7D70]{
		loadFirst: func(owner *countInventoryClassObject4E7D70) *countInventoryClassObject4E7D70 {
			*events = append(*events, "first")
			return owner.first
		},
		loadClass: func(obj *countInventoryClassObject4E7D70) uint32 {
			*events = append(*events, "class:"+obj.name)
			return obj.class
		},
		loadNext: func(obj *countInventoryClassObject4E7D70) *countInventoryClassObject4E7D70 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestCountInventoryClass4E7D70GuardAndEmptyOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		owner      *countInventoryClassObject4E7D70
		mask       uint32
		wantEvents []string
	}{
		{name: "nil owner", mask: 1},
		{name: "zero mask", owner: &countInventoryClassObject4E7D70{}},
		{name: "empty owner", owner: &countInventoryClassObject4E7D70{}, mask: 1, wantEvents: []string{"first"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			if got := countInventoryClass4E7D70(tc.owner, tc.mask, countInventoryClassTestHooks4E7D70(&events)); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestCountInventoryClass4E7D70ReadOrderAndIntersection(t *testing.T) {
	fourth := &countInventoryClassObject4E7D70{name: "fourth", class: 0x80000000}
	third := &countInventoryClassObject4E7D70{name: "third", class: 0, next: fourth}
	second := &countInventoryClassObject4E7D70{name: "second", class: 0x6, next: third}
	first := &countInventoryClassObject4E7D70{name: "first", class: 0x1, next: second}
	owner := &countInventoryClassObject4E7D70{first: first}
	var events []string

	if got := countInventoryClass4E7D70(owner, 0x80000002, countInventoryClassTestHooks4E7D70(&events)); got != 2 {
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
		t.Fatal("count mutated the inventory list")
	}
}

func TestCountInventoryClass4E7D70UsesLiveClassAndNext(t *testing.T) {
	replacement := &countInventoryClassObject4E7D70{name: "replacement", class: 0x8}
	stale := &countInventoryClassObject4E7D70{name: "stale", class: 0x8}
	first := &countInventoryClassObject4E7D70{name: "first", class: 0, next: stale}
	owner := &countInventoryClassObject4E7D70{first: first}
	var events []string
	hooks := countInventoryClassTestHooks4E7D70(&events)
	hooks.loadClass = func(obj *countInventoryClassObject4E7D70) uint32 {
		events = append(events, "class:"+obj.name)
		if obj == first {
			obj.class = 0x8
			obj.next = replacement
		}
		return obj.class
	}

	if got := countInventoryClass4E7D70(owner, 0x8, hooks); got != 2 {
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

func TestCountInventoryClass4E7D70UsesCompleteUint32Mask(t *testing.T) {
	item := &countInventoryClassObject4E7D70{name: "item", class: math.MaxUint32}
	owner := &countInventoryClassObject4E7D70{first: item}
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
			if got := countInventoryClass4E7D70(owner, tc.mask, countInventoryClassTestHooks4E7D70(&events)); got != tc.want {
				t.Fatalf("mask %#x result = %d, want %d", tc.mask, got, tc.want)
			}
			if tc.mask == 0 && len(events) != 0 {
				t.Fatalf("zero mask events = %v, want none", events)
			}
		})
	}
}

func TestIncrementInventoryClassCount4E7D70WrapsInt32(t *testing.T) {
	if got := incrementInventoryClassCount4E7D70(math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("increment(MaxInt32) = %d, want %d", got, int32(math.MinInt32))
	}
}
