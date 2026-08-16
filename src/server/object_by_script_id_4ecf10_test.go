package server

import (
	"reflect"
	"testing"
)

type objectByScriptIDTestObject4ECF10 struct {
	name           string
	flags          uint8
	scriptID       int32
	next           *objectByScriptIDTestObject4ECF10
	firstInventory *objectByScriptIDTestObject4ECF10
	invNext        *objectByScriptIDTestObject4ECF10
}

type objectByScriptIDTestWorld4ECF10 struct {
	active  *objectByScriptIDTestObject4ECF10
	pending *objectByScriptIDTestObject4ECF10
	missile *objectByScriptIDTestObject4ECF10
	wanted  int32
	events  []string
	faultAt int
}

func (w *objectByScriptIDTestWorld4ECF10) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *objectByScriptIDTestWorld4ECF10) hooks() objectByScriptIDHooks4ECF10[*objectByScriptIDTestObject4ECF10] {
	return objectByScriptIDHooks4ECF10[*objectByScriptIDTestObject4ECF10]{
		firstActive: func() *objectByScriptIDTestObject4ECF10 {
			w.event("first-active")
			return w.active
		},
		loadScriptIDArg: func() int32 {
			w.event("argument")
			return w.wanted
		},
		nextActive: func(obj *objectByScriptIDTestObject4ECF10) *objectByScriptIDTestObject4ECF10 {
			w.event("next-active:" + obj.name)
			return obj.next
		},
		firstInventory: func(obj *objectByScriptIDTestObject4ECF10) *objectByScriptIDTestObject4ECF10 {
			w.event("first-inventory:" + obj.name)
			return obj.firstInventory
		},
		nextInventory: func(obj *objectByScriptIDTestObject4ECF10) *objectByScriptIDTestObject4ECF10 {
			w.event("next-inventory:" + obj.name)
			return obj.invNext
		},
		firstPending: func() *objectByScriptIDTestObject4ECF10 {
			w.event("first-pending")
			return w.pending
		},
		nextPending: func(obj *objectByScriptIDTestObject4ECF10) *objectByScriptIDTestObject4ECF10 {
			w.event("next-pending:" + obj.name)
			return obj.next
		},
		firstMissile: func() *objectByScriptIDTestObject4ECF10 {
			w.event("first-missile")
			return w.missile
		},
		nextMissile: func(obj *objectByScriptIDTestObject4ECF10) *objectByScriptIDTestObject4ECF10 {
			w.event("next-missile:" + obj.name)
			return obj.next
		},
		loadFlagsLow: func(obj *objectByScriptIDTestObject4ECF10) uint8 {
			w.event("flags:" + obj.name)
			return obj.flags
		},
		loadScriptID: func(obj *objectByScriptIDTestObject4ECF10) int32 {
			w.event("script-id:" + obj.name)
			return obj.scriptID
		},
	}
}

func newObjectByScriptIDFullSearch4ECF10() (*objectByScriptIDTestWorld4ECF10, *objectByScriptIDTestObject4ECF10, []string) {
	const wanted = int32(-2147483647)
	activeDead := &objectByScriptIDTestObject4ECF10{name: "active-dead", flags: objectDeadFlagLow4ECF10, scriptID: wanted}
	itemDead := &objectByScriptIDTestObject4ECF10{name: "item-dead", flags: objectDeadFlagLow4ECF10, scriptID: wanted}
	itemOther := &objectByScriptIDTestObject4ECF10{name: "item-other", scriptID: wanted + 1}
	itemDead.invNext = itemOther
	activeDead.firstInventory = itemDead
	activeOther := &objectByScriptIDTestObject4ECF10{name: "active-other", scriptID: wanted + 2}
	activeDead.next = activeOther
	pendingDead := &objectByScriptIDTestObject4ECF10{name: "pending-dead", flags: objectDeadFlagLow4ECF10, scriptID: wanted}
	pendingOther := &objectByScriptIDTestObject4ECF10{name: "pending-other", scriptID: wanted + 3}
	pendingDead.next = pendingOther
	missileDead := &objectByScriptIDTestObject4ECF10{name: "missile-dead", flags: objectDeadFlagLow4ECF10, scriptID: wanted}
	missileMatch := &objectByScriptIDTestObject4ECF10{name: "missile-match", scriptID: wanted}
	missileDead.next = missileMatch
	w := &objectByScriptIDTestWorld4ECF10{
		active:  activeDead,
		pending: pendingDead,
		missile: missileDead,
		wanted:  wanted,
	}
	wantEvents := []string{
		"first-active",
		"argument",
		"flags:active-dead",
		"first-inventory:active-dead",
		"flags:item-dead",
		"next-inventory:item-dead",
		"flags:item-other",
		"script-id:item-other",
		"next-inventory:item-other",
		"next-active:active-dead",
		"flags:active-other",
		"script-id:active-other",
		"first-inventory:active-other",
		"next-active:active-other",
		"first-pending",
		"flags:pending-dead",
		"next-pending:pending-dead",
		"flags:pending-other",
		"script-id:pending-other",
		"next-pending:pending-other",
		"first-missile",
		"flags:missile-dead",
		"next-missile:missile-dead",
		"flags:missile-match",
		"script-id:missile-match",
	}
	return w, missileMatch, wantEvents
}

