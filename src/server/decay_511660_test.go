package server

import (
	"fmt"
	"reflect"
	"testing"
)

type decayWorld511660 struct {
	setObject  string
	setDelay   uint32
	flags      map[string]uint32
	deadlines  map[string]uint32
	head       string
	next       map[string]string
	holders    map[string]string
	deleteFlag map[string]uint32
	frame      uint32
	events     []string
	faultAt    int
	onDelete   func(*decayWorld511660, string)
}

func newDecayWorld511660() *decayWorld511660 {
	return &decayWorld511660{
		flags:      make(map[string]uint32),
		deadlines:  make(map[string]uint32),
		next:       make(map[string]string),
		holders:    make(map[string]string),
		deleteFlag: make(map[string]uint32),
	}
}

func (w *decayWorld511660) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *decayWorld511660) hooks() decayHooks511660[string] {
	return decayHooks511660[string]{
		loadSetObjectArg: func() string {
			w.record("set-object:" + w.setObject)
			return w.setObject
		},
		loadSetDelayArg: func() uint32 {
			w.record(fmt.Sprintf("set-delay:%08x", w.setDelay))
			return w.setDelay
		},
		loadObjectFlags: func(obj string) uint32 {
			value := w.flags[obj]
			w.record(fmt.Sprintf("flags:%s=%08x", obj, value))
			return value
		},
		storeObjectFlags: func(obj string, value uint32) {
			w.record(fmt.Sprintf("store-flags:%s=%08x", obj, value))
			w.flags[obj] = value
		},
		loadFrame: func() uint32 {
			w.record(fmt.Sprintf("frame:%08x", w.frame))
			return w.frame
		},
		loadDeadline: func(obj string) uint32 {
			value := w.deadlines[obj]
			w.record(fmt.Sprintf("deadline:%s=%08x", obj, value))
			return value
		},
		storeDeadline: func(obj string, value uint32) {
			w.record(fmt.Sprintf("store-deadline:%s=%08x", obj, value))
			w.deadlines[obj] = value
		},
		loadHead: func() string {
			w.record("head:" + w.head)
			return w.head
		},
		storeHead: func(obj string) {
			w.record("store-head:" + obj)
			w.head = obj
		},
		loadNext: func(obj string) string {
			next := w.next[obj]
			w.record("next:" + obj + "=" + next)
			return next
		},
		storeNext: func(obj, next string) {
			w.record("store-next:" + obj + "=" + next)
			w.next[obj] = next
		},
		loadHolder: func(obj string) string {
			holder := w.holders[obj]
			w.record("holder:" + obj + "=" + holder)
			return holder
		},
		loadDeleteFlags: func(obj string) uint32 {
			value := w.deleteFlag[obj]
			w.record(fmt.Sprintf("delete-flags:%s=%08x", obj, value))
			return value
		},
		storeDeleteFlags: func(obj string, value uint32) {
			w.record(fmt.Sprintf("store-delete-flags:%s=%08x", obj, value))
			w.deleteFlag[obj] = value
		},
		delayedDelete: func(obj string) {
			w.record("delayed-delete:" + obj)
			if w.onDelete != nil {
				w.onDelete(w, obj)
			}
		},
	}
}

func verifyDecayFaultPrefixes511660(t *testing.T, want []string, build func() *decayWorld511660, run func(*decayWorld511660)) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			world := build()
			world.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if got := world.events; !reflect.DeepEqual(got, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", got, want[:faultAt])
				}
			}()
			run(world)
		})
	}
}

func TestDecaySetTime511660PendingGate(t *testing.T) {
	build := func() *decayWorld511660 {
		w := newDecayWorld511660()
		w.setObject = "item"
		w.setDelay = 7
		w.flags["item"] = 0x80010001
		return w
	}
	want := []string{"set-object:item", "flags:item=80010001"}
	w := build()
	if got := decaySetTime511660(w.hooks()); got != 0x80010001 {
		t.Fatalf("result = %08x, want 80010001", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decaySetTime511660(w.hooks()) })
}

func TestDecaySetTime511660EmptyWrapAndHeadOrder(t *testing.T) {
	build := func() *decayWorld511660 {
		w := newDecayWorld511660()
		w.setObject = "item"
		w.setDelay = 5
		w.frame = 0xfffffffe
		w.flags["item"] = 0x20
		return w
	}
	want := []string{
		"set-object:item", "flags:item=00000020", "set-delay:00000005", "frame:fffffffe",
		"store-deadline:item=00000003", "head:", "store-head:item", "store-next:item=",
		"flags:item=00000020", "store-flags:item=00400020",
	}
	w := build()
	if got := decaySetTime511660(w.hooks()); got != 0x00400020 {
		t.Fatalf("result = %08x, want 00400020", got)
	}
	if w.head != "item" || w.next["item"] != "" || w.deadlines["item"] != 3 || w.flags["item"] != 0x00400020 {
		t.Fatalf("state = head %q next %q deadline %08x flags %08x", w.head, w.next["item"], w.deadlines["item"], w.flags["item"])
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decaySetTime511660(w.hooks()) })
}

