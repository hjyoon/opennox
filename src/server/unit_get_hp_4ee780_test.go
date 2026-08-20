package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitGetHPTestHealth4EE780 struct {
	name    string
	current uint16
}

type unitGetHPTestObject4EE780 struct {
	name   string
	health *unitGetHPTestHealth4EE780
}

type unitGetHPTestWorld4EE780 struct {
	unit    *unitGetHPTestObject4EE780
	events  []string
	faultAt int

	afterHealth  func(*unitGetHPTestWorld4EE780, *unitGetHPTestHealth4EE780)
	afterCurrent func(*unitGetHPTestWorld4EE780, *unitGetHPTestHealth4EE780)
}

func unitGetHPObjectName4EE780(obj *unitGetHPTestObject4EE780) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func unitGetHPHealthName4EE780(health *unitGetHPTestHealth4EE780) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func (w *unitGetHPTestWorld4EE780) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitGetHPTestWorld4EE780) hooks() unitGetHPHooks4EE780[
	*unitGetHPTestObject4EE780,
	*unitGetHPTestHealth4EE780,
] {
	return unitGetHPHooks4EE780[
		*unitGetHPTestObject4EE780,
		*unitGetHPTestHealth4EE780,
	]{
		loadUnitArg: func() *unitGetHPTestObject4EE780 {
			w.record("arg:" + unitGetHPObjectName4EE780(w.unit))
			return w.unit
		},
		loadHealth: func(obj *unitGetHPTestObject4EE780) *unitGetHPTestHealth4EE780 {
			health := obj.health
			w.record("health:" + unitGetHPObjectName4EE780(obj) + "=" + unitGetHPHealthName4EE780(health))
			if w.afterHealth != nil {
				w.afterHealth(w, health)
			}
			return health
		},
		loadCurrent: func(health *unitGetHPTestHealth4EE780) uint16 {
			current := health.current
			w.record(fmt.Sprintf("current:%s=%#04x", unitGetHPHealthName4EE780(health), current))
			if w.afterCurrent != nil {
				w.afterCurrent(w, health)
			}
			return current
		},
	}
}

func newUnitGetHPTestWorld4EE780() *unitGetHPTestWorld4EE780 {
	return &unitGetHPTestWorld4EE780{
		unit: &unitGetHPTestObject4EE780{
			name:   "unit",
			health: &unitGetHPTestHealth4EE780{name: "health", current: 0x4321},
		},
	}
}

func TestUnitGetHP4EE780EntryGates(t *testing.T) {
	tests := []struct {
		name string
		edit func(*unitGetHPTestWorld4EE780)
		want []string
	}{
		{
			name: "nil unit",
			edit: func(w *unitGetHPTestWorld4EE780) { w.unit = nil },
			want: []string{"arg:nil"},
		},
		{
			name: "nil health",
			edit: func(w *unitGetHPTestWorld4EE780) { w.unit.health = nil },
			want: []string{"arg:unit", "health:unit=nil"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newUnitGetHPTestWorld4EE780()
			test.edit(w)
			if got := unitGetHP4EE780(w.hooks()); got != 0 {
				t.Fatalf("result = %#04x, want zero", got)
			}
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %q, want %q", w.events, test.want)
			}
		})
	}
}

func TestUnitGetHP4EE780PreservesEveryCurrentWord(t *testing.T) {
	for _, current := range []uint16{0, 1, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", current), func(t *testing.T) {
			w := newUnitGetHPTestWorld4EE780()
			w.unit.health.current = current
			if got := unitGetHP4EE780(w.hooks()); got != current {
				t.Fatalf("result = %#04x, want %#04x", got, current)
			}
		})
	}
}

func TestUnitGetHP4EE780CachesHealthAndCurrent(t *testing.T) {
	w := newUnitGetHPTestWorld4EE780()
	original := w.unit.health
	replacement := &unitGetHPTestHealth4EE780{name: "replacement", current: 0xabcd}
	w.afterHealth = func(w *unitGetHPTestWorld4EE780, _ *unitGetHPTestHealth4EE780) {
		w.unit.health = replacement
	}
	w.afterCurrent = func(_ *unitGetHPTestWorld4EE780, health *unitGetHPTestHealth4EE780) {
		health.current = 0x9876
	}

	if got := unitGetHP4EE780(w.hooks()); got != 0x4321 {
		t.Fatalf("result = %#04x, want cached 0x4321", got)
	}
	if original.current != 0x9876 || w.unit.health != replacement {
		t.Fatalf("mutations were not observed: original=%#04x live=%p replacement=%p", original.current, w.unit.health, replacement)
	}
	want := []string{"arg:unit", "health:unit=health", "current:health=0x4321"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestUnitGetHP4EE780AllFaultPrefixes(t *testing.T) {
	want := []string{"arg:unit", "health:unit=health", "current:health=0x4321"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newUnitGetHPTestWorld4EE780()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
			}()
			unitGetHP4EE780(w.hooks())
		})
	}
}
