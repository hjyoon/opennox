package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unusedHealthLinksRemoveWorld4EE3E0 struct {
	healthByObject   map[string]string
	nextByHealth     map[string]string
	previousByHealth map[string]string
	head             string

	events     []string
	faultAt    int
	afterEvent map[string]func()
}

func newUnusedHealthLinksRemoveWorld4EE3E0() *unusedHealthLinksRemoveWorld4EE3E0 {
	return &unusedHealthLinksRemoveWorld4EE3E0{
		healthByObject:   make(map[string]string),
		nextByHealth:     make(map[string]string),
		previousByHealth: make(map[string]string),
		afterEvent:       make(map[string]func()),
	}
}

func (w *unusedHealthLinksRemoveWorld4EE3E0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		delete(w.afterEvent, event)
		after()
	}
}

func (w *unusedHealthLinksRemoveWorld4EE3E0) hooks() unusedHealthLinksRemoveHooks4EE3E0[string, string] {
	return unusedHealthLinksRemoveHooks4EE3E0[string, string]{
		loadHealth: func(obj string) string {
			health := w.healthByObject[obj]
			w.record("load-health:" + obj + "=" + health)
			return health
		},
		loadNext: func(health string) string {
			next := w.nextByHealth[health]
			w.record("load-next:" + health + "=" + next)
			return next
		},
		loadPrevious: func(health string) string {
			previous := w.previousByHealth[health]
			w.record("load-previous:" + health + "=" + previous)
			return previous
		},
		storePrevious: func(health, previous string) {
			w.record("store-previous:" + health + "=" + previous)
			w.previousByHealth[health] = previous
		},
		storeNext: func(health, next string) {
			w.record("store-next:" + health + "=" + next)
			w.nextByHealth[health] = next
		},
		storeHead: func(head string) {
			w.record("store-head:" + head)
			w.head = head
		},
	}
}

func TestUnusedHealthLinksRemove4EE3E0NullGuards(t *testing.T) {
	t.Run("object", func(t *testing.T) {
		w := newUnusedHealthLinksRemoveWorld4EE3E0()
		unusedHealthLinksRemove4EE3E0("", w.hooks())
		if len(w.events) != 0 {
			t.Fatalf("events = %v, want none", w.events)
		}
	})

	t.Run("health", func(t *testing.T) {
		w := newUnusedHealthLinksRemoveWorld4EE3E0()
		unusedHealthLinksRemove4EE3E0("object", w.hooks())
		want := []string{"load-health:object="}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})
}

func TestUnusedHealthLinksRemove4EE3E0AllListPositions(t *testing.T) {
	tests := []struct {
		name       string
		next       string
		previous   string
		wantEvents []string
		wantHead   string
	}{
		{
			name: "isolated-head",
			wantEvents: []string{
				"load-health:object=health",
				"load-next:health=",
				"load-health:object=health",
				"load-previous:health=",
				"load-next:health=",
				"store-head:",
			},
		},
		{
			name:     "head",
			next:     "next",
			wantHead: "next",
			wantEvents: []string{
				"load-health:object=health",
				"load-next:health=next",
				"load-health:next=next-health",
				"load-previous:health=",
				"store-previous:next-health=",
				"load-health:object=health",
				"load-previous:health=",
				"load-next:health=next",
				"store-head:next",
			},
		},
		{
			name:     "middle",
			next:     "next",
			previous: "previous",
			wantHead: "old-head",
			wantEvents: []string{
				"load-health:object=health",
				"load-next:health=next",
				"load-health:next=next-health",
				"load-previous:health=previous",
				"store-previous:next-health=previous",
				"load-health:object=health",
				"load-previous:health=previous",
				"load-health:previous=previous-health",
				"load-next:health=next",
				"store-next:previous-health=next",
			},
		},
		{
			name:     "tail",
			previous: "previous",
			wantHead: "old-head",
			wantEvents: []string{
				"load-health:object=health",
				"load-next:health=",
				"load-health:object=health",
				"load-previous:health=previous",
				"load-health:previous=previous-health",
				"load-next:health=",
				"store-next:previous-health=",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newUnusedHealthLinksRemoveWorld4EE3E0()
			w.healthByObject["object"] = "health"
			w.healthByObject["next"] = "next-health"
			w.healthByObject["previous"] = "previous-health"
			w.nextByHealth["health"] = test.next
			w.previousByHealth["health"] = test.previous
			w.nextByHealth["next-health"] = "keep-next-next"
			w.previousByHealth["next-health"] = "old-next-previous"
			w.nextByHealth["previous-health"] = "old-previous-next"
			w.previousByHealth["previous-health"] = "keep-previous-previous"
			w.head = "old-head"

			unusedHealthLinksRemove4EE3E0("object", w.hooks())

			if !reflect.DeepEqual(w.events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", w.events, test.wantEvents)
			}
			if w.head != test.wantHead {
				t.Errorf("head = %q, want %q", w.head, test.wantHead)
			}
			if test.next != "" && w.previousByHealth["next-health"] != test.previous {
				t.Errorf("next previous = %q, want %q", w.previousByHealth["next-health"], test.previous)
			}
			if test.previous != "" && w.nextByHealth["previous-health"] != test.next {
				t.Errorf("previous next = %q, want %q", w.nextByHealth["previous-health"], test.next)
			}
			if w.nextByHealth["health"] != test.next || w.previousByHealth["health"] != test.previous {
				t.Errorf("removed links changed: next=%q previous=%q", w.nextByHealth["health"], w.previousByHealth["health"])
			}
		})
	}
}