func TestDecaySetTime511660EqualDeadlineGoesAfterAndTailReloadsFlagsFirst(t *testing.T) {
	build := func() *decayWorld511660 {
		w := newDecayWorld511660()
		w.setObject = "item"
		w.setDelay = 5
		w.head = "first"
		w.next["first"] = "second"
		w.deadlines["first"] = 5
		w.deadlines["second"] = 5
		return w
	}
	want := []string{
		"set-object:item", "flags:item=00000000", "set-delay:00000005", "frame:00000000",
		"store-deadline:item=00000005", "head:first", "deadline:first=00000005", "next:first=second",
		"deadline:second=00000005", "next:second=", "store-next:second=item", "flags:item=00000000",
		"store-next:item=", "store-flags:item=00400000",
	}
	w := build()
	if got := decaySetTime511660(w.hooks()); got != decayListedFlag511660 {
		t.Fatalf("result = %08x, want %08x", got, decayListedFlag511660)
	}
	if w.next["first"] != "second" || w.next["second"] != "item" || w.next["item"] != "" {
		t.Fatalf("links = %#v", w.next)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decaySetTime511660(w.hooks()) })
}

func TestDecaySetTime511660RescheduleRemovesBeforeDelayLoad(t *testing.T) {
	build := func() *decayWorld511660 {
		w := newDecayWorld511660()
		w.setObject = "item"
		w.setDelay = 2
		w.frame = 8
		w.flags["item"] = decayListedFlag511660 | 0x40
		w.head = "before"
		w.next["before"] = "item"
		w.next["item"] = "after"
		w.deadlines["before"] = 1
		w.deadlines["after"] = 20
		return w
	}
	want := []string{
		"set-object:item", "flags:item=00400040", "flags:item=00400040", "store-flags:item=00000040",
		"head:before", "next:before=item", "next:item=after", "store-next:before=after",
		"set-delay:00000002", "frame:00000008", "store-deadline:item=0000000a", "head:before",
		"deadline:before=00000001", "next:before=after", "deadline:after=00000014",
		"store-next:before=item", "store-next:item=after", "flags:item=00000040", "store-flags:item=00400040",
	}
	w := build()
	if got := decaySetTime511660(w.hooks()); got != 0x00400040 {
		t.Fatalf("result = %08x, want 00400040", got)
	}
	if w.next["before"] != "item" || w.next["item"] != "after" || w.deadlines["item"] != 10 {
		t.Fatalf("state = links %#v deadline %d", w.next, w.deadlines["item"])
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decaySetTime511660(w.hooks()) })
}

func TestDecayRemove5116F0ReturnDomainsAndLinkPreservation(t *testing.T) {
	t.Run("unlisted-flags", func(t *testing.T) {
		w := newDecayWorld511660()
		w.flags["item"] = 0x80000040
		w.next["item"] = "stale"
		got := decayRemove5116F0("item", w.hooks())
		want := decayRemoveResult5116F0[string]{kind: decayRemoveWord5116F0, word: 0x80000040}
		if got != want || w.next["item"] != "stale" {
			t.Fatalf("result/state = %#v next %q", got, w.next["item"])
		}
		if events := []string{"flags:item=80000040"}; !reflect.DeepEqual(w.events, events) {
			t.Fatalf("events = %v, want %v", w.events, events)
		}
	})

	t.Run("head-match-returns-item", func(t *testing.T) {
		build := func() *decayWorld511660 {
			w := newDecayWorld511660()
			w.flags["item"] = decayListedFlag511660 | 0x20
			w.head = "item"
			w.next["item"] = "tail"
			return w
		}
		wantEvents := []string{
			"flags:item=00400020", "store-flags:item=00000020", "head:item",
			"next:item=tail", "store-head:tail",
		}
		w := build()
		got := decayRemove5116F0("item", w.hooks())
		want := decayRemoveResult5116F0[string]{kind: decayRemoveObject5116F0, object: "item"}
		if got != want || w.head != "tail" || w.next["item"] != "tail" {
			t.Fatalf("result/state = %#v head %q next %q", got, w.head, w.next["item"])
		}
		if !reflect.DeepEqual(w.events, wantEvents) {
			t.Fatalf("events = %v, want %v", w.events, wantEvents)
		}
		verifyDecayFaultPrefixes511660(t, wantEvents, build, func(w *decayWorld511660) { decayRemove5116F0("item", w.hooks()) })
	})

	t.Run("later-match-returns-next", func(t *testing.T) {
		build := func() *decayWorld511660 {
			w := newDecayWorld511660()
			w.flags["item"] = decayListedFlag511660
			w.head = "first"
			w.next["first"] = "item"
			w.next["item"] = "tail"
			return w
		}
		wantEvents := []string{
			"flags:item=00400000", "store-flags:item=00000000", "head:first",
			"next:first=item", "next:item=tail", "store-next:first=tail",
		}
		w := build()
		got := decayRemove5116F0("item", w.hooks())
		want := decayRemoveResult5116F0[string]{kind: decayRemoveObject5116F0, object: "tail"}
		if got != want || w.next["first"] != "tail" || w.next["item"] != "tail" {
			t.Fatalf("result/state = %#v links %#v", got, w.next)
		}
		if !reflect.DeepEqual(w.events, wantEvents) {
			t.Fatalf("events = %v, want %v", w.events, wantEvents)
		}
		verifyDecayFaultPrefixes511660(t, wantEvents, build, func(w *decayWorld511660) { decayRemove5116F0("item", w.hooks()) })
	})

	t.Run("listed-but-missing", func(t *testing.T) {
		w := newDecayWorld511660()
		w.flags["item"] = decayListedFlag511660
		w.head = "first"
		w.next["first"] = ""
		got := decayRemove5116F0("item", w.hooks())
		want := decayRemoveResult5116F0[string]{kind: decayRemoveWord5116F0}
		if got != want || w.flags["item"] != 0 || w.head != "first" {
			t.Fatalf("result/state = %#v flags %08x head %q", got, w.flags["item"], w.head)
		}
		wantEvents := []string{
			"flags:item=00400000", "store-flags:item=00000000", "head:first", "next:first=",
		}
		if !reflect.DeepEqual(w.events, wantEvents) {
			t.Fatalf("events = %v, want %v", w.events, wantEvents)
		}
	})
}

