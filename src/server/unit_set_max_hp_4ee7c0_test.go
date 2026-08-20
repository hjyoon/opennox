package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitSetMaxHPTestHealth4EE7C0 struct {
	name    string
	maximum uint16
}

type unitSetMaxHPTestObject4EE7C0 struct {
	name   string
	health *unitSetMaxHPTestHealth4EE7C0
}

type unitSetMaxHPTestWorld4EE7C0 struct {
	unit       *unitSetMaxHPTestObject4EE7C0
	maximumArg uint16
	events     []string
	faultAt    int

	afterUnit    func(*unitSetMaxHPTestWorld4EE7C0, *unitSetMaxHPTestObject4EE7C0)
	afterHealth  func(*unitSetMaxHPTestWorld4EE7C0, *unitSetMaxHPTestHealth4EE7C0)
	afterMaximum func(*unitSetMaxHPTestWorld4EE7C0, uint16)
	afterStore   func(*unitSetMaxHPTestWorld4EE7C0, *unitSetMaxHPTestHealth4EE7C0, uint16)
}

func unitSetMaxHPObjectName4EE7C0(obj *unitSetMaxHPTestObject4EE7C0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func unitSetMaxHPHealthName4EE7C0(health *unitSetMaxHPTestHealth4EE7C0) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func (w *unitSetMaxHPTestWorld4EE7C0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitSetMaxHPTestWorld4EE7C0) hooks() unitSetMaxHPHooks4EE7C0[
	*unitSetMaxHPTestObject4EE7C0,
	*unitSetMaxHPTestHealth4EE7C0,
] {
	return unitSetMaxHPHooks4EE7C0[
		*unitSetMaxHPTestObject4EE7C0,
		*unitSetMaxHPTestHealth4EE7C0,
	]{
		loadUnitArg: func() *unitSetMaxHPTestObject4EE7C0 {
			unit := w.unit
			w.record("arg:" + unitSetMaxHPObjectName4EE7C0(unit))
			if w.afterUnit != nil {
				w.afterUnit(w, unit)
			}
			return unit
		},
		loadHealth: func(unit *unitSetMaxHPTestObject4EE7C0) *unitSetMaxHPTestHealth4EE7C0 {
			health := unit.health
			w.record("health:" + unitSetMaxHPObjectName4EE7C0(unit) + "=" + unitSetMaxHPHealthName4EE7C0(health))
			if w.afterHealth != nil {
				w.afterHealth(w, health)
			}
			return health
		},
		loadMaximumArg: func() uint16 {
			maximum := w.maximumArg
			w.record(fmt.Sprintf("maximum:%#04x", maximum))
			if w.afterMaximum != nil {
				w.afterMaximum(w, maximum)
			}
			return maximum
		},
		storeMaximum: func(health *unitSetMaxHPTestHealth4EE7C0, maximum uint16) {
			w.record(fmt.Sprintf("store:%s=%#04x", unitSetMaxHPHealthName4EE7C0(health), maximum))
			health.maximum = maximum
			if w.afterStore != nil {
				w.afterStore(w, health, maximum)
			}
		},
	}
}

func newUnitSetMaxHPTestWorld4EE7C0() *unitSetMaxHPTestWorld4EE7C0 {
	return &unitSetMaxHPTestWorld4EE7C0{
		unit: &unitSetMaxHPTestObject4EE7C0{
			name:   "unit",
			health: &unitSetMaxHPTestHealth4EE7C0{name: "health", maximum: 0x1111},
		},
		maximumArg: 0x4321,
	}
}

func TestUnitSetMaxHP4EE7C0EntryGates(t *testing.T) {
	tests := []struct {
		name string
		edit func(*unitSetMaxHPTestWorld4EE7C0)
		want []string
	}{
		{
			name: "nil unit",
			edit: func(w *unitSetMaxHPTestWorld4EE7C0) { w.unit = nil },
			want: []string{"arg:nil"},
		},
		{
			name: "nil health",
			edit: func(w *unitSetMaxHPTestWorld4EE7C0) { w.unit.health = nil },
			want: []string{"arg:unit", "health:unit=nil"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newUnitSetMaxHPTestWorld4EE7C0()
			test.edit(w)
			if got := unitSetMaxHP4EE7C0(w.hooks()); got != nil {
				t.Fatalf("result = %p, want nil", got)
			}
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %q, want %q", w.events, test.want)
			}
		})
	}
}

func TestUnitSetMaxHP4EE7C0PreservesEveryMaximumWord(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", maximum), func(t *testing.T) {
			w := newUnitSetMaxHPTestWorld4EE7C0()
			w.maximumArg = maximum
			health := w.unit.health
			if got := unitSetMaxHP4EE7C0(w.hooks()); got != health {
				t.Fatalf("result = %p, want cached health %p", got, health)
			}
			if health.maximum != maximum {
				t.Fatalf("maximum = %#04x, want %#04x", health.maximum, maximum)
			}
		})
	}
}

func TestUnitSetMaxHP4EE7C0CachesObjectHealthAndMaximum(t *testing.T) {
	w := newUnitSetMaxHPTestWorld4EE7C0()
	originalUnit := w.unit
	originalHealth := originalUnit.health
	replacementHealth := &unitSetMaxHPTestHealth4EE7C0{name: "replacement-health", maximum: 0xabcd}
	replacementUnit := &unitSetMaxHPTestObject4EE7C0{name: "replacement-unit", health: replacementHealth}
	w.afterUnit = func(w *unitSetMaxHPTestWorld4EE7C0, _ *unitSetMaxHPTestObject4EE7C0) {
		w.unit = replacementUnit
	}
	w.afterHealth = func(_ *unitSetMaxHPTestWorld4EE7C0, _ *unitSetMaxHPTestHealth4EE7C0) {
		originalUnit.health = replacementHealth
	}
	w.afterMaximum = func(w *unitSetMaxHPTestWorld4EE7C0, _ uint16) {
		w.maximumArg = 0x9876
	}

	if got := unitSetMaxHP4EE7C0(w.hooks()); got != originalHealth {
		t.Fatalf("result = %p, want cached health %p", got, originalHealth)
	}
	if originalHealth.maximum != 0x4321 {
		t.Fatalf("cached health maximum = %#04x, want 0x4321", originalHealth.maximum)
	}
	if replacementHealth.maximum != 0xabcd || w.maximumArg != 0x9876 {
		t.Fatalf("replacement state changed: health=%#04x argument=%#04x", replacementHealth.maximum, w.maximumArg)
	}
	want := []string{"arg:unit", "health:unit=health", "maximum:0x4321", "store:health=0x4321"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestUnitSetMaxHP4EE7C0AllFaultPrefixes(t *testing.T) {
	want := []string{"arg:unit", "health:unit=health", "maximum:0x4321", "store:health=0x4321"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newUnitSetMaxHPTestWorld4EE7C0()
			health := w.unit.health
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
				if health.maximum != 0x1111 {
					t.Fatalf("fault prefix stored maximum = %#04x", health.maximum)
				}
			}()
			unitSetMaxHP4EE7C0(w.hooks())
		})
	}
}