func TestUnusedHealthLinksRemove4EE3E0ReloadsObjectHealthAndLiveLinks(t *testing.T) {
	w := newUnusedHealthLinksRemoveWorld4EE3E0()
	w.healthByObject["object"] = "health-1"
	w.healthByObject["next-1"] = "next-health-1"
	w.healthByObject["previous-2"] = "previous-health-2"
	w.nextByHealth["health-1"] = "next-1"
	w.previousByHealth["health-1"] = "previous-1-old"
	w.nextByHealth["health-2"] = "next-2"
	w.previousByHealth["health-2"] = "previous-2"
	w.afterEvent["load-health:next-1=next-health-1"] = func() {
		w.previousByHealth["health-1"] = "previous-1"
	}
	w.afterEvent["store-previous:next-health-1=previous-1"] = func() {
		w.healthByObject["object"] = "health-2"
	}
	w.afterEvent["load-health:previous-2=previous-health-2"] = func() {
		w.nextByHealth["health-2"] = "next-3"
	}

	unusedHealthLinksRemove4EE3E0("object", w.hooks())

	wantEvents := []string{
		"load-health:object=health-1",
		"load-next:health-1=next-1",
		"load-health:next-1=next-health-1",
		"load-previous:health-1=previous-1",
		"store-previous:next-health-1=previous-1",
		"load-health:object=health-2",
		"load-previous:health-2=previous-2",
		"load-health:previous-2=previous-health-2",
		"load-next:health-2=next-3",
		"store-next:previous-health-2=next-3",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
	if w.previousByHealth["next-health-1"] != "previous-1" || w.nextByHealth["previous-health-2"] != "next-3" {
		t.Fatalf("repairs = previous:%q next:%q", w.previousByHealth["next-health-1"], w.nextByHealth["previous-health-2"])
	}
}

func verifyUnusedHealthLinksRemoveFaultPrefixes4EE3E0(t *testing.T, want []string, build func() *unusedHealthLinksRemoveWorld4EE3E0) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			unusedHealthLinksRemove4EE3E0("object", w.hooks())
		})
	}
}

func TestUnusedHealthLinksRemove4EE3E0FaultPrefixes(t *testing.T) {
	t.Run("middle", func(t *testing.T) {
		want := []string{
			"load-health:object=health",
			"load-next:health=next",
			"load-health:next=next-health",
			"load-previous:health=previous",
			"store-previous:next-health=previous",
			"load-health:object=health",
			"load-previous:health=previous",
			"load-health:previous=previous-health",
			"load-next:health=next",
			"store-next:previous-health=next",
		}
		verifyUnusedHealthLinksRemoveFaultPrefixes4EE3E0(t, want, func() *unusedHealthLinksRemoveWorld4EE3E0 {
			w := newUnusedHealthLinksRemoveWorld4EE3E0()
			w.healthByObject["object"] = "health"
			w.healthByObject["next"] = "next-health"
			w.healthByObject["previous"] = "previous-health"
			w.nextByHealth["health"] = "next"
			w.previousByHealth["health"] = "previous"
			return w
		})
	})

	t.Run("head", func(t *testing.T) {
		want := []string{
			"load-health:object=health",
			"load-next:health=next",
			"load-health:next=next-health",
			"load-previous:health=",
			"store-previous:next-health=",
			"load-health:object=health",
			"load-previous:health=",
			"load-next:health=next",
			"store-head:next",
		}
		verifyUnusedHealthLinksRemoveFaultPrefixes4EE3E0(t, want, func() *unusedHealthLinksRemoveWorld4EE3E0 {
			w := newUnusedHealthLinksRemoveWorld4EE3E0()
			w.healthByObject["object"] = "health"
			w.healthByObject["next"] = "next-health"
			w.nextByHealth["health"] = "next"
			return w
		})
	})
}
