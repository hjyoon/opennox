package server

import (
	"fmt"
	"reflect"
	"testing"
)

type healthLinksNextWorld4EE440 struct {
	healthByObject map[string]string
	nextByHealth   map[string]string
	events         []string
	faultAt        int
	afterEvent     map[string]func()
}

func newHealthLinksNextWorld4EE440() *healthLinksNextWorld4EE440 {
	return &healthLinksNextWorld4EE440{
		healthByObject: make(map[string]string),
		nextByHealth:   make(map[string]string),
		afterEvent:     make(map[string]func()),
	}
}

func (w *healthLinksNextWorld4EE440) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		delete(w.afterEvent, event)
		after()
	}
}

func (w *healthLinksNextWorld4EE440) hooks() healthLinksNextHooks4EE440[string, string] {
	return healthLinksNextHooks4EE440[string, string]{
		loadHealth: func(obj string) string {
			health := w.healthByObject[obj]
			w.record("load-health:" + obj + "=" + health)
			return health
		},
		loadNext: func(health string) string {
			next := w.nextByHealth[health]
			w.record("load-next:" + health + "=" + next)
			if health == "" {
				panic("nil-health-next")
			}
			return next
		},
	}
}

func TestHealthLinksNext4EE440NullObjectDoesNotReadMemory(t *testing.T) {
	w := newHealthLinksNextWorld4EE440()
	if got := healthLinksNext4EE440("", w.hooks()); got != "" {
		t.Fatalf("result = %q, want null", got)
	}
	if len(w.events) != 0 {
		t.Fatalf("events = %v, want none", w.events)
	}
}

func TestHealthLinksNext4EE440LoadsHealthAndReturnsCachedNext(t *testing.T) {
	w := newHealthLinksNextWorld4EE440()
	w.healthByObject["object"] = "health-1"
	w.nextByHealth["health-1"] = "next-1"
	w.afterEvent["load-health:object=health-1"] = func() {
		w.healthByObject["object"] = "health-2"
	}
	w.afterEvent["load-next:health-1=next-1"] = func() {
		w.nextByHealth["health-1"] = "next-2"
	}

	if got := healthLinksNext4EE440("object", w.hooks()); got != "next-1" {
		t.Fatalf("result = %q, want next-1", got)
	}
	want := []string{"load-health:object=health-1", "load-next:health-1=next-1"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestHealthLinksNext4EE440NullHealthFaultsAtNextRead(t *testing.T) {
	w := newHealthLinksNextWorld4EE440()
	defer func() {
		if got := recover(); got != "nil-health-next" {
			t.Fatalf("panic = %v, want nil-health-next", got)
		}
		want := []string{"load-health:object=", "load-next:="}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	healthLinksNext4EE440("object", w.hooks())
}

func TestHealthLinksNext4EE440FaultPrefixes(t *testing.T) {
	want := []string{"load-health:object=health", "load-next:health=next"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newHealthLinksNextWorld4EE440()
			w.healthByObject["object"] = "health"
			w.nextByHealth["health"] = "next"
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			healthLinksNext4EE440("object", w.hooks())
		})
	}
}