func TestDecayTick511750FutureStopsAfterCachedNext(t *testing.T) {
	build := func() *decayWorld511660 {
		w := newDecayWorld511660()
		w.head = "future"
		w.next["future"] = "unread-tail"
		w.deadlines["future"] = 11
		w.frame = 10
		return w
	}
	want := []string{
		"head:future", "holder:future=", "next:future=unread-tail",
		"deadline:future=0000000b", "frame:0000000a",
	}
	w := build()
	decayTick511750(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decayTick511750(w.hooks()) })
}

func TestDecayTick511750HolderRemovalAndDueDeleteUseCachedNext(t *testing.T) {
	build := func() *decayWorld511660 {
		w := newDecayWorld511660()
		w.head = "held"
		w.next["held"] = "due"
		w.next["due"] = ""
		w.holders["held"] = "owner"
		w.flags["held"] = decayListedFlag511660 | 1
		w.flags["due"] = decayListedFlag511660 | 2
		w.deadlines["due"] = 9
		w.frame = 9
		w.deleteFlag["due"] = 0x12340001
		w.onDelete = func(w *decayWorld511660, _ string) {
			w.next["due"] = "mutated"
		}
		return w
	}
	want := []string{
		"head:held", "holder:held=owner", "next:held=due",
		"flags:held=00400001", "store-flags:held=00000001", "head:held", "next:held=due", "store-head:due",
		"holder:due=", "next:due=", "deadline:due=00000009", "frame:00000009",
		"flags:due=00400002", "store-flags:due=00000002", "head:due", "next:due=", "store-head:",
		"delete-flags:due=12340001", "store-delete-flags:due=12340081", "delayed-delete:due",
	}
	w := build()
	decayTick511750(w.hooks())
	if w.head != "" || w.flags["held"] != 1 || w.flags["due"] != 2 || w.deleteFlag["due"] != 0x12340081 {
		t.Fatalf("state = head %q flags %#v delete %08x", w.head, w.flags, w.deleteFlag["due"])
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decayTick511750(w.hooks()) })
}

func TestDecayDestroy5117B0CachesNextAndAlwaysClearsHead(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		build := func() *decayWorld511660 { return newDecayWorld511660() }
		want := []string{"head:", "store-head:"}
		w := build()
		if got := decayDestroy5117B0(w.hooks()); got != 0 {
			t.Fatalf("result = %08x", got)
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
		verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decayDestroy5117B0(w.hooks()) })
	})

	t.Run("two-items", func(t *testing.T) {
		build := func() *decayWorld511660 {
			w := newDecayWorld511660()
			w.head = "first"
			w.next["first"] = "second"
			w.flags["first"] = decayListedFlag511660 | 1
			w.flags["second"] = decayListedFlag511660 | 2
			return w
		}
		want := []string{
			"head:first", "next:first=second",
			"flags:first=00400001", "store-flags:first=00000001", "head:first", "next:first=second", "store-head:second",
			"next:second=", "flags:second=00400002", "store-flags:second=00000002", "head:second", "next:second=", "store-head:",
			"store-head:",
		}
		w := build()
		if got := decayDestroy5117B0(w.hooks()); got != 0 {
			t.Fatalf("result = %08x", got)
		}
		if w.head != "" || w.flags["first"] != 1 || w.flags["second"] != 2 || w.next["first"] != "second" {
			t.Fatalf("state = head %q flags %#v next %#v", w.head, w.flags, w.next)
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
		verifyDecayFaultPrefixes511660(t, want, build, func(w *decayWorld511660) { decayDestroy5117B0(w.hooks()) })
	})
}
