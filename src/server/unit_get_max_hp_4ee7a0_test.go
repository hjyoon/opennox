package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitGetMaxHPTestHealth4EE7A0 struct {
	name    string
	maximum uint16
}

type unitGetMaxHPTestObject4EE7A0 struct {
	name   string
	health *unitGetMaxHPTestHealth4EE7A0
}

type unitGetMaxHPTestWorld4EE7A0 struct {
	unit    *unitGetMaxHPTestObject4EE7A0
	events  []string
	faultAt int

	afterHealth  func(*unitGetMaxHPTestWorld4EE7A0, *unitGetMaxHPTestHealth4EE7A0)
	afterMaximum func(*unitGetMaxHPTestWorld4EE7A0, *unitGetMaxHPTestHealth4EE7A0)
}

func unitGetMaxHPObjectName4EE7A0(obj *unitGetMaxHPTestObject4EE7A0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func unitGetMaxHPHealthName4EE7A0(health *unitGetMaxHPTestHealth4EE7A0) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func (w *unitGetMaxHPTestWorld4EE7A0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitGetMaxHPTestWorld4EE7A0) hooks() unitGetMaxHPHooks4EE7A0[
	*unitGetMaxHPTestObject4EE7A0,
	*unitGetMaxHPTestHealth4EE7A0,
] {
	return unitGetMaxHPHooks4EE7A0[
		*unitGetMaxHPTestObject4EE7A0,
		*unitGetMaxHPTestHealth4EE7A0,
	]{
		loadUnitArg: func() *unitGetMaxHPTestObject4EE7A0 {
			w.record("arg:" + unitGetMaxHPObjectName4EE7A0(w.unit))
			return w.unit
		},
		loadHealth: func(obj *unitGetMaxHPTestObject4EE7A0) *unitGetMaxHPTestHealth4EE7A0 {
			health := obj.health
			w.record("health:" + unitGetMaxHPObjectName4EE7A0(obj) + "=" + unitGetMaxHPHealthName4EE7A0(health))
			if w.afterHealth != nil {
				w.afterHealth(w, health)
			}
			return health
		},
		loadMaximum: func(health *unitGetMaxHPTestHealth4EE7A0) uint16 {
			maximum := health.maximum
			w.record(fmt.Sprintf("maximum:%s=%#04x", unitGetMaxHPHealthName4EE7A0(health), maximum))
			if w.afterMaximum != nil {
				w.afterMaximum(w, health)
			}
			return maximum
		},
	}
}

func newUnitGetMaxHPTestWorld4EE7A0() *unitGetMaxHPTestWorld4EE7A0 {
	return &unitGetMaxHPTestWorld4EE7A0{
		unit: &unitGetMaxHPTestObject4EE7A0{
			name:   "unit",
			health: &unitGetMaxHPTestHealth4EE7A0{name: "health", maximum: 0x4321},
		},
	}
}

func TestUnitGetMaxHP4EE7A0EntryGates(t *testing.T) {
	tests := []struct {
		name string
		edit func(*unitGetMaxHPTestWorld4EE7A0)
		want []string
	}{
		{
			name: "nil unit",
			edit: func(w *unitGetMaxHPTestWorld4EE7A0) { w.unit = nil },
			want: []string{"arg:nil"},
		},
		{
			name: "nil health",
			edit: func(w *unitGetMaxHPTestWorld4EE7A0) { w.unit.health = nil },
			want: []string{"arg:unit", "health:unit=nil"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newUnitGetMaxHPTestWorld4EE7A0()
			test.edit(w)
			if got := unitGetMaxHP4EE7A0(w.hooks()); got != 0 {
				t.Fatalf("result = %#04x, want zero", got)
			}
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %q, want %q", w.events, test.want)
			}
		})
	}
}

func TestUnitGetMaxHP4EE7A0PreservesEveryMaximumWord(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", maximum), func(t *testing.T) {
			w := newUnitGetMaxHPTestWorld4EE7A0()
			w.unit.health.maximum = maximum
			if got := unitGetMaxHP4EE7A0(w.hooks()); got != maximum {
				t.Fatalf("result = %#04x, want %#04x", got, maximum)
			}
		})
	}
}

func TestUnitGetMaxHP4EE7A0CachesHealthAndMaximum(t *testing.T) {
	w := newUnitGetMaxHPTestWorld4EE7A0()
	original := w.unit.health
	replacement := &unitGetMaxHPTestHealth4EE7A0{name: "replacement", maximum: 0xabcd}
	w.afterHealth = func(w *unitGetMaxHPTestWorld4EE7A0, _ *unitGetMaxHPTestHealth4EE7A0) {
		w.unit.health = replacement
	}
	w.afterMaximum = func(_ *unitGetMaxHPTestWorld4EE7A0, health *unitGetMaxHPTestHealth4EE7A0) {
		health.maximum = 0x9876
	}

	if got := unitGetMaxHP4EE7A0(w.hooks()); got != 0x4321 {
		t.Fatalf("result = %#04x, want cached 0x4321", got)
	}
	if original.maximum != 0x9876 || w.unit.health != replacement {
		t.Fatalf("mutations were not observed: original=%#04x live=%p replacement=%p", original.maximum, w.unit.health, replacement)
	}
	want := []string{"arg:unit", "health:unit=health", "maximum:health=0x4321"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestUnitGetMaxHP4EE7A0AllFaultPrefixes(t *testing.T) {
	want := []string{"arg:unit", "health:unit=health", "maximum:health=0x4321"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newUnitGetMaxHPTestWorld4EE7A0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
			}()
			unitGetMaxHP4EE7A0(w.hooks())
		})
	}
}