func TestObjectByScriptID4ECF10SearchOrder(t *testing.T) {
	w, wantObject, wantEvents := newObjectByScriptIDFullSearch4ECF10()
	got := objectByScriptID4ECF10(w.hooks())
	if got != wantObject {
		t.Fatalf("result = %p, want %p", got, wantObject)
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, wantEvents)
	}
}

func TestObjectByScriptID4ECF10FaultOrder(t *testing.T) {
	_, _, wantEvents := newObjectByScriptIDFullSearch4ECF10()
	for faultAt := range wantEvents {
		faultAt++
		t.Run(wantEvents[faultAt-1], func(t *testing.T) {
			w, _, _ := newObjectByScriptIDFullSearch4ECF10()
			w.faultAt = faultAt
			defer func() {
				gotPanic := recover()
				if gotPanic != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", gotPanic, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %v, want %v", w.events, want)
				}
			}()
			objectByScriptID4ECF10(w.hooks())
		})
	}
}

func TestObjectByScriptID4ECF10StopsAtEachDomainMatch(t *testing.T) {
	for _, domain := range []string{"active", "inventory", "pending", "missile"} {
		t.Run(domain, func(t *testing.T) {
			const wanted = int32(-1)
			match := &objectByScriptIDTestObject4ECF10{name: "match", scriptID: wanted}
			owner := &objectByScriptIDTestObject4ECF10{name: "owner", scriptID: 1}
			w := &objectByScriptIDTestWorld4ECF10{wanted: wanted}
			switch domain {
			case "active":
				w.active = match
			case "inventory":
				owner.firstInventory = match
				w.active = owner
			case "pending":
				w.pending = match
			case "missile":
				w.missile = match
			}
			got := objectByScriptID4ECF10(w.hooks())
			if got != match {
				t.Fatalf("result = %p, want %p", got, match)
			}
			last := w.events[len(w.events)-1]
			if last != "script-id:match" {
				t.Fatalf("last event = %q, want match ScriptID load; events = %v", last, w.events)
			}
		})
	}
}

func TestObjectByScriptID4ECF10EmptyDomainsAndArgumentOrder(t *testing.T) {
	w := &objectByScriptIDTestWorld4ECF10{wanted: -1}
	if got := objectByScriptID4ECF10(w.hooks()); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	want := []string{"first-active", "argument", "first-pending", "first-missile"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestObjectByScriptID4ECF10UsesLowDeadBitAndFullSignedID(t *testing.T) {
	match := &objectByScriptIDTestObject4ECF10{name: "match", flags: 0x80, scriptID: -2147483648}
	w := &objectByScriptIDTestWorld4ECF10{active: match, wanted: -2147483648}
	if got := objectByScriptID4ECF10(w.hooks()); got != match {
		t.Fatalf("result = %p, want %p", got, match)
	}
	want := []string{"first-active", "argument", "flags:match", "script-id:match"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
