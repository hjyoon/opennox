package server

import (
	"fmt"
	"reflect"
	"testing"
)

type netCodeCacheClearWorld4ECFE0 struct {
	needsInit     uint32
	firstUsed     string
	entryNext     map[string]string
	prependResult map[string]string
	events        []string
	faultAt       int
	onRemove      func(*netCodeCacheClearWorld4ECFE0, string)
	onPrepend     func(*netCodeCacheClearWorld4ECFE0, string)
}

func (w *netCodeCacheClearWorld4ECFE0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *netCodeCacheClearWorld4ECFE0) hooks() netCodeCacheClearHooks4ECFE0[string] {
	return netCodeCacheClearHooks4ECFE0[string]{
		loadNeedsInit: func() uint32 {
			w.record(fmt.Sprintf("needs-init:%08x", w.needsInit))
			return w.needsInit
		},
		loadFirstUsed: func() string {
			w.record("first-used:" + w.firstUsed)
			return w.firstUsed
		},
		loadEntryNext: func(entry string) string {
			next := w.entryNext[entry]
			w.record("entry-next:" + entry + "=" + next)
			return next
		},
		removeUsed: func(entry string) {
			w.record("remove-used:" + entry)
			if w.onRemove != nil {
				w.onRemove(w, entry)
			}
		},
		prependFree: func(entry string) string {
			w.record("prepend-free:" + entry)
			if w.onPrepend != nil {
				w.onPrepend(w, entry)
			}
			return w.prependResult[entry]
		},
	}
}

func netCodeCacheClearCases4ECFE0() []struct {
	name       string
	world      netCodeCacheClearWorld4ECFE0
	wantResult netCodeCacheClearResult4ECFE0[string]
	wantEvents []string
} {
	return []struct {
		name       string
		world      netCodeCacheClearWorld4ECFE0
		wantResult netCodeCacheClearResult4ECFE0[string]
		wantEvents []string
	}{
		{
			name:  "needs-initialization",
			world: netCodeCacheClearWorld4ECFE0{needsInit: 0x80000001},
			wantResult: netCodeCacheClearResult4ECFE0[string]{
				kind:        netCodeCacheClearInitial4ECFE0,
				initialFlag: 0x80000001,
			},
			wantEvents: []string{"needs-init:80000001"},
		},
		{
			name: "initialized-empty",
			world: netCodeCacheClearWorld4ECFE0{
				entryNext:     map[string]string{},
				prependResult: map[string]string{},
			},
			wantResult: netCodeCacheClearResult4ECFE0[string]{
				kind: netCodeCacheClearInitial4ECFE0,
			},
			wantEvents: []string{"needs-init:00000000", "first-used:"},
		},
		{
			name: "one-entry",
			world: netCodeCacheClearWorld4ECFE0{
				firstUsed:     "entry-1",
				entryNext:     map[string]string{"entry-1": ""},
				prependResult: map[string]string{"entry-1": "free-result-1"},
			},
			wantResult: netCodeCacheClearResult4ECFE0[string]{
				kind:  netCodeCacheClearEntry4ECFE0,
				entry: "free-result-1",
			},
			wantEvents: []string{
				"needs-init:00000000", "first-used:entry-1",
				"entry-next:entry-1=", "remove-used:entry-1", "prepend-free:entry-1",
			},
		},
		{
			name: "three-entries",
			world: netCodeCacheClearWorld4ECFE0{
				firstUsed: "entry-1",
				entryNext: map[string]string{
					"entry-1": "entry-2",
					"entry-2": "entry-3",
					"entry-3": "",
				},
				prependResult: map[string]string{
					"entry-1": "free-result-1",
					"entry-2": "free-result-2",
					"entry-3": "free-result-3",
				},
			},
			wantResult: netCodeCacheClearResult4ECFE0[string]{
				kind:  netCodeCacheClearEntry4ECFE0,
				entry: "free-result-3",
			},
			wantEvents: []string{
				"needs-init:00000000", "first-used:entry-1",
				"entry-next:entry-1=entry-2", "remove-used:entry-1", "prepend-free:entry-1",
				"entry-next:entry-2=entry-3", "remove-used:entry-2", "prepend-free:entry-2",
				"entry-next:entry-3=", "remove-used:entry-3", "prepend-free:entry-3",
			},
		},
	}
}

func TestNetCodeCacheClear4ECFE0OrderAndReturn(t *testing.T) {
	for _, test := range netCodeCacheClearCases4ECFE0() {
		t.Run(test.name, func(t *testing.T) {
			world := test.world
			got := netCodeCacheClear4ECFE0(world.hooks())
			if got != test.wantResult {
				t.Fatalf("result = %#v, want %#v", got, test.wantResult)
			}
			if !reflect.DeepEqual(world.events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", world.events, test.wantEvents)
			}
		})
	}
}

func TestNetCodeCacheClear4ECFE0UsesCachedNext(t *testing.T) {
	world := netCodeCacheClearWorld4ECFE0{
		firstUsed: "entry-1",
		entryNext: map[string]string{
			"entry-1": "entry-2",
			"entry-2": "",
		},
		prependResult: map[string]string{
			"entry-1": "free-result-1",
			"entry-2": "free-result-2",
		},
		onRemove: func(world *netCodeCacheClearWorld4ECFE0, entry string) {
			if entry == "entry-1" {
				world.entryNext[entry] = "poison-after-remove"
			}
		},
		onPrepend: func(world *netCodeCacheClearWorld4ECFE0, entry string) {
			if entry == "entry-1" {
				world.firstUsed = "poison-after-prepend"
			}
		},
	}
	got := netCodeCacheClear4ECFE0(world.hooks())
	if got.kind != netCodeCacheClearEntry4ECFE0 || got.entry != "free-result-2" {
		t.Fatalf("result = %#v, want final cached-successor prepend", got)
	}
	wantEvents := []string{
		"needs-init:00000000", "first-used:entry-1",
		"entry-next:entry-1=entry-2", "remove-used:entry-1", "prepend-free:entry-1",
		"entry-next:entry-2=", "remove-used:entry-2", "prepend-free:entry-2",
	}
	if !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("events = %v, want %v", world.events, wantEvents)
	}
}

func TestNetCodeCacheClear4ECFE0FaultOrder(t *testing.T) {
	for _, test := range netCodeCacheClearCases4ECFE0() {
		for faultAt := 1; faultAt <= len(test.wantEvents); faultAt++ {
			t.Run(fmt.Sprintf("%s/event-%d", test.name, faultAt), func(t *testing.T) {
				world := test.world
				world.faultAt = faultAt
				defer func() {
					if got := recover(); got != test.wantEvents[faultAt-1] {
						t.Fatalf("panic = %v, want %q", got, test.wantEvents[faultAt-1])
					}
					if want := test.wantEvents[:faultAt]; !reflect.DeepEqual(world.events, want) {
						t.Fatalf("events = %v, want %v", world.events, want)
					}
				}()
				netCodeCacheClear4ECFE0(world.hooks())
			})
		}
	}
}
