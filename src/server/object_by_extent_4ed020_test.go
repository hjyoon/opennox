package server

import (
	"fmt"
	"reflect"
	"testing"
)

type objectByExtentTestObject4ED020 struct {
	name     string
	flagsLow uint8
	extent   uint32
	next     *objectByExtentTestObject4ED020
}

type objectByExtentTestWorld4ED020 struct {
	first   *objectByExtentTestObject4ED020
	wanted  uint32
	events  []string
	faultAt int
}

func (w *objectByExtentTestWorld4ED020) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *objectByExtentTestWorld4ED020) hooks() objectByExtentHooks4ED020[*objectByExtentTestObject4ED020] {
	return objectByExtentHooks4ED020[*objectByExtentTestObject4ED020]{
		first: func() *objectByExtentTestObject4ED020 {
			w.event("first")
			return w.first
		},
		loadExtentArg: func() uint32 {
			w.event(fmt.Sprintf("argument:%08x", w.wanted))
			return w.wanted
		},
		loadFlagsLow: func(obj *objectByExtentTestObject4ED020) uint8 {
			w.event("flags:" + obj.name)
			return obj.flagsLow
		},
		loadExtent: func(obj *objectByExtentTestObject4ED020) uint32 {
			w.event("extent:" + obj.name)
			return obj.extent
		},
		next: func(obj *objectByExtentTestObject4ED020) *objectByExtentTestObject4ED020 {
			w.event("next:" + obj.name)
			return obj.next
		},
	}
}

func newObjectByExtentFullSearch4ED020() (*objectByExtentTestWorld4ED020, *objectByExtentTestObject4ED020, []string) {
	const wanted = uint32(0xffffffff)
	dead := &objectByExtentTestObject4ED020{
		name:     "dead",
		flagsLow: objectDestroyedFlagLow4ED020,
		extent:   wanted,
	}
	other := &objectByExtentTestObject4ED020{name: "other", extent: wanted - 1}
	match := &objectByExtentTestObject4ED020{name: "match", extent: wanted}
	dead.next = other
	other.next = match
	w := &objectByExtentTestWorld4ED020{first: dead, wanted: wanted}
	wantEvents := []string{
		"first", "argument:ffffffff",
		"flags:dead", "next:dead",
		"flags:other", "extent:other", "next:other",
		"flags:match", "extent:match",
	}
	return w, match, wantEvents
}

func TestObjectByExtent4ED020SearchOrder(t *testing.T) {
	w, wantObject, wantEvents := newObjectByExtentFullSearch4ED020()
	got := objectByExtent4ED020(w.hooks())
	if got != wantObject {
		t.Fatalf("result = %p, want %p", got, wantObject)
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestObjectByExtent4ED020FaultOrder(t *testing.T) {
	_, _, wantEvents := newObjectByExtentFullSearch4ED020()
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w, _, _ := newObjectByExtentFullSearch4ED020()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %v, want %v", w.events, want)
				}
			}()
			objectByExtent4ED020(w.hooks())
		})
	}
}

func TestObjectByExtent4ED020EmptySkipsArgument(t *testing.T) {
	w := &objectByExtentTestWorld4ED020{wanted: 0xffffffff}
	if got := objectByExtent4ED020(w.hooks()); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if want := []string{"first"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestObjectByExtent4ED020UsesLowDestroyedBitAndFullUnsignedExtent(t *testing.T) {
	match := &objectByExtentTestObject4ED020{name: "match", flagsLow: 0x80, extent: 0xffffffff}
	w := &objectByExtentTestWorld4ED020{first: match, wanted: 0xffffffff}
	if got := objectByExtent4ED020(w.hooks()); got != match {
		t.Fatalf("result = %p, want %p", got, match)
	}
	want := []string{"first", "argument:ffffffff", "flags:match", "extent:match"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestObjectByExtent4ED020MissStopsAfterNilSuccessor(t *testing.T) {
	other := &objectByExtentTestObject4ED020{name: "other", extent: 1}
	w := &objectByExtentTestWorld4ED020{first: other, wanted: 2}
	if got := objectByExtent4ED020(w.hooks()); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	want := []string{"first", "argument:00000002", "flags:other", "extent:other", "next:other"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
