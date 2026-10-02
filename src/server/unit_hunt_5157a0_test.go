package server

import (
	"fmt"
	"reflect"
	"testing"
)

const unitHuntHighHandle5157A0 = uint64(0x7fd4c759df40)

type unitHuntTestWorld5157A0 struct {
	events   []string
	faultAt  int
	classLow uint8
	flags    uint32
}

func (w *unitHuntTestWorld5157A0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitHuntTestWorld5157A0) hooks() unitHuntHooks5157A0[uint64] {
	return unitHuntHooks5157A0[uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadFlags: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		clearActionStack: func(unit uint64) { w.observe(fmt.Sprintf("clear:%x", unit)) },
		pushAction: func(unit uint64, action uint32) {
			w.observe(fmt.Sprintf("push:%x:%d", unit, action))
		},
	}
}

func TestUnitHunt5157A0ExactOrderAndGates(t *testing.T) {
	full := []string{"class:7fd4c759df40", "flags:7fd4c759df40", "clear:7fd4c759df40", "push:7fd4c759df40:5"}
	for _, tc := range []struct {
		name  string
		unit  uint64
		class uint8
		flags uint32
		want  []string
	}{
		{name: "null"},
		{"not monster", unitHuntHighHandle5157A0, 0xfd, 0, full[:1]},
		{"dead", unitHuntHighHandle5157A0, 0x82, 0x8000, full[:2]},
		{"live", unitHuntHighHandle5157A0, 2, 0, full},
		{"all unrelated flags", unitHuntHighHandle5157A0, 0x82, 0xffff7fff, full},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &unitHuntTestWorld5157A0{classLow: tc.class, flags: tc.flags}
			unitHunt5157A0(tc.unit, w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events=%q want=%q", w.events, tc.want)
			}
		})
	}
}

func TestUnitHunt5157A0AllFaultPrefixes(t *testing.T) {
	baseline := &unitHuntTestWorld5157A0{classLow: 2}
	unitHunt5157A0(unitHuntHighHandle5157A0, baseline.hooks())
	for faultAt := 1; faultAt <= len(baseline.events); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &unitHuntTestWorld5157A0{classLow: 2, faultAt: faultAt}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				unitHunt5157A0(unitHuntHighHandle5157A0, w.hooks())
			}()
			if recovered == nil || !reflect.DeepEqual(w.events, baseline.events[:faultAt]) {
				t.Fatalf("fault=%v events=%q want=%q", recovered, w.events, baseline.events[:faultAt])
			}
		})
	}
}

func TestUnitHunt5157A0DoesNotReloadGatesAfterClear(t *testing.T) {
	w := &unitHuntTestWorld5157A0{classLow: 2}
	hooks := w.hooks()
	clear := hooks.clearActionStack
	hooks.clearActionStack = func(unit uint64) {
		clear(unit)
		w.classLow, w.flags = 0, 0x8000
	}
	unitHunt5157A0(unitHuntHighHandle5157A0, hooks)
	want := []string{"class:7fd4c759df40", "flags:7fd4c759df40", "clear:7fd4c759df40", "push:7fd4c759df40:5"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events=%q want=%q", w.events, want)
	}
}
