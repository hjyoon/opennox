package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unusedHealthLinksResetWorld4EE390 struct {
	healthByObject   map[string]string
	previousByHealth map[string]string
	nextByHealth     map[string]string
	objectPrevious   map[string]string
	head             string

	events  []string
	faultAt int

	onStorePrevious func()
	onStoreNext     func()
	onNullStore     func()
}

func (w *unusedHealthLinksResetWorld4EE390) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unusedHealthLinksResetWorld4EE390) hooks() unusedHealthLinksResetHooks4EE390[string, string] {
	return unusedHealthLinksResetHooks4EE390[string, string]{
		loadHealth: func(obj string) string {
			health := w.healthByObject[obj]
			w.record("health:" + obj + "=" + health)
			return health
		},
		storeHealthPrevious: func(health, previous string) {
			w.record("previous:" + health + "=" + previous)
			w.previousByHealth[health] = previous
			if w.onStorePrevious != nil {
				w.onStorePrevious()
			}
		},
		storeHealthNext: func(health, next string) {
			w.record("next:" + health + "=" + next)
			w.nextByHealth[health] = next
			if w.onStoreNext != nil {
				w.onStoreNext()
			}
		},
		storeAbsoluteNullPrevious: func() {
			w.record("absolute-null+12=0")
			if w.onNullStore != nil {
				w.onNullStore()
			}
		},
		loadHead: func() string {
			w.record("head=" + w.head)
			return w.head
		},
		storeObjectPrevious: func(obj, previous string) {
			w.record("object-previous:" + obj + "=" + previous)
			w.objectPrevious[obj] = previous
		},
		storeHead: func(obj string) {
			w.record("store-head:" + obj)
			w.head = obj
		},
	}
}

func newUnusedHealthLinksResetWorld4EE390() *unusedHealthLinksResetWorld4EE390 {
	return &unusedHealthLinksResetWorld4EE390{
		healthByObject:   make(map[string]string),
		previousByHealth: make(map[string]string),
		nextByHealth:     make(map[string]string),
		objectPrevious:   make(map[string]string),
	}
}

func TestUnusedHealthLinksReset4EE390NullObject(t *testing.T) {
	w := newUnusedHealthLinksResetWorld4EE390()
	got := unusedHealthLinksReset4EE390("", w.hooks())
	want := unusedHealthLinksResetResult4EE390[string, string]{
		kind: unusedHealthLinksResetObject4EE390,
	}
	if got != want {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if len(w.events) != 0 {
		t.Fatalf("events = %v, want none", w.events)
	}
}

func TestUnusedHealthLinksReset4EE390ReloadsHealthBetweenStores(t *testing.T) {
	w := newUnusedHealthLinksResetWorld4EE390()
	w.healthByObject["object"] = "health-1"
	w.previousByHealth["health-1"] = "old-previous"
	w.nextByHealth["health-1"] = "keep-next"
	w.previousByHealth["health-2"] = "keep-previous"
	w.nextByHealth["health-2"] = "old-next"
	w.onStorePrevious = func() {
		w.healthByObject["object"] = "health-2"
	}

	got := unusedHealthLinksReset4EE390("object", w.hooks())
	want := unusedHealthLinksResetResult4EE390[string, string]{
		kind:   unusedHealthLinksResetHealth4EE390,
		health: "health-2",
	}
	if got != want {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	wantEvents := []string{
		"health:object=health-1",
		"previous:health-1=",
		"health:object=health-2",
		"next:health-2=",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
	if w.previousByHealth["health-1"] != "" || w.nextByHealth["health-1"] != "keep-next" ||
		w.previousByHealth["health-2"] != "keep-previous" || w.nextByHealth["health-2"] != "" {
		t.Fatalf("links = prev1:%q next1:%q prev2:%q next2:%q",
			w.previousByHealth["health-1"], w.nextByHealth["health-1"],
			w.previousByHealth["health-2"], w.nextByHealth["health-2"])
	}
}

func TestUnusedHealthLinksReset4EE390SealsPostFaultInstructionStream(t *testing.T) {
	w := newUnusedHealthLinksResetWorld4EE390()
	w.head = "head-1"
	w.onNullStore = func() {
		w.healthByObject["object"] = "late-health"
	}
	w.onStoreNext = func() {
		w.head = "head-2"
	}

	got := unusedHealthLinksReset4EE390("object", w.hooks())
	want := unusedHealthLinksResetResult4EE390[string, string]{
		kind:   unusedHealthLinksResetObject4EE390,
		object: "object",
	}
	if got != want {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	wantEvents := []string{
		"health:object=",
		"absolute-null+12=0",
		"health:object=late-health",
		"head=head-1",
		"next:late-health=head-1",
		"head=head-2",
		"object-previous:head-2=object",
		"store-head:object",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
	if w.nextByHealth["late-health"] != "head-1" || w.objectPrevious["head-2"] != "object" || w.head != "object" {
		t.Fatalf("post-fault bytes = next:%q previous:%q head:%q",
			w.nextByHealth["late-health"], w.objectPrevious["head-2"], w.head)
	}
}

func TestUnusedHealthLinksReset4EE390FaultPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*unusedHealthLinksResetWorld4EE390)
		wantEvents []string
	}{
		{
			name: "health-present",
			setup: func(w *unusedHealthLinksResetWorld4EE390) {
				w.healthByObject["object"] = "health"
			},
			wantEvents: []string{
				"health:object=health", "previous:health=", "health:object=health", "next:health=",
			},
		},
		{
			name: "health-null-post-fault-bytes",
			setup: func(w *unusedHealthLinksResetWorld4EE390) {
				w.head = "head"
			},
			wantEvents: []string{
				"health:object=", "absolute-null+12=0", "health:object=", "head=head",
				"next:=head", "head=head", "object-previous:head=object", "store-head:object",
			},
		},
	}
	for _, test := range tests {
		for faultAt := 1; faultAt <= len(test.wantEvents); faultAt++ {
			t.Run(fmt.Sprintf("%s/event-%d", test.name, faultAt), func(t *testing.T) {
				w := newUnusedHealthLinksResetWorld4EE390()
				test.setup(w)
				w.faultAt = faultAt
				defer func() {
					if got := recover(); got != test.wantEvents[faultAt-1] {
						t.Fatalf("panic = %v, want %q", got, test.wantEvents[faultAt-1])
					}
					if want := test.wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
						t.Fatalf("events = %v, want %v", w.events, want)
					}
				}()
				unusedHealthLinksReset4EE390("object", w.hooks())
			})
		}
	}
}
