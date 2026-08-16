package server

import (
	"fmt"
	"reflect"
	"testing"
)

type netCodeCacheRemoveObjectWorld4ECFA0 struct {
	needsInit     uint32
	firstUsed     string
	objectArg     string
	entryObjects  map[string]string
	entryNext     map[string]string
	prependResult string
	events        []string
	faultAt       int
}

func (w *netCodeCacheRemoveObjectWorld4ECFA0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *netCodeCacheRemoveObjectWorld4ECFA0) hooks() netCodeCacheRemoveObjectHooks4ECFA0[string, string] {
	return netCodeCacheRemoveObjectHooks4ECFA0[string, string]{
		loadNeedsInit: func() uint32 {
			w.record(fmt.Sprintf("needs-init:%08x", w.needsInit))
			return w.needsInit
		},
		loadFirstUsed: func() string {
			w.record("first-used:" + w.firstUsed)
			return w.firstUsed
		},
		loadObjectArg: func() string {
			w.record("object-arg:" + w.objectArg)
			return w.objectArg
		},
		loadEntryObject: func(entry string) string {
			obj := w.entryObjects[entry]
			w.record("entry-object:" + entry + "=" + obj)
			return obj
		},
		loadEntryNext: func(entry string) string {
			next := w.entryNext[entry]
			w.record("entry-next:" + entry + "=" + next)
			return next
		},
		removeUsed: func(entry string) {
			w.record("remove-used:" + entry)
		},
		prependFree: func(entry string) string {
			w.record("prepend-free:" + entry)
			return w.prependResult
		},
	}
}

func netCodeCacheRemoveObjectCases4ECFA0() []struct {
	name       string
	world      netCodeCacheRemoveObjectWorld4ECFA0
	wantResult netCodeCacheRemoveObjectResult4ECFA0[string, string]
	wantEvents []string
} {
	return []struct {
		name       string
		world      netCodeCacheRemoveObjectWorld4ECFA0
		wantResult netCodeCacheRemoveObjectResult4ECFA0[string, string]
		wantEvents []string
	}{
		{
			name:  "needs-initialization",
			world: netCodeCacheRemoveObjectWorld4ECFA0{needsInit: 0x80000001},
			wantResult: netCodeCacheRemoveObjectResult4ECFA0[string, string]{
				kind:        netCodeCacheRemoveObjectInitial4ECFA0,
				initialFlag: 0x80000001,
			},
			wantEvents: []string{"needs-init:80000001"},
		},
		{
			name: "initialized-empty",
			world: netCodeCacheRemoveObjectWorld4ECFA0{
				entryObjects: map[string]string{},
				entryNext:    map[string]string{},
			},
			wantResult: netCodeCacheRemoveObjectResult4ECFA0[string, string]{
				kind: netCodeCacheRemoveObjectInitial4ECFA0,
			},
			wantEvents: []string{"needs-init:00000000", "first-used:"},
		},
		{
			name: "first-entry-match",
			world: netCodeCacheRemoveObjectWorld4ECFA0{
				firstUsed:     "entry-1",
				objectArg:     "target",
				entryObjects:  map[string]string{"entry-1": "target"},
				entryNext:     map[string]string{"entry-1": "unread"},
				prependResult: "free-result",
			},
			wantResult: netCodeCacheRemoveObjectResult4ECFA0[string, string]{
				kind:  netCodeCacheRemoveObjectEntry4ECFA0,
				entry: "free-result",
			},
			wantEvents: []string{
				"needs-init:00000000", "first-used:entry-1", "object-arg:target",
				"entry-object:entry-1=target", "remove-used:entry-1", "prepend-free:entry-1",
			},
		},
		{
			name: "later-entry-match",
			world: netCodeCacheRemoveObjectWorld4ECFA0{
				firstUsed:     "entry-1",
				objectArg:     "target",
				entryObjects:  map[string]string{"entry-1": "other-1", "entry-2": "other-2", "entry-3": "target"},
				entryNext:     map[string]string{"entry-1": "entry-2", "entry-2": "entry-3", "entry-3": "unread"},
				prependResult: "free-result",
			},
			wantResult: netCodeCacheRemoveObjectResult4ECFA0[string, string]{
				kind:  netCodeCacheRemoveObjectEntry4ECFA0,
				entry: "free-result",
			},
			wantEvents: []string{
				"needs-init:00000000", "first-used:entry-1", "object-arg:target",
				"entry-object:entry-1=other-1", "entry-next:entry-1=entry-2",
				"entry-object:entry-2=other-2", "entry-next:entry-2=entry-3",
				"entry-object:entry-3=target", "remove-used:entry-3", "prepend-free:entry-3",
			},
		},
		{
			name: "not-found",
			world: netCodeCacheRemoveObjectWorld4ECFA0{
				firstUsed:    "entry-1",
				objectArg:    "target",
				entryObjects: map[string]string{"entry-1": "other-1", "entry-2": "other-2"},
				entryNext:    map[string]string{"entry-1": "entry-2"},
			},
			wantResult: netCodeCacheRemoveObjectResult4ECFA0[string, string]{
				kind:   netCodeCacheRemoveObjectArgument4ECFA0,
				object: "target",
			},
			wantEvents: []string{
				"needs-init:00000000", "first-used:entry-1", "object-arg:target",
				"entry-object:entry-1=other-1", "entry-next:entry-1=entry-2",
				"entry-object:entry-2=other-2", "entry-next:entry-2=",
			},
		},
		{
			name: "null-object-match",
			world: netCodeCacheRemoveObjectWorld4ECFA0{
				firstUsed:     "entry-1",
				entryObjects:  map[string]string{"entry-1": ""},
				entryNext:     map[string]string{"entry-1": "unread"},
				prependResult: "free-result",
			},
			wantResult: netCodeCacheRemoveObjectResult4ECFA0[string, string]{
				kind:  netCodeCacheRemoveObjectEntry4ECFA0,
				entry: "free-result",
			},
			wantEvents: []string{
				"needs-init:00000000", "first-used:entry-1", "object-arg:",
				"entry-object:entry-1=", "remove-used:entry-1", "prepend-free:entry-1",
			},
		},
	}
}

func TestNetCodeCacheRemoveObject4ECFA0OrderAndReturn(t *testing.T) {
	for _, test := range netCodeCacheRemoveObjectCases4ECFA0() {
		t.Run(test.name, func(t *testing.T) {
			world := test.world
			got := netCodeCacheRemoveObject4ECFA0(world.hooks())
			if got != test.wantResult {
				t.Fatalf("result = %#v, want %#v", got, test.wantResult)
			}
			if !reflect.DeepEqual(world.events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", world.events, test.wantEvents)
			}
		})
	}
}

func TestNetCodeCacheRemoveObject4ECFA0FaultOrder(t *testing.T) {
	for _, test := range netCodeCacheRemoveObjectCases4ECFA0() {
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
				netCodeCacheRemoveObject4ECFA0(world.hooks())
			})
		}
	}
}
